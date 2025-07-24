# WTask - Worktree-Based Task Automation

## Overview

WTask는 Claude Squad의 확장 기능으로, Git worktree 기반의 자동화된 태스크 실행 시스템입니다. 각 태스크는 독립적인 워크트리에서 실행되며, 서브태스크가 완료될 때마다 웹훅을 통해 외부 시스템에 알림을 보냅니다.

## Key Features

- **🌳 Worktree 격리**: 각 메인 태스크는 독립된 Git worktree에서 실행
- **📋 순차 실행**: 서브태스크를 정의된 순서대로 실행
- **🔗 Webhook 연동**: 각 서브태스크 완료 시 실시간 알림
- **⏱️ 타임아웃 관리**: 서브태스크별 타임아웃 설정
- **🎯 완료 감지**: 사용자 정의 완료 마커로 태스크 완료 판단
- **🔄 자동 정리**: 태스크 완료 후 워크트리 자동 정리
- **🤖 멀티 AI 지원**: Claude, Gemini, Aider 등 다양한 AI 에이전트 사용 가능

## Installation

WTask는 Claude Squad에 내장되어 있습니다:

```bash
# Claude Squad 설치 후
cs wtask --help
```

## Quick Start

### 1. 태스크 파일 생성

`my-task.json`:
```json
{
  "id": "feature-development",
  "title": "새로운 기능 개발",
  "repo_path": ".",
  "webhook_url": "https://api.example.com/webhooks/progress",
  "subtasks": [
    {
      "id": "create-api",
      "title": "API 엔드포인트 생성",
      "prompt": "REST API 엔드포인트를 만들어주세요.",
      "program": "claude",
      "completion_markers": ["API created", "tests passing"],
      "timeout": "30m"
    }
  ]
}
```

### 2. 태스크 실행

```bash
# 기본 실행 (Claude 사용)
cs wtask my-task.json

# 웹훅 URL 오버라이드
cs wtask my-task.json --webhook https://your-webhook.com

# 타임아웃 설정
cs wtask my-task.json --timeout 1h

# AI 에이전트 변경
cs wtask my-task.json --program claude      # Claude Code (기본값)
cs wtask my-task.json --program gemini      # Google Gemini
cs wtask my-task.json --program aider       # Aider
cs wtask my-task.json --program codex       # OpenAI Codex
```

## CLI Options

### 기본 사용법
```
cs wtask [task-file] [flags]
```

### 플래그 옵션

| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--webhook` | string | - | 태스크 파일의 webhook URL 오버라이드 |
| `--timeout` | string | "30m" | 모든 서브태스크의 기본 타임아웃 |
| `--program` | string | "claude" | 모든 서브태스크의 AI 에이전트 (claude, gemini, aider, codex) |
| `--help` | - | - | 도움말 표시 |

### 타임아웃 형식
- `30m` - 30분
- `1h` - 1시간  
- `2h30m` - 2시간 30분
- `90s` - 90초

## Multi-AI Support

WTask는 Claude Squad의 멀티 AI 에이전트 아키텍처를 활용하여 다양한 AI 도구를 지원합니다.

### 지원되는 AI 에이전트

| AI 에이전트 | 설명 | 사용 예시 |
|-------------|------|-----------|
| **claude** | Claude Code (기본값) | 범용 코딩 작업에 최적화 |
| **gemini** | Google Gemini | 다양한 언어와 복잡한 추론 |
| **aider** | Aider AI | Git 기반 코드 변경에 특화 |
| **codex** | OpenAI Codex | 코드 생성 및 완성 |
| **커스텀** | 사용자 정의 도구 | 로컬 AI 도구나 스크립트 |

### 태스크별 AI 에이전트 설정

```json
{
  "subtasks": [
    {
      "title": "API 설계",
      "program": "claude",
      "prompt": "RESTful API를 설계해주세요..."
    },
    {
      "title": "코드 생성", 
      "program": "gemini",
      "prompt": "위 설계를 바탕으로 Go 코드를 생성해주세요..."
    },
    {
      "title": "Git 커밋",
      "program": "aider",
      "prompt": "변경사항을 검토하고 적절한 커밋 메시지로 커밋해주세요..."
    }
  ]
}
```

### AI 에이전트별 최적 사용 사례

#### Claude Code
- **장점**: 범용성, 안정성, 한국어 지원
- **적합한 작업**: 일반적인 개발, 문서 작성, 코드 리뷰
```json
{
  "program": "claude",
  "prompt": "사용자 인증 시스템을 구현해주세요. JWT를 사용하고..."
}
```

#### Google Gemini  
- **장점**: 멀티모달, 복잡한 추론, 최신 정보
- **적합한 작업**: 복잡한 알고리즘, 데이터 분석, 아키텍처 설계
```json
{
  "program": "gemini", 
  "prompt": "대용량 데이터 처리를 위한 분산 시스템 아키텍처를 설계해주세요..."
}
```

#### Aider
- **장점**: Git 통합, 기존 코드 수정에 특화
- **적합한 작업**: 코드 리팩토링, 버그 수정, 기능 추가
```json
{
  "program": "aider --model gpt-4",
  "prompt": "기존 API에 인증 미들웨어를 추가해주세요..."
}
```

### 설정 우선순위

AI 에이전트 선택 우선순위:
1. **서브태스크의 `program` 필드** (최우선)
2. **CLI `--program` 플래그** 
3. **설정 파일의 `default_program`**
4. **시스템 기본값** (`claude`)

```bash
# 전체 태스크에 Gemini 사용
cs wtask task.json --program gemini

# JSON에서 태스크별로 다른 AI 사용  
# (JSON의 program 필드가 CLI 플래그보다 우선)
```

### 커스텀 AI 도구 사용

```json
{
  "program": "my-custom-ai --model llama3 --temperature 0.3",
  "prompt": "..."
}
```

### AI 에이전트별 완료 마커 최적화

각 AI 에이전트의 특성에 맞는 완료 마커 설정:

```json
{
  "subtasks": [
    {
      "program": "claude",
      "completion_markers": ["✅ 완료", "Task completed", "작업 완료"]
    },
    {
      "program": "gemini", 
      "completion_markers": ["DONE", "Finished", "Complete"]
    },
    {
      "program": "aider",
      "completion_markers": ["Changes committed", "Files updated"]
    }
  ]
}
```

## Path Management

### repo_path 설정

`repo_path`는 Git 저장소의 위치를 지정합니다:

```json
{
  "repo_path": ".",                    // CLI 실행 위치 (권장)
  "repo_path": "/absolute/path/repo",  // 절대 경로
  "repo_path": "../other-repo"         // 상대 경로 (CLI 기준)
}
```

### Path 처리 과정

1. **Git 저장소 검증**: 지정된 경로가 Git 저장소인지 확인
2. **절대 경로 변환**: 상대 경로를 절대 경로로 변환
3. **저장소 루트 탐지**: Git 저장소의 루트 디렉토리 자동 탐지
4. **Worktree 생성**: `~/.claude-squad/worktrees/` 하위에 독립된 작업 공간 생성

### 실행 위치 예시

```bash
# Case 1: 저장소 루트에서 실행 (권장)
cd /path/to/my-project
cs wtask task.json

# Case 2: 서브디렉토리에서 실행  
cd /path/to/my-project/subdirectory
cs wtask task.json  # 자동으로 저장소 루트 탐지

