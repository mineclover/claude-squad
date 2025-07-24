package session

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"claude-squad/log"
	"claude-squad/session/git"
)

// WorktreeTaskManager manages the execution of worktree-based tasks
type WorktreeTaskManager struct {
	storage         *Storage
	mainTasks       map[string]*MainTask // mainTaskID -> MainTask
	activeSubTasks  map[string]*SubTask  // subTaskID -> SubTask
	instances       map[string]*Instance // instanceID -> Instance
	gitWorktrees    map[string]*git.GitWorktree // mainTaskID -> GitWorktree
	webhookQueue    *WebhookQueue
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	stopCh          chan struct{}
	doneCh          chan struct{}
}

// SubTaskCompletion represents a completed subtask notification
type SubTaskCompletion struct {
	SubTaskID   string
	MainTaskID  string
	Success     bool
	Output      string
	ErrorMessage string
}

// NewWorktreeTaskManager creates a new WorktreeTaskManager
func NewWorktreeTaskManager(storage *Storage) *WorktreeTaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	webhookClient := NewWebhookClient()
	webhookQueue := NewWebhookQueue(webhookClient, 3, 100) // 3 workers, queue size 100
	
	return &WorktreeTaskManager{
		storage:        storage,
		mainTasks:      make(map[string]*MainTask),
		activeSubTasks: make(map[string]*SubTask),
		instances:      make(map[string]*Instance),
		gitWorktrees:   make(map[string]*git.GitWorktree),
		webhookQueue:   webhookQueue,
		ctx:            ctx,
		cancel:         cancel,
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}
}

// Start initializes the WorktreeTaskManager
func (wtm *WorktreeTaskManager) Start() error {
	wtm.webhookQueue.Start()
	
	// Start the main processing loop
	go wtm.processLoop()
	
	log.InfoLog.Printf("WorktreeTaskManager started")
	return nil
}

// Stop gracefully stops the WorktreeTaskManager
func (wtm *WorktreeTaskManager) Stop() error {
	log.InfoLog.Printf("Stopping WorktreeTaskManager...")
	
	wtm.cancel()
	close(wtm.stopCh)
	
	// Stop webhook queue
	wtm.webhookQueue.Stop()
	
	// Cleanup all active instances
	wtm.mu.Lock()
	for _, instance := range wtm.instances {
		if err := instance.Kill(); err != nil {
			log.ErrorLog.Printf("Failed to kill instance %s: %v", instance.Title, err)
		}
	}
	
	// Cleanup all worktrees
	for _, worktree := range wtm.gitWorktrees {
		if err := worktree.Cleanup(); err != nil {
			log.ErrorLog.Printf("Failed to cleanup worktree: %v", err)
		}
	}
	wtm.mu.Unlock()
	
	<-wtm.doneCh
	log.InfoLog.Printf("WorktreeTaskManager stopped")
	return nil
}

// ExecuteMainTask executes a main task with all its subtasks
func (wtm *WorktreeTaskManager) ExecuteMainTask(mainTask *MainTask) error {
	// Validate the main task
	if err := ValidateMainTask(mainTask); err != nil {
		return fmt.Errorf("main task validation failed: %w", err)
	}
	
	wtm.mu.Lock()
	wtm.mainTasks[mainTask.ID] = mainTask
	wtm.mu.Unlock()
	
	log.InfoLog.Printf("Starting execution of MainTask: %s (%s)", mainTask.Title, mainTask.ID)
	
	// Create worktree for the main task
	if err := wtm.setupWorktree(mainTask); err != nil {
		return fmt.Errorf("failed to setup worktree: %w", err)
	}
	
	// Execute subtasks sequentially
	go wtm.executeMainTaskAsync(mainTask)
	
	return nil
}

// setupWorktree creates and sets up a git worktree for the main task
func (wtm *WorktreeTaskManager) setupWorktree(mainTask *MainTask) error {
	// Create git worktree
	worktree, branchName, err := git.NewGitWorktree(mainTask.RepoPath, mainTask.ID)
	if err != nil {
		return fmt.Errorf("failed to create git worktree: %w", err)
	}
	
	// Setup the worktree
	if err := worktree.Setup(); err != nil {
		return fmt.Errorf("failed to setup worktree: %w", err)
	}
	
	// Update main task with worktree info
	mainTask.WorktreePath = worktree.GetWorktreePath()
	mainTask.BranchName = branchName
	
	wtm.mu.Lock()
	wtm.gitWorktrees[mainTask.ID] = worktree
	wtm.mu.Unlock()
	
	log.InfoLog.Printf("Created worktree for MainTask %s: %s (branch: %s)", 
		mainTask.ID, mainTask.WorktreePath, mainTask.BranchName)
	
	return nil
}

