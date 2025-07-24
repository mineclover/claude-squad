# Automated Task Execution Specification

## Overview
자동 태스크 실행 시스템은 미리 정의된 작업을 세션에서 자동으로 생성, 실행, 완료 후 정리하는 기능입니다.

## Implementation Feasibility Analysis

### ✅ 가능한 구현 요소들

#### 1. 세션 자동 생성
**기존 코드**: `session/instance.go:155` - `NewInstance()`
```go
// 새로운 task 전용 instance 생성
func NewTaskInstance(opts TaskInstanceOptions) (*Instance, error) {
    instance, err := NewInstance(InstanceOptions{
        Title:   opts.TaskID,
        Path:    opts.WorkspacePath,
        Program: opts.Program,
        AutoYes: true, // 자동 승인 모드
    })
    return instance, err
}
```

#### 2. 자동 프롬프트 전송
**기존 코드**: `session/instance.go:514` - `SendPrompt()`
```go
// 태스크 프롬프트 자동 전송
func (i *Instance) SendTaskPrompt(taskPrompt string) error {
    return i.SendPrompt(taskPrompt)
}
```

#### 3. AutoYes 모드 활용
**기존 코드**: `session/instance.go:50`, `daemon/daemon.go:33`
- 기존 AutoYes 시스템 활용 가능
- daemon에서 자동으로 Enter 키 입력 처리

#### 4. 진행 상태 모니터링
**기존 코드**: `session/instance.go:304` - `HasUpdated()`
```go
// 태스크 완료 감지
func (i *Instance) IsTaskCompleted(completionMarkers []string) bool {
    content, err := i.Preview()
    if err != nil {
        return false
    }
    // completionMarkers 확인하여 완료 판단
    for _, marker := range completionMarkers {
        if strings.Contains(content, marker) {
            return true
        }
    }
    return false
}
```

#### 5. 자동 세션 정리
**기존 코드**: `session/instance.go:255` - `Kill()`, `session/storage.go:96` - `DeleteInstance()`
```go
// 태스크 완료 후 정리
func CleanupTaskInstance(storage *Storage, instanceTitle string) error {
    if err := storage.DeleteInstance(instanceTitle); err != nil {
        return err
    }
    return nil
}
```

## 제안 구조

### 1. Task 구조체 정의
```go
type Task struct {
    ID               string            `json:"id"`
    Title            string            `json:"title"`
    Prompt           string            `json:"prompt"`
    WorkspacePath    string            `json:"workspace_path"`
    Program          string            `json:"program"`
    CompletionMarkers []string         `json:"completion_markers"`
    Timeout          time.Duration     `json:"timeout"`
    AutoCleanup      bool             `json:"auto_cleanup"`
    Status           TaskStatus       `json:"status"`
    CreatedAt        time.Time        `json:"created_at"`
    CompletedAt      *time.Time       `json:"completed_at,omitempty"`
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

### 2. TaskManager 컴포넌트
```go
type TaskManager struct {
    storage     *session.Storage
    tasks       map[string]*Task
    instances   map[string]*session.Instance
    completedCh chan string
}

func (tm *TaskManager) ExecuteTask(task *Task) error {
    // 1. 세션 생성
    // 2. 프롬프트 전송  
    // 3. 모니터링 시작
    // 4. 완료 시 정리
}
```

### 3. CLI 인터페이스 확장
```go
var taskCmd = &cobra.Command{
    Use:   "task [task-file]",
    Short: "Execute predefined tasks automatically",
    RunE: func(cmd *cobra.Command, args []string) error {
        // task file 로드 및 실행
    },
}
```

## 구현 단계

### Phase 1: Core Task System
1. `session/task.go` - Task 구조체 및 기본 함수
2. `session/task_manager.go` - TaskManager 구현
3. Task 저장/로드 기능

### Phase 2: Integration
1. 기존 daemon 시스템과 통합
2. CLI 명령어 추가
3. 설정 파일 지원

### Phase 3: Advanced Features
1. Task 템플릿 시스템
2. 병렬 Task 실행
3. Task 실행 결과 저장

## 기술적 제약사항 및 해결방안

### 1. 완료 감지의 어려움
**문제**: AI 응답 완료를 정확히 감지하기 어려움
**해결**: 
- 특정 완료 마커 문자열 감지
- 일정 시간 동안 출력 없음으로 판단
- 프롬프트 패턴 분석

### 2. 오류 처리
**문제**: 실행 중 오류 발생 시 처리
**해결**:
- 타임아웃 설정
- 재시도 메커니즘
- 실패 시 로그 저장

### 3. 리소스 관리
**문제**: 많은 Task 동시 실행 시 리소스 부족
**해결**:
- Task 큐 시스템 도입
- 동시 실행 제한
- 우선순위 기반 스케줄링

## 필요한 추가 정보

**세션 삭제 관련**: 다음 코드 분석이 필요합니다:
- `ui/list.go` - List.Kill() 메서드
- 세션 완료 후 UI에서 제거하는 로직
- Storage에서 완전 삭제하는 추가 메서드 필요 여부

## 결론

제안된 자동 Task 실행 시스템은 **기술적으로 구현 가능**합니다. 기존 코드베이스의 다음 요소들을 활용할 수 있습니다:

1. ✅ 세션 생성/관리 시스템
2. ✅ AutoYes 자동 응답 시스템  
3. ✅ tmux 세션 격리
4. ✅ Git worktree 시스템
5. ✅ 데몬 모니터링 시스템
6. ✅ 저장/복원 시스템

주요 구현 포인트는 Task 정의, 완료 감지, 자동 정리 로직이며, 기존 아키텍처와 자연스럽게 통합 가능합니다.