# Case 3: 다른 위치에서 절대경로 지정
cd /anywhere
cs wtask task.json  # task.json에서 repo_path 절대경로 필요
```

## Webhook Integration

### Webhook 이벤트

WTask는 다음 이벤트에 대해 웹훅을 발송합니다:

1. **subtask_started** - 서브태스크 시작
2. **subtask_completed** - 서브태스크 완료 (성공)
3. **subtask_failed** - 서브태스크 실패
4. **maintask_completed** - 메인태스크 완료
5. **maintask_failed** - 메인태스크 실패

### Webhook Payload

```json
{
  "event_type": "subtask_completed",
  "main_task_id": "feature-development",
  "subtask_id": "create-api",
  "status": "success",
  "worktree_path": "/Users/dev/.claude-squad/worktrees/task_abc123",
  "branch_name": "dev/feature-development",
  "timestamp": "2025-01-23T12:00:00Z",
  "output": "API endpoint created successfully...",
  "progress": 33.3,
  "custom_data": {
    "priority": "high",
    "component": "backend"
  }
}
```

### 웹훅 설정

```json
{
  "webhook_url": "https://api.example.com/webhooks/progress",
  "subtasks": [
    {
      "webhook_payload": {
        "priority": "high",
        "component": "backend",
        "assignee": "developer@example.com"
      }
    }
  ]
}
```

## Task Completion Detection

### 완료 마커 설정

서브태스크의 완료는 다음 방법으로 감지됩니다:

```json
{
  "completion_markers": [
    "API endpoint created successfully",
    "All tests passing",
    "✅ Task completed"
  ]
}
```

### 완료 감지 로직

1. **마커 기반**: `completion_markers`에 지정된 문자열이 출력에 나타나면 완료
2. **타임아웃 기반**: 지정된 시간 내에 완료되지 않으면 실패
3. **휴리스틱**: 마커가 없으면 일정 시간 동안 출력이 없을 때 완료로 간주

### 완료 마커 작성 팁

```json
{
  "completion_markers": [
    "✅ Task completed",           // 명확한 완료 표시
    "BUILD SUCCESSFUL",           // 빌드 성공
    "All tests passed",           // 테스트 성공
    "Deployment completed",       // 배포 완료
    "Documentation updated"       // 문서 업데이트
  ]
}
```

## Error Handling

### 에러 유형

1. **구성 에러**: 잘못된 JSON 형식, 필수 필드 누락
2. **경로 에러**: 존재하지 않는 repo_path, Git 저장소 아님
3. **실행 에러**: 프로그램 실행 실패, tmux 세션 생성 실패
4. **타임아웃 에러**: 지정된 시간 내에 완료되지 않음
5. **웹훅 에러**: 웹훅 전송 실패 (재시도됨)

### 에러 처리 전략

```json
{
  "subtasks": [
    {
      "timeout": "30m",           // 충분한 타임아웃 설정
      "completion_markers": [     // 여러 완료 조건 제공
        "Success",
        "Complete",
        "Done"
      ],
      "webhook_payload": {
        "retry_on_failure": true  // 실패 시 재시도 정보
      }
    }
  ]
}
```

## Best Practices

### 1. 태스크 설계

```json
{
  "subtasks": [
    {
      "title": "구체적이고 명확한 제목",
      "prompt": "상세하고 구체적인 지시사항. 예상 결과물과 성공 조건을 명시하세요.",
      "timeout": "20m",          // 보수적인 타임아웃
      "completion_markers": [     // 명확한 완료 조건
        "Build successful",
        "Tests passed",
        "✅ Completed"
      ]
    }
  ]
}
```

### 2. 웹훅 활용

```json
{
  "webhook_url": "https://api.yourservice.com/hooks",
  "subtasks": [
    {
      "webhook_payload": {
        "project": "my-project",
        "environment": "development",
        "notify_channels": ["#dev-team", "#project-updates"]
      }
    }
  ]
}
```

### 3. 프롬프트 작성

```json
{
  "prompt": "React 컴포넌트를 생성해주세요. 요구사항: 1) TypeScript 사용, 2) Props 인터페이스 정의, 3) 단위 테스트 포함, 4) Storybook 스토리 생성. 완료되면 '✅ Component ready'를 출력해주세요."
}
```

### 4. 타임아웃 설정

```json
{
  "subtasks": [
    {
      "title": "간단한 설정 변경",
      "timeout": "5m"
    },
    {
      "title": "복잡한 기능 개발", 
      "timeout": "45m"
    },
    {
      "title": "전체 시스템 테스트",
      "timeout": "1h30m"
    }
  ]
}
```

## Troubleshooting

### 일반적인 문제들

1. **"not a git repository" 에러**
   ```bash
   # 해결: Git 저장소에서 실행하거나 올바른 repo_path 설정
   cd /path/to/git/repo
   cs wtask task.json
   ```

2. **태스크가 완료되지 않음**
   ```json
   {
     "completion_markers": ["success", "done", "complete"],
     "timeout": "60m"  // 타임아웃 증가
   }
   ```

3. **웹훅 전송 실패**
   - 웹훅 URL 확인
   - 네트워크 연결 확인
   - 서버 응답 상태 확인 (재시도 자동 처리)

### 디버깅

```bash
# 로그 확인
tail -f /var/folders/.../claudesquad.log