// executeMainTaskAsync executes the main task asynchronously
func (wtm *WorktreeTaskManager) executeMainTaskAsync(mainTask *MainTask) {
	mainTask.Status = TaskRunning
	
	for i := range mainTask.SubTasks {
		subTask := &mainTask.SubTasks[i]
		
		// Execute subtask
		if err := wtm.executeSubTask(mainTask, subTask); err != nil {
			log.ErrorLog.Printf("Failed to execute SubTask %s: %v", subTask.ID, err)
			mainTask.UpdateSubTaskStatus(subTask.ID, TaskFailed, "", err.Error())
			
			// Send failure webhook
			payload := CreateSubTaskCompletedPayload(mainTask, subTask)
			wtm.webhookQueue.Enqueue(wtm.ctx, mainTask.WebhookURL, payload)
			
			// Mark main task as failed if any subtask fails
			mainTask.Status = TaskFailed
			mainTask.ErrorMessage = fmt.Sprintf("SubTask %s failed: %s", subTask.ID, err.Error())
			break
		}
	}
	
	// Send main task completion webhook
	payload := CreateMainTaskCompletedPayload(mainTask)
	wtm.webhookQueue.Enqueue(wtm.ctx, mainTask.WebhookURL, payload)
	
	// Cleanup worktree after completion
	go wtm.cleanupMainTask(mainTask.ID)
}

