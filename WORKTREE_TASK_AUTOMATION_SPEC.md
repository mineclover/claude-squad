# Worktree-Based Task Automation Specification

## Overview
Worktree 기반 자동 태스크 실행 시스템은 각 워크스페이스마다 하나의 메인 태스크를 관리하고, 메인 태스크 하위의 서브태스크들이 완료될 때마다 webhook을 통해 외부 시스템에 알림을 보내는 구조입니다.

## Implementation Feasibility Analysis

### ✅ 가능한 구현 요소들

#### 1. Worktree 기반 격리 작업공간
**기존 코드**: `session/git/worktree.go:44`, `session/git/worktree_ops.go:16`
- 각 MainTask는 독립적인 git worktree에서 실행
- 브랜치 충돌 없이 병렬 작업 가능
- 작업 완료 후 worktree 자동 정리 (`Cleanup()`, `Remove()`)

#### 2. 세션 자동 생성 및 관리
**기존 코드**: `session/instance.go:155` - `NewInstance()`
```go
// MainTask 전용 instance 생성 (worktree 내에서 실행)
func NewWorktreeTaskInstance(mainTask *MainTask, subTask *SubTask) (*Instance, error) {
    instance, err := NewInstance(InstanceOptions{
        Title:   fmt.Sprintf("%s-%s", mainTask.ID, subTask.ID),
        Path:    mainTask.WorktreePath, // worktree 경로 사용
        Program: subTask.Program,
        AutoYes: true,
    })
    return instance, err
}
```

#### 3. AutoYes 자동 실행 모드
**기존 코드**: `session/instance.go:50`, `daemon/daemon.go:33`
- 기존 AutoYes 시스템으로 무인 실행 지원
- daemon이 자동으로 Enter 키 입력 처리

#### 4. SubTask 진행 상태 모니터링
**기존 코드**: `session/instance.go:304` - `HasUpdated()`
```go
// SubTask 완료 감지
func (i *Instance) IsSubTaskCompleted(completionMarkers []string) bool {
    content, err := i.Preview()
    if err != nil {
        return false
    }
    for _, marker := range completionMarkers {
        if strings.Contains(content, marker) {
            return true
        }
    }
    return false
}
```

#### 5. 자동 정리 시스템
**기존 코드**: `session/instance.go:255` - `Kill()`, `session/storage.go:96` - `DeleteInstance()`
- SubTask 완료 후 세션 정리
- MainTask 완료 후 worktree 정리

## 제안 구조

### 1. Worktree-Task 구조체 정의

#### MainTask (워크트리 단위)
```go
type MainTask struct {
    ID               string            `json:"id"`
    Title            string            `json:"title"`
    WorktreePath     string            `json:"worktree_path"`
    BranchName       string            `json:"branch_name"`
    RepoPath         string            `json:"repo_path"`
    Status           TaskStatus        `json:"status"`
    CreatedAt        time.Time         `json:"created_at"`
    CompletedAt      *time.Time        `json:"completed_at,omitempty"`
    WebhookURL       string            `json:"webhook_url"`
    SubTasks         []SubTask         `json:"subtasks"`
    CompletedSubTasks int              `json:"completed_subtasks"`
}
```

#### SubTask (개별 실행 단위)
```go
type SubTask struct {
    ID               string            `json:"id"`
    MainTaskID       string            `json:"main_task_id"`
    Title            string            `json:"title"`
    Prompt           string            `json:"prompt"`
    Program          string            `json:"program"`
    CompletionMarkers []string         `json:"completion_markers"`
    Timeout          time.Duration     `json:"timeout"`
    Status           TaskStatus        `json:"status"`
    CreatedAt        time.Time         `json:"created_at"`
    CompletedAt      *time.Time        `json:"completed_at,omitempty"`
    WebhookPayload   map[string]interface{} `json:"webhook_payload,omitempty"`
}

type TaskStatus int
const (
    TaskPending TaskStatus = iota
    TaskRunning
    TaskCompleted
    TaskFailed
    TaskTimedOut
)
```

### 2. WorktreeTaskManager 컴포넌트
```go
type WorktreeTaskManager struct {
    storage         *session.Storage
    mainTasks       map[string]*MainTask    // worktree별 main task
    activeSubTasks  map[string]*SubTask     // 현재 실행 중인 subtask
    instances       map[string]*session.Instance
    webhookClient   *http.Client
    completedCh     chan SubTaskCompletion
}

type SubTaskCompletion struct {
    SubTaskID   string
    MainTaskID  string
    Success     bool
    Output      string
}

func (wtm *WorktreeTaskManager) ExecuteMainTask(mainTask *MainTask) error {
    // 1. Worktree 생성
    // 2. 각 SubTask 순차 실행
    // 3. SubTask 완료 시 webhook 발송
    // 4. 모든 SubTask 완료 시 정리
}

func (wtm *WorktreeTaskManager) ExecuteSubTask(subTask *SubTask) error {
    // 1. 해당 worktree에서 세션 생성
    // 2. 프롬프트 전송
    // 3. 완료 대기
    // 4. Webhook 발송
}
```

