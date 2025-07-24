package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"claude-squad/config"
)

// WorktreeTaskStorage handles saving and loading worktree tasks
type WorktreeTaskStorage struct {
	state config.InstanceStorage
}

// WorktreeTaskData represents serializable data for a MainTask
type WorktreeTaskData struct {
	MainTask     MainTask  `json:"main_task"`
	SavedAt      string    `json:"saved_at"`
	Version      string    `json:"version"`
}

const WorktreeTaskStorageVersion = "1.0"

// NewWorktreeTaskStorage creates a new worktree task storage instance
func NewWorktreeTaskStorage(state config.InstanceStorage) *WorktreeTaskStorage {
	return &WorktreeTaskStorage{
		state: state,
	}
}

// SaveMainTask saves a MainTask to persistent storage
func (wts *WorktreeTaskStorage) SaveMainTask(mainTask *MainTask) error {
	taskData := WorktreeTaskData{
		MainTask:     *mainTask,
		SavedAt:      mainTask.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Version:      WorktreeTaskStorageVersion,
	}

	jsonData, err := json.MarshalIndent(taskData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal main task: %w", err)
	}

	// Use a task-specific storage key
	storageKey := fmt.Sprintf("wtask_%s", mainTask.ID)
	
	// For now, save as a file in the config directory
	// In a real implementation, this might use a database or the existing storage system
	return wts.saveTaskToFile(storageKey, jsonData)
}

// LoadMainTask loads a MainTask from persistent storage
func (wts *WorktreeTaskStorage) LoadMainTask(taskID string) (*MainTask, error) {
	storageKey := fmt.Sprintf("wtask_%s", taskID)
	
	jsonData, err := wts.loadTaskFromFile(storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load task data: %w", err)
	}

	var taskData WorktreeTaskData
	if err := json.Unmarshal(jsonData, &taskData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	return &taskData.MainTask, nil
}

// ListMainTasks returns a list of all saved MainTask IDs
func (wts *WorktreeTaskStorage) ListMainTasks() ([]string, error) {
	// This would be implemented to scan the storage for task files
	// For now, return an empty list
	return []string{}, nil
}

// DeleteMainTask removes a MainTask from persistent storage
func (wts *WorktreeTaskStorage) DeleteMainTask(taskID string) error {
	storageKey := fmt.Sprintf("wtask_%s", taskID)
	return wts.deleteTaskFile(storageKey)
}

// saveTaskToFile saves task data to a file in the config directory
func (wts *WorktreeTaskStorage) saveTaskToFile(storageKey string, jsonData []byte) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	taskFile := filepath.Join(configDir, fmt.Sprintf("%s.json", storageKey))
	
	// Write to file system
	// This is a simple implementation - in production, you might want to use
	// atomic writes, backup files, etc.
	if err := writeFile(taskFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// loadTaskFromFile loads task data from a file in the config directory
func (wts *WorktreeTaskStorage) loadTaskFromFile(storageKey string) ([]byte, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	taskFile := filepath.Join(configDir, fmt.Sprintf("%s.json", storageKey))
	
	data, err := readFile(taskFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	return data, nil
}

// deleteTaskFile removes a task file from the config directory
func (wts *WorktreeTaskStorage) deleteTaskFile(storageKey string) error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	taskFile := filepath.Join(configDir, fmt.Sprintf("%s.json", storageKey))
	
	if err := removeFile(taskFile); err != nil {
		return fmt.Errorf("failed to delete task file: %w", err)
	}

	return nil
}

// LoadMainTaskFromFile loads a MainTask from a specified JSON file
func LoadMainTaskFromFile(filepath string) (*MainTask, error) {
	data, err := readFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file: %w", err)
	}

	var mainTask MainTask
	if err := json.Unmarshal(data, &mainTask); err != nil {
		return nil, fmt.Errorf("failed to parse task file: %w", err)
	}

	return &mainTask, nil
}

// SaveMainTaskToFile saves a MainTask to a specified JSON file
func SaveMainTaskToFile(mainTask *MainTask, filepath string) error {
	jsonData, err := json.MarshalIndent(mainTask, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal main task: %w", err)
	}

	if err := writeFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// File system abstraction functions for easier testing
var (
	readFile   = readFileFunc
	writeFile  = writeFileFunc
	removeFile = removeFileFunc
)

func readFileFunc(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func writeFileFunc(filename string, data []byte, perm int) error {
	return os.WriteFile(filename, data, os.FileMode(perm))
}

func removeFileFunc(filename string) error {
	return os.Remove(filename)
}