// executeSubTask executes a single subtask
func (wtm *WorktreeTaskManager) executeSubTask(mainTask *MainTask, subTask *SubTask) error {
	log.InfoLog.Printf("Executing SubTask: %s (%s)", subTask.Title, subTask.ID)
	
	// Mark subtask as running
	mainTask.UpdateSubTaskStatus(subTask.ID, TaskRunning, "", "")
	
	wtm.mu.Lock()
	wtm.activeSubTasks[subTask.ID] = subTask
	wtm.mu.Unlock()
	
	// Send subtask started webhook
	payload := CreateSubTaskStartedPayload(mainTask, subTask)
	wtm.webhookQueue.Enqueue(wtm.ctx, mainTask.WebhookURL, payload)
	
	// Create instance for the subtask
	instanceTitle := fmt.Sprintf("%s-%s", mainTask.ID, subTask.ID)
	instance, err := NewInstance(InstanceOptions{
		Title:   instanceTitle,
		Path:    mainTask.WorktreePath, // Use the worktree path
		Program: subTask.Program,
		AutoYes: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}
	
	// Start the instance
	if err := instance.Start(true); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	
	wtm.mu.Lock()
	wtm.instances[subTask.ID] = instance
	wtm.mu.Unlock()
	
	// Send the prompt
	if err := instance.SendPrompt(subTask.Prompt); err != nil {
		wtm.cleanupSubTask(subTask.ID)
		return fmt.Errorf("failed to send prompt: %w", err)
	}
	
	// Wait for completion with timeout
	if err := wtm.waitForSubTaskCompletion(mainTask, subTask, instance); err != nil {
		wtm.cleanupSubTask(subTask.ID)
		return err
	}
	
	// Cleanup the instance
	wtm.cleanupSubTask(subTask.ID)
	
	log.InfoLog.Printf("SubTask completed successfully: %s", subTask.ID)
	return nil
}

// waitForSubTaskCompletion waits for a subtask to complete with timeout
func (wtm *WorktreeTaskManager) waitForSubTaskCompletion(mainTask *MainTask, subTask *SubTask, instance *Instance) error {
	timeout := time.After(subTask.Timeout)
	checkInterval := time.NewTicker(5 * time.Second)
	defer checkInterval.Stop()
	
	for {
		select {
		case <-wtm.ctx.Done():
			return wtm.ctx.Err()
			
		case <-timeout:
			mainTask.UpdateSubTaskStatus(subTask.ID, TaskTimedOut, "", "")
			payload := CreateSubTaskCompletedPayload(mainTask, subTask)
			wtm.webhookQueue.Enqueue(wtm.ctx, mainTask.WebhookURL, payload)
			return fmt.Errorf("subtask timed out after %s", subTask.Timeout)
			
		case <-checkInterval.C:
			// Check if the task is completed based on completion markers
			if wtm.checkSubTaskCompletion(subTask, instance) {
				output, err := instance.Preview()
				if err != nil {
					output = ""
				}
				
				mainTask.UpdateSubTaskStatus(subTask.ID, TaskCompleted, output, "")
				payload := CreateSubTaskCompletedPayload(mainTask, subTask)
				wtm.webhookQueue.Enqueue(wtm.ctx, mainTask.WebhookURL, payload)
				return nil
			}
		}
	}
}

// checkSubTaskCompletion checks if a subtask is completed based on completion markers
func (wtm *WorktreeTaskManager) checkSubTaskCompletion(subTask *SubTask, instance *Instance) bool {
	content, err := instance.Preview()
	if err != nil {
		return false
	}
	
	// If no completion markers specified, use a simple heuristic
	if len(subTask.CompletionMarkers) == 0 {
		// Check if the instance is not actively running (simple heuristic)
		updated, _ := instance.HasUpdated()
		return !updated // If not updated recently, assume completed
	}
	
	// Check for completion markers
	for _, marker := range subTask.CompletionMarkers {
		if strings.Contains(content, marker) {
			return true
		}
	}
	
	return false
}

// cleanupSubTask cleans up resources for a completed subtask
func (wtm *WorktreeTaskManager) cleanupSubTask(subTaskID string) {
	wtm.mu.Lock()
	defer wtm.mu.Unlock()
	
	// Remove from active subtasks
	delete(wtm.activeSubTasks, subTaskID)
	
	// Kill and remove instance
	if instance, exists := wtm.instances[subTaskID]; exists {
		if err := instance.Kill(); err != nil {
			log.ErrorLog.Printf("Failed to kill instance for subtask %s: %v", subTaskID, err)
		}
		delete(wtm.instances, subTaskID)
	}
}

// cleanupMainTask cleans up resources for a completed main task
func (wtm *WorktreeTaskManager) cleanupMainTask(mainTaskID string) {
	wtm.mu.Lock()
	defer wtm.mu.Unlock()
	
	// Cleanup worktree
	if worktree, exists := wtm.gitWorktrees[mainTaskID]; exists {
		if err := worktree.Cleanup(); err != nil {
			log.ErrorLog.Printf("Failed to cleanup worktree for main task %s: %v", mainTaskID, err)
		}
		delete(wtm.gitWorktrees, mainTaskID)
	}
	
	// Remove main task
	delete(wtm.mainTasks, mainTaskID)
	
	log.InfoLog.Printf("Cleaned up MainTask: %s", mainTaskID)
}

// GetMainTask returns a main task by ID
func (wtm *WorktreeTaskManager) GetMainTask(mainTaskID string) (*MainTask, bool) {
	wtm.mu.RLock()
	defer wtm.mu.RUnlock()
	
	mainTask, exists := wtm.mainTasks[mainTaskID]
	return mainTask, exists
}

// ListMainTasks returns all current main tasks
func (wtm *WorktreeTaskManager) ListMainTasks() []*MainTask {
	wtm.mu.RLock()
	defer wtm.mu.RUnlock()
	
	tasks := make([]*MainTask, 0, len(wtm.mainTasks))
	for _, task := range wtm.mainTasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetActiveSubTasks returns all currently active subtasks
func (wtm *WorktreeTaskManager) GetActiveSubTasks() []*SubTask {
	wtm.mu.RLock()
	defer wtm.mu.RUnlock()
	
	subTasks := make([]*SubTask, 0, len(wtm.activeSubTasks))
	for _, subTask := range wtm.activeSubTasks {
		subTasks = append(subTasks, subTask)
	}
	return subTasks
}

// processLoop is the main processing loop for the task manager
func (wtm *WorktreeTaskManager) processLoop() {
	defer close(wtm.doneCh)
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-wtm.stopCh:
			return
		case <-ticker.C:
			wtm.healthCheck()
		}
	}
}

// healthCheck performs periodic health checks on active tasks
func (wtm *WorktreeTaskManager) healthCheck() {
	wtm.mu.RLock()
	defer wtm.mu.RUnlock()
	
	// Check for stuck instances
	for subTaskID, instance := range wtm.instances {
		if !instance.Started() || !instance.TmuxAlive() {
			log.WarningLog.Printf("Instance for subtask %s appears to be stuck or dead", subTaskID)
			// Could implement recovery logic here
		}
	}
}