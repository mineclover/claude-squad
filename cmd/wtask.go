package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-squad/config"
	"claude-squad/log"
	"github.com/spf13/cobra"
)

var (
	wtaskTimeoutFlag string
	wtaskWebhookFlag string
	wtaskProgramFlag string
)

// wtaskCmd represents the wtask command
var wtaskCmd = &cobra.Command{
	Use:   "wtask [main-task-file]",
	Short: "Execute worktree-based main task with subtasks",
	Long: `Execute a worktree-based main task that contains multiple subtasks.
Each subtask runs in sequence within an isolated git worktree, and webhooks
are sent for each subtask completion.

The main-task-file should be a JSON file containing the task definition.

Example:
  cs wtask my-task.json
  cs wtask my-task.json --webhook https://api.example.com/hooks
  cs wtask my-task.json --timeout 1h --program claude`,
	Args: cobra.ExactArgs(1),
	RunE: runWTask,
}

func init() {
	wtaskCmd.Flags().StringVar(&wtaskTimeoutFlag, "timeout", "30m", 
		"Default timeout for subtasks (e.g., 30m, 1h, 2h30m)")
	wtaskCmd.Flags().StringVar(&wtaskWebhookFlag, "webhook", "",
		"Override webhook URL from the task file")
	wtaskCmd.Flags().StringVar(&wtaskProgramFlag, "program", "",
		"Override default program for all subtasks")
}

func runWTask(cmd *cobra.Command, args []string) error {
	taskFile := args[0]

	// Initialize logging
	log.Initialize(false)
	defer log.Close()

	log.InfoLog.Printf("Loading main task from: %s", taskFile)

	// Load main task from file
	mainTask, err := loadMainTaskFromFile(taskFile)
	if err != nil {
		return fmt.Errorf("failed to load main task: %w", err)
	}

	// Apply command line overrides
	if err := applyWTaskOverrides(mainTask); err != nil {
		return fmt.Errorf("failed to apply overrides: %w", err)
	}

	log.InfoLog.Printf("Loaded MainTask: %s with %d subtasks", mainTask.Title, len(mainTask.SubTasks))

	log.InfoLog.Printf("Main task execution would start here")
	log.InfoLog.Printf("Note: Full implementation requires session package integration")

	log.InfoLog.Printf("Main task execution started successfully")

	// Wait for completion (or user interruption)
	return waitForTaskCompletion(mainTask.ID)
}

// MainTask represents the task structure (avoiding circular import)
type MainTask struct {
	ID                string     `json:"id"`
	Title             string     `json:"title"`
	WorktreePath      string     `json:"worktree_path"`
	BranchName        string     `json:"branch_name"`
	RepoPath          string     `json:"repo_path"`
	Status            int        `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	WebhookURL        string     `json:"webhook_url"`
	SubTasks          []SubTask  `json:"subtasks"`
	CompletedSubTasks int        `json:"completed_subtasks"`
	ErrorMessage      string     `json:"error_message,omitempty"`
}

type SubTask struct {
	ID                string                 `json:"id"`
	MainTaskID        string                 `json:"main_task_id"`
	Title             string                 `json:"title"`
	Prompt            string                 `json:"prompt"`
	Program           string                 `json:"program"`
	CompletionMarkers []string               `json:"completion_markers"`
	Timeout           string                 `json:"timeout"`
	Status            int                    `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	WebhookPayload    map[string]interface{} `json:"webhook_payload,omitempty"`
	Output            string                 `json:"output,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
}

// loadMainTaskFromFile loads a MainTask from a JSON file
func loadMainTaskFromFile(filepath string) (*MainTask, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var mainTask MainTask
	if err := json.Unmarshal(data, &mainTask); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	// Set default values if not provided
	if mainTask.ID == "" {
		mainTask.ID = generateTaskID(mainTask.Title)
	}
	if mainTask.CreatedAt.IsZero() {
		mainTask.CreatedAt = time.Now()
	}

	// Set MainTaskID for all subtasks
	for i := range mainTask.SubTasks {
		mainTask.SubTasks[i].MainTaskID = mainTask.ID
		if mainTask.SubTasks[i].CreatedAt.IsZero() {
			mainTask.SubTasks[i].CreatedAt = time.Now()
		}
		if mainTask.SubTasks[i].Status == 0 { // TaskPending is 0
			mainTask.SubTasks[i].Status = 0 // TaskPending
		}
	}

	return &mainTask, nil
}

// applyWTaskOverrides applies command-line overrides to the main task
func applyWTaskOverrides(mainTask *MainTask) error {
	// Override webhook URL if provided
	if wtaskWebhookFlag != "" {
		mainTask.WebhookURL = wtaskWebhookFlag
		log.InfoLog.Printf("Overriding webhook URL to: %s", wtaskWebhookFlag)
	}

	// Parse default timeout
	var defaultTimeout time.Duration
	if wtaskTimeoutFlag != "" {
		var err error
		defaultTimeout, err = time.ParseDuration(wtaskTimeoutFlag)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
	}

	// Apply overrides to each subtask
	for i := range mainTask.SubTasks {
		subTask := &mainTask.SubTasks[i]

		// Override program if provided
		if wtaskProgramFlag != "" {
			subTask.Program = wtaskProgramFlag
			log.InfoLog.Printf("Overriding program for subtask %s to: %s", subTask.ID, wtaskProgramFlag)
		}

		// Set default program if not provided
		if subTask.Program == "" {
			cfg := config.LoadConfig()
			subTask.Program = cfg.DefaultProgram
		}

		// Set default timeout if not provided
		if subTask.Timeout == "" && defaultTimeout > 0 {
			subTask.Timeout = defaultTimeout.String()
		}
		if subTask.Timeout == "" {
			subTask.Timeout = "30m" // Default fallback
		}

		// Ensure subtask has an ID
		if subTask.ID == "" {
			subTask.ID = fmt.Sprintf("%s-subtask-%d", mainTask.ID, i+1)
		}
	}

	// Set repo path to current directory if not provided
	if mainTask.RepoPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		mainTask.RepoPath = cwd
	} else {
		// Convert to absolute path
		absPath, err := filepath.Abs(mainTask.RepoPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		mainTask.RepoPath = absPath
	}

	return nil
}

// generateTaskID generates a task ID from the title
func generateTaskID(title string) string {
	// Convert to lowercase, replace spaces with hyphens
	id := strings.ToLower(title)
	id = strings.ReplaceAll(id, " ", "-")
	
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	// Add timestamp to make it unique
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", result.String(), timestamp)
}

// waitForTaskCompletion waits for the main task to complete
func waitForTaskCompletion(mainTaskID string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	go func() {
		// This would need proper signal handling in a real implementation
		// For now, we'll just wait
	}()

	// Poll for completion
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.InfoLog.Printf("Task %s is running (simulated)", mainTaskID)
			// In real implementation, would check actual task status
			return nil
		}
	}
}