### 3. Webhook Integration
```go
type WebhookPayload struct {
    EventType     string                 `json:"event_type"` // "subtask_completed", "maintask_completed"
    MainTaskID    string                 `json:"main_task_id"`
    SubTaskID     string                 `json:"subtask_id,omitempty"`
    Status        string                 `json:"status"` // "success", "failed"
    WorktreePath  string                 `json:"worktree_path"`
    BranchName    string                 `json:"branch_name"`
    Timestamp     time.Time              `json:"timestamp"`
    Output        string                 `json:"output,omitempty"`
    CustomData    map[string]interface{} `json:"custom_data,omitempty"`
}

func (wtm *WorktreeTaskManager) sendWebhook(webhookURL string, payload WebhookPayload) error {
    jsonData, _ := json.Marshal(payload)
    resp, err := wtm.webhookClient.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
    // error handling...
    return nil
}
```

### 4. CLI 인터페이스 확장
```go
var worktreeTaskCmd = &cobra.Command{
    Use:   "wtask [main-task-file]",
    Short: "Execute worktree-based main task with subtasks",
    RunE: func(cmd *cobra.Command, args []string) error {
        // main task file 로드 및 실행
    },
}
```

## 구현 단계

### Phase 1: Worktree-Task Core System
1. `session/worktree_task.go` - MainTask, SubTask 구조체
2. `session/worktree_task_manager.go` - WorktreeTaskManager 구현
3. `session/webhook.go` - Webhook 발송 기능
4. MainTask/SubTask 저장/로드 기능

### Phase 2: Worktree Integration  
1. 기존 Git worktree 시스템과 연동
2. SubTask별 세션 생성/관리
3. CLI 명령어 추가 (`wtask`)
4. MainTask 설정 파일 지원

### Phase 3: Advanced Webhook Features
1. Webhook 재시도 메커니즘
2. 커스텀 webhook payload 지원
3. 실행 결과 저장 및 히스토리
4. 병렬 MainTask 실행 지원

## Worktree-Task 아키텍처의 특징

### 1. 워크스페이스 격리
- 각 MainTask는 독립적인 git worktree에서 실행
- 브랜치 충돌 없이 병렬 작업 가능
- 작업 완료 후 worktree 자동 정리

### 2. 단계별 진행 관리
- SubTask 단위로 세밀한 진행 관리
- 각 SubTask 완료마다 webhook 발송
- MainTask 진행률 실시간 추적

### 3. 외부 시스템 연동
- SubTask 완료 시마다 webhook으로 알림
- 커스텀 payload로 상세 정보 전송
- 외부 시스템에서 진행 상황 모니터링 가능

## 기술적 제약사항 및 해결방안

### 1. SubTask 완료 감지
**문제**: AI 응답 완료를 정확히 감지하기 어려움
**해결**: 
- SubTask별 완료 마커 문자열 정의
- 타임아웃 기반 완료 판단
- 출력 패턴 분석으로 완료 감지

### 2. Webhook 전송 신뢰성
**문제**: 네트워크 오류로 webhook 전송 실패
**해결**:
- 재시도 메커니즘 (exponential backoff)
- webhook 전송 큐 시스템
- 실패 시 로그 저장 및 수동 재전송

### 3. Worktree 리소스 관리
**문제**: 많은 MainTask 동시 실행 시 디스크/메모리 부족
**해결**:
- MainTask 동시 실행 제한
- 완료된 worktree 즉시 정리
- 디스크 사용량 모니터링

## MainTask 예시 구조

```json
{
  "id": "maintask-001",
  "title": "Feature Implementation",
  "repo_path": "/path/to/repo",
  "webhook_url": "https://api.example.com/webhooks/task-progress",
  "subtasks": [
    {
      "id": "subtask-001",
      "title": "Create API endpoint",
      "prompt": "Create a REST API endpoint for user management",
      "program": "claude",
      "completion_markers": ["API endpoint created successfully", "Tests passing"],
      "timeout": "30m",
      "webhook_payload": {
        "priority": "high",
        "component": "backend"
      }
    },
    {
      "id": "subtask-002", 
      "title": "Update frontend",
      "prompt": "Update frontend to use the new API",
      "program": "claude",
      "completion_markers": ["Frontend updated", "UI tests passing"],
      "timeout": "20m"
    }
  ]
}
```

## 필요한 추가 정보

**세션 삭제 관련**: 다음 코드 분석이 필요합니다:
- `ui/list.go` - List.Kill() 메서드
- 세션 완료 후 UI에서 제거하는 로직
- Storage에서 완전 삭제하는 추가 메서드 필요 여부

## 결론

제안된 Worktree 기반 Task 자동화 시스템은 **기술적으로 구현 가능**하며, 기존 claude-squad 아키텍처와 완벽하게 호환됩니다:

### ✅ 활용 가능한 기존 시스템들:
1. **Git Worktree 시스템** - 독립적인 작업공간 생성
2. **세션 생성/관리** - SubTask별 세션 실행
3. **AutoYes 자동 응답** - 무인 실행 지원
4. **tmux 세션 격리** - 안정적인 프로세스 관리
5. **저장/복원 시스템** - MainTask 상태 persistence

### 🆕 새로 구현할 요소들:
1. **MainTask/SubTask 구조** - 계층적 작업 관리
2. **Webhook 시스템** - HTTP 클라이언트 및 payload 관리
3. **워크플로우 엔진** - SubTask 순차 실행 및 상태 관리
4. **CLI 확장** - `wtask` 명령어 추가

이 시스템을 통해 **워크스페이스 단위의 대규모 작업**을 **세밀한 단계별 관리**와 **실시간 외부 알림**으로 효율적으로 자동화할 수 있습니다.