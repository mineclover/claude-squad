package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"claude-squad/log"
)

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	EventType     string                 `json:"event_type"` // "subtask_started", "subtask_completed", "subtask_failed", "maintask_completed", "maintask_failed"
	MainTaskID    string                 `json:"main_task_id"`
	SubTaskID     string                 `json:"subtask_id,omitempty"`
	Status        string                 `json:"status"` // "success", "failed", "timeout"
	WorktreePath  string                 `json:"worktree_path,omitempty"`
	BranchName    string                 `json:"branch_name,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Output        string                 `json:"output,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Progress      float64                `json:"progress,omitempty"` // Completion percentage for main task
	CustomData    map[string]interface{} `json:"custom_data,omitempty"`
}

// WebhookClient handles webhook delivery with retry logic
type WebhookClient struct {
	httpClient *http.Client
	retryCount int
	retryDelay time.Duration
}

// NewWebhookClient creates a new webhook client with default settings
func NewWebhookClient() *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retryCount: 3,
		retryDelay: 1 * time.Second,
	}
}

// NewWebhookClientWithConfig creates a webhook client with custom configuration
func NewWebhookClientWithConfig(timeout time.Duration, retryCount int, retryDelay time.Duration) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
}

// SendWebhook sends a webhook payload to the specified URL with retry logic
func (wc *WebhookClient) SendWebhook(ctx context.Context, webhookURL string, payload WebhookPayload) error {
	if webhookURL == "" {
		log.InfoLog.Printf("No webhook URL configured, skipping webhook for %s", payload.EventType)
		return nil
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= wc.retryCount; attempt++ {
		if attempt > 0 {
			log.InfoLog.Printf("Retrying webhook delivery (attempt %d/%d) for %s", attempt+1, wc.retryCount+1, payload.EventType)
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wc.retryDelay * time.Duration(attempt)):
			}
		}

		lastErr = wc.sendWebhookAttempt(ctx, webhookURL, jsonData)
		if lastErr == nil {
			log.InfoLog.Printf("Successfully sent webhook for %s (attempt %d)", payload.EventType, attempt+1)
			return nil
		}

		log.WarningLog.Printf("Webhook delivery attempt %d failed for %s: %v", attempt+1, payload.EventType, lastErr)
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", wc.retryCount+1, lastErr)
}

// sendWebhookAttempt makes a single webhook delivery attempt
func (wc *WebhookClient) sendWebhookAttempt(ctx context.Context, webhookURL string, jsonData []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "claude-squad-webhook/1.0")

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for logging
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateSubTaskStartedPayload creates a webhook payload for subtask started event
func CreateSubTaskStartedPayload(mainTask *MainTask, subTask *SubTask) WebhookPayload {
	payload := WebhookPayload{
		EventType:    "subtask_started",
		MainTaskID:   mainTask.ID,
		SubTaskID:    subTask.ID,
		Status:       "running",
		WorktreePath: mainTask.WorktreePath,
		BranchName:   mainTask.BranchName,
		Timestamp:    time.Now(),
		Progress:     mainTask.GetProgress(),
		CustomData:   subTask.WebhookPayload,
	}
	return payload
}

// CreateSubTaskCompletedPayload creates a webhook payload for subtask completed event
func CreateSubTaskCompletedPayload(mainTask *MainTask, subTask *SubTask) WebhookPayload {
	status := "success"
	if subTask.IsFailed() {
		if subTask.Status == TaskTimedOut {
			status = "timeout"
		} else {
			status = "failed"
		}
	}

	payload := WebhookPayload{
		EventType:    "subtask_completed",
		MainTaskID:   mainTask.ID,
		SubTaskID:    subTask.ID,
		Status:       status,
		WorktreePath: mainTask.WorktreePath,
		BranchName:   mainTask.BranchName,
		Timestamp:    time.Now(),
		Output:       subTask.Output,
		ErrorMessage: subTask.ErrorMessage,
		Progress:     mainTask.GetProgress(),
		CustomData:   subTask.WebhookPayload,
	}
	return payload
}

// CreateMainTaskCompletedPayload creates a webhook payload for main task completed event
func CreateMainTaskCompletedPayload(mainTask *MainTask) WebhookPayload {
	status := "success"
	eventType := "maintask_completed"
	
	if mainTask.IsFailed() {
		status = "failed"
		eventType = "maintask_failed"
	}

	payload := WebhookPayload{
		EventType:    eventType,
		MainTaskID:   mainTask.ID,
		Status:       status,
		WorktreePath: mainTask.WorktreePath,
		BranchName:   mainTask.BranchName,
		Timestamp:    time.Now(),
		ErrorMessage: mainTask.ErrorMessage,
		Progress:     mainTask.GetProgress(),
		CustomData:   make(map[string]interface{}),
	}

	// Add summary information
	payload.CustomData["total_subtasks"] = len(mainTask.SubTasks)
	payload.CustomData["completed_subtasks"] = mainTask.CompletedSubTasks
	payload.CustomData["title"] = mainTask.Title
	
	if mainTask.CompletedAt != nil {
		payload.CustomData["duration"] = mainTask.CompletedAt.Sub(mainTask.CreatedAt).String()
	}

	return payload
}

// WebhookQueue manages queued webhook deliveries for reliability
type WebhookQueue struct {
	client   *WebhookClient
	queue    chan WebhookDelivery
	workers  int
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// WebhookDelivery represents a queued webhook delivery
type WebhookDelivery struct {
	URL     string
	Payload WebhookPayload
	Context context.Context
}

// NewWebhookQueue creates a new webhook queue with specified number of workers
func NewWebhookQueue(client *WebhookClient, workers int, queueSize int) *WebhookQueue {
	return &WebhookQueue{
		client:  client,
		queue:   make(chan WebhookDelivery, queueSize),
		workers: workers,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

// Start begins processing webhook deliveries
func (wq *WebhookQueue) Start() {
	log.InfoLog.Printf("Starting webhook queue with %d workers", wq.workers)
	
	for i := 0; i < wq.workers; i++ {
		go wq.worker(i)
	}
}

// Stop gracefully stops the webhook queue
func (wq *WebhookQueue) Stop() {
	log.InfoLog.Printf("Stopping webhook queue...")
	close(wq.stopCh)
	<-wq.doneCh
}

// Enqueue adds a webhook delivery to the queue
func (wq *WebhookQueue) Enqueue(ctx context.Context, url string, payload WebhookPayload) error {
	select {
	case wq.queue <- WebhookDelivery{URL: url, Payload: payload, Context: ctx}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("webhook queue is full")
	}
}

// worker processes webhook deliveries from the queue
func (wq *WebhookQueue) worker(id int) {
	defer func() {
		if id == 0 { // Only first worker signals completion
			close(wq.doneCh)
		}
	}()

	for {
		select {
		case <-wq.stopCh:
			return
		case delivery := <-wq.queue:
			if err := wq.client.SendWebhook(delivery.Context, delivery.URL, delivery.Payload); err != nil {
				log.ErrorLog.Printf("Worker %d failed to deliver webhook: %v", id, err)
			}
		}
	}
}