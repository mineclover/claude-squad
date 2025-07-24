package session

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskTimedOut
)

func (ts TaskStatus) String() string {
	switch ts {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	case TaskTimedOut:
		return "timed_out"
	default:
		return "unknown"
	}
}

// SubTask represents an individual executable unit within a MainTask
type SubTask struct {
	ID                string                 `json:"id"`
	MainTaskID        string                 `json:"main_task_id"`
	Title             string                 `json:"title"`
	Prompt            string                 `json:"prompt"`
	Program           string                 `json:"program"`
	CompletionMarkers []string               `json:"completion_markers"`
	Timeout           time.Duration          `json:"timeout"`
	Status            TaskStatus             `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	WebhookPayload    map[string]interface{} `json:"webhook_payload,omitempty"`
	Output            string                 `json:"output,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
}

// MainTask represents a worktree-level task containing multiple SubTasks
type MainTask struct {
	ID                string     `json:"id"`
	Title             string     `json:"title"`
	WorktreePath      string     `json:"worktree_path"`
	BranchName        string     `json:"branch_name"`
	RepoPath          string     `json:"repo_path"`
	Status            TaskStatus `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	WebhookURL        string     `json:"webhook_url"`
	SubTasks          []SubTask  `json:"subtasks"`
	CompletedSubTasks int        `json:"completed_subtasks"`
	ErrorMessage      string     `json:"error_message,omitempty"`
}

// NewSubTask creates a new SubTask instance
func NewSubTask(id, mainTaskID, title, prompt, program string, completionMarkers []string, timeout time.Duration) *SubTask {
	return &SubTask{
		ID:                id,
		MainTaskID:        mainTaskID,
		Title:             title,
		Prompt:            prompt,
		Program:           program,
		CompletionMarkers: completionMarkers,
		Timeout:           timeout,
		Status:            TaskPending,
		CreatedAt:         time.Now(),
		WebhookPayload:    make(map[string]interface{}),
	}
}

// NewMainTask creates a new MainTask instance
func NewMainTask(id, title, repoPath, webhookURL string, subTasks []SubTask) *MainTask {
	// Set MainTaskID for all subtasks
	for i := range subTasks {
		subTasks[i].MainTaskID = id
	}

	return &MainTask{
		ID:                id,
		Title:             title,
		RepoPath:          repoPath,
		Status:            TaskPending,
		CreatedAt:         time.Now(),
		WebhookURL:        webhookURL,
		SubTasks:          subTasks,
		CompletedSubTasks: 0,
	}
}

// IsCompleted returns true if the subtask is completed successfully
func (st *SubTask) IsCompleted() bool {
	return st.Status == TaskCompleted
}

// IsFailed returns true if the subtask has failed
func (st *SubTask) IsFailed() bool {
	return st.Status == TaskFailed || st.Status == TaskTimedOut
}

// MarkCompleted marks the subtask as completed
func (st *SubTask) MarkCompleted(output string) {
	st.Status = TaskCompleted
	now := time.Now()
	st.CompletedAt = &now
	st.Output = output
}

// MarkFailed marks the subtask as failed
func (st *SubTask) MarkFailed(errorMsg string) {
	st.Status = TaskFailed
	now := time.Now()
	st.CompletedAt = &now
	st.ErrorMessage = errorMsg
}

// MarkTimedOut marks the subtask as timed out
func (st *SubTask) MarkTimedOut() {
	st.Status = TaskTimedOut
	now := time.Now()
	st.CompletedAt = &now
	st.ErrorMessage = "Task timed out"
}

// SetRunning marks the subtask as running
func (st *SubTask) SetRunning() {
	st.Status = TaskRunning
}

// IsCompleted returns true if all subtasks are completed
func (mt *MainTask) IsCompleted() bool {
	return mt.CompletedSubTasks == len(mt.SubTasks) && mt.Status == TaskCompleted
}

// IsFailed returns true if the main task has failed
func (mt *MainTask) IsFailed() bool {
	return mt.Status == TaskFailed
}

// GetProgress returns the completion progress as a percentage
func (mt *MainTask) GetProgress() float64 {
	if len(mt.SubTasks) == 0 {
		return 100.0
	}
	return float64(mt.CompletedSubTasks) / float64(len(mt.SubTasks)) * 100.0
}

// GetNextPendingSubTask returns the next pending subtask, or nil if none
func (mt *MainTask) GetNextPendingSubTask() *SubTask {
	for i := range mt.SubTasks {
		if mt.SubTasks[i].Status == TaskPending {
			return &mt.SubTasks[i]
		}
	}
	return nil
}

// UpdateSubTaskStatus updates a subtask's status and recalculates main task progress
func (mt *MainTask) UpdateSubTaskStatus(subTaskID string, status TaskStatus, output, errorMsg string) error {
	for i := range mt.SubTasks {
		if mt.SubTasks[i].ID == subTaskID {
			oldStatus := mt.SubTasks[i].Status
			
			switch status {
			case TaskCompleted:
				mt.SubTasks[i].MarkCompleted(output)
				if oldStatus != TaskCompleted {
					mt.CompletedSubTasks++
				}
			case TaskFailed:
				mt.SubTasks[i].MarkFailed(errorMsg)
			case TaskTimedOut:
				mt.SubTasks[i].MarkTimedOut()
			case TaskRunning:
				mt.SubTasks[i].SetRunning()
			}
			
			// Update main task status based on subtask progress
			mt.updateMainTaskStatus()
			return nil
		}
	}
	return fmt.Errorf("subtask with ID %s not found", subTaskID)
}

// updateMainTaskStatus updates the main task status based on subtask progress
func (mt *MainTask) updateMainTaskStatus() {
	totalSubTasks := len(mt.SubTasks)
	completedCount := 0
	failedCount := 0
	runningCount := 0

	for _, subTask := range mt.SubTasks {
		switch subTask.Status {
		case TaskCompleted:
			completedCount++
		case TaskFailed, TaskTimedOut:
			failedCount++
		case TaskRunning:
			runningCount++
		}
	}

	if completedCount == totalSubTasks {
		mt.Status = TaskCompleted
		now := time.Now()
		mt.CompletedAt = &now
	} else if failedCount > 0 {
		mt.Status = TaskFailed
		now := time.Now()
		mt.CompletedAt = &now
		mt.ErrorMessage = fmt.Sprintf("%d subtasks failed", failedCount)
	} else if runningCount > 0 {
		mt.Status = TaskRunning
	}
}

// ToJSON serializes the MainTask to JSON
func (mt *MainTask) ToJSON() ([]byte, error) {
	return json.MarshalIndent(mt, "", "  ")
}

// FromJSON deserializes a MainTask from JSON
func (mt *MainTask) FromJSON(data []byte) error {
	return json.Unmarshal(data, mt)
}


// ValidateMainTask validates that a MainTask is properly configured
func ValidateMainTask(mt *MainTask) error {
	if mt.ID == "" {
		return fmt.Errorf("main task ID cannot be empty")
	}
	if mt.Title == "" {
		return fmt.Errorf("main task title cannot be empty")
	}
	if mt.RepoPath == "" {
		return fmt.Errorf("repo path cannot be empty")
	}
	if len(mt.SubTasks) == 0 {
		return fmt.Errorf("main task must have at least one subtask")
	}
	
	// Validate each subtask
	for i, subTask := range mt.SubTasks {
		if err := ValidateSubTask(&subTask); err != nil {
			return fmt.Errorf("subtask %d validation failed: %w", i, err)
		}
	}
	
	return nil
}

// ValidateSubTask validates that a SubTask is properly configured
func ValidateSubTask(st *SubTask) error {
	if st.ID == "" {
		return fmt.Errorf("subtask ID cannot be empty")
	}
	if st.Title == "" {
		return fmt.Errorf("subtask title cannot be empty")
	}
	if st.Prompt == "" {
		return fmt.Errorf("subtask prompt cannot be empty")
	}
	if st.Program == "" {
		return fmt.Errorf("subtask program cannot be empty")
	}
	if st.Timeout <= 0 {
		return fmt.Errorf("subtask timeout must be positive")
	}
	
	return nil
}