# 태스크 상태 확인  
cs wtask task.json --verbose  # (향후 구현 예정)
```

## Advanced Usage

### 복잡한 워크플로우

```json
{
  "id": "full-stack-feature",
  "title": "풀스택 기능 개발",
  "subtasks": [
    {
      "id": "backend-api",
      "title": "백엔드 API 개발",
      "prompt": "User 관리 API를 개발해주세요. CRUD 작업과 인증이 포함되어야 합니다.",
      "timeout": "45m",
      "completion_markers": ["API tests passing", "Swagger docs updated"]
    },
    {
      "id": "frontend-ui",
      "title": "프론트엔드 UI 개발", 
      "prompt": "위에서 만든 API를 사용하는 React 컴포넌트를 만들어주세요.",
      "timeout": "40m",
      "completion_markers": ["Component rendered", "UI tests passing"]
    },
    {
      "id": "integration-test",
      "title": "통합 테스트",
      "prompt": "전체 시스템의 통합 테스트를 실행하고 결과를 확인해주세요.",
      "timeout": "20m",
      "completion_markers": ["All tests passed", "Coverage report generated"]
    }
  ]
}
```

### 프로덕션 배포 워크플로우

```json
{
  "id": "production-deployment",
  "title": "프로덕션 배포",
  "webhook_url": "https://api.company.com/deployment-webhooks",
  "subtasks": [
    {
      "id": "build-test",
      "title": "빌드 및 테스트",
      "prompt": "프로덕션 빌드를 생성하고 모든 테스트를 실행해주세요.",
      "timeout": "30m",
      "webhook_payload": {
        "stage": "build",
        "environment": "production",
        "notify": ["devops-team"]
      }
    },
    {
      "id": "deploy",
      "title": "배포 실행",
      "prompt": "프로덕션 환경에 애플리케이션을 배포해주세요.",
      "timeout": "15m",
      "webhook_payload": {
        "stage": "deploy",
        "environment": "production",
        "notify": ["all-hands"]
      }
    },
    {
      "id": "health-check",
      "title": "헬스 체크",
      "prompt": "배포된 애플리케이션의 헬스 체크를 수행해주세요.",
      "timeout": "10m",
      "webhook_payload": {
        "stage": "verification", 
        "environment": "production"
      }
    }
  ]
}
```

## Migration Guide

### 기존 Claude Squad 세션에서 WTask로 마이그레이션

1. **기존 워크플로우 분석**
2. **서브태스크로 분할**
3. **완료 조건 정의**
4. **웹훅 설정**
5. **테스트 및 검증**

WTask는 기존 Claude Squad와 함께 사용할 수 있으며, 복잡한 멀티스텝 작업에 특히 유용합니다.