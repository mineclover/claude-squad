# WTask JSON Schema Specification

## Overview

WTask는 JSON 형식의 태스크 정의 파일을 사용합니다. 이 문서는 JSON 스키마의 상세 사양을 설명합니다.

## Schema Version

현재 스키마 버전: `1.0`

## Root Schema - MainTask

### 전체 구조

```json
{
  "id": "string",
  "title": "string", 
  "repo_path": "string",
  "webhook_url": "string",
  "subtasks": [SubTask]
}
```

### 필드 상세 설명

#### `id` (string, optional)
- **설명**: 메인태스크의 고유 식별자
- **기본값**: 자동 생성 (title 기반 + 타임스탬프)
- **형식**: 영숫자, 하이픈, 언더스코어만 허용
- **예시**: `"feature-api-development"`, `"hotfix-bug-123"`

```json
{
  "id": "user-management-system"
}
```

#### `title` (string, required)
- **설명**: 메인태스크의 사람이 읽을 수 있는 제목
- **제약**: 1-100자, 비어있을 수 없음
- **용도**: 로그, 웹훅, UI 표시

```json
{
  "title": "사용자 관리 시스템 개발"
}
```

#### `repo_path` (string, optional)
- **설명**: Git 저장소의 경로
- **기본값**: CLI 실행 위치 (`"."`)
- **형식**: 절대경로 또는 CLI 실행 위치 기준 상대경로
- **처리**: 내부적으로 절대경로로 변환됨

```json
{
  "repo_path": ".",                        // CLI 실행 위치
  "repo_path": "/Users/dev/my-project",    // 절대경로
  "repo_path": "../other-project"          // 상대경로
}
```

#### `webhook_url` (string, optional)
- **설명**: 태스크 진행 상황을 전송할 웹훅 URL
- **형식**: 유효한 HTTP/HTTPS URL
- **오버라이드**: CLI 플래그 `--webhook`로 덮어쓰기 가능

```json
{
  "webhook_url": "https://api.example.com/webhooks/task-progress"
}
```

#### `subtasks` (array of SubTask, required)
- **설명**: 순차적으로 실행할 서브태스크 목록
- **제약**: 최소 1개 이상
- **처리**: 배열 순서대로 실행

## SubTask Schema

### 전체 구조

```json
{
  "id": "string",
  "title": "string",
  "prompt": "string", 
  "program": "string",
  "completion_markers": ["string"],
  "timeout": "string",
  "webhook_payload": {}
}
```

### 필드 상세 설명

#### `id` (string, optional)
- **설명**: 서브태스크의 고유 식별자
- **기본값**: 자동 생성 (`{main_task_id}-subtask-{index}`)
- **형식**: 영숫자, 하이픈, 언더스코어만 허용
- **용도**: 웹훅, 로그에서 식별

```json
{
  "id": "create-user-api"
}
```

#### `title` (string, required)
- **설명**: 서브태스크의 사람이 읽을 수 있는 제목
- **제약**: 1-100자, 비어있을 수 없음
- **용도**: 진행 상황 표시, 로그

```json
{
  "title": "사용자 생성 API 엔드포인트 개발"
}
```

#### `prompt` (string, required)
- **설명**: AI 에이전트에게 전달할 지시사항
- **제약**: 1-10000자, 구체적이고 명확해야 함
- **팁**: 완료 조건과 기대 결과를 명시

```json
{
  "prompt": "사용자 생성을 위한 REST API 엔드포인트를 개발해주세요. 요구사항: 1) POST /api/users 엔드포인트, 2) 요청 유효성 검사, 3) 데이터베이스 저장, 4) JSON 응답, 5) 에러 처리. 완료되면 '✅ User API created'를 출력해주세요."
}
```

#### `program` (string, optional)
- **설명**: 사용할 AI 에이전트/도구
- **기본값**: `"claude"` (설정파일의 default_program)
- **지원 에이전트**: 
  - `"claude"` - Claude Code (기본값, 범용 코딩)
  - `"gemini"` - Google Gemini (복잡한 추론, 멀티모달)
  - `"aider"` - Aider AI (Git 통합, 코드 수정 특화)
  - `"codex"` - OpenAI Codex (코드 생성)
  - 커스텀 프로그램 (`"my-ai-tool --args"`)
- **오버라이드**: CLI 플래그 `--program`으로 덮어쓰기 가능

```json
{
  "program": "claude",              // Claude Code 사용
  "program": "gemini",              // Google Gemini 사용  
  "program": "aider --model gpt-4", // Aider with GPT-4
  "program": "custom-ai-tool"       // 커스텀 도구
}
```

#### `completion_markers` (array of string, optional)
- **설명**: 태스크 완료를 감지할 문자열 목록
- **처리**: OR 조건 (하나라도 매치되면 완료)
- **기본값**: 휴리스틱 감지 (출력 없음 기반)
- **권장**: 명확한 완료 표시 사용

```json
{
  "completion_markers": [
    "✅ Task completed",
    "BUILD SUCCESSFUL", 
    "All tests passed",
    "API endpoint created"
  ]
}
```

#### `timeout` (string, required)
- **설명**: 태스크 타임아웃 시간
- **형식**: Go duration 형식 (`"30m"`, `"1h30m"`, `"2h"`)
- **기본값**: CLI 플래그 `--timeout` 또는 `"30m"`
- **권장**: 보수적으로 설정

```json
{
  "timeout": "45m"
}
```

**타임아웃 형식 예시:**
- `"30s"` - 30초
- `"5m"` - 5분  
- `"1h"` - 1시간
- `"2h30m"` - 2시간 30분
- `"90m"` - 90분 (1시간 30분과 동일)

#### `webhook_payload` (object, optional)
- **설명**: 웹훅과 함께 전송할 커스텀 데이터
- **형식**: JSON 객체 (중첩 가능)
- **용도**: 외부 시스템에서 태스크 식별 및 처리

```json
{
  "webhook_payload": {
    "priority": "high",
    "component": "backend",
    "assignee": "john.doe@company.com",
    "labels": ["api", "user-management"],
    "estimated_hours": 3,
    "metadata": {
      "project_id": "proj-123",
      "sprint": "2025-Q1-S3"
    }
  }
}
```

## Schema Validation Rules

### 전역 규칙

1. **JSON 형식**: 유효한 JSON 문법
2. **UTF-8 인코딩**: 모든 텍스트는 UTF-8
3. **필수 필드**: `title`, `subtasks`
4. **배열 크기**: `subtasks` 최소 1개

### 필드별 검증

#### 문자열 필드
- **null 불허**: 모든 문자열 필드는 null일 수 없음
- **공백 처리**: 앞뒤 공백 자동 제거
- **최대 길이**: 제한 초과 시 오류

#### 배열 필드
- **빈 배열**: `subtasks`는 빈 배열 불허
- **중복 ID**: 동일한 `id` 값 불허

#### URL 필드
- **형식 검증**: `webhook_url`은 유효한 HTTP/HTTPS URL
- **접근성**: URL 유효성은 런타임에 검증

### 예시 검증 오류

```json
{
  "error": "validation_failed",
  "details": [
    {
      "field": "title",
      "message": "title cannot be empty"
    },
    {
      "field": "subtasks[0].timeout", 
      "message": "invalid timeout format: '30minutes'"
    },
    {
      "field": "webhook_url",
      "message": "invalid URL format"
    }
  ]
}
```

## Complete Schema Examples

### 최소 설정

```json
{
  "title": "Simple Task",
  "subtasks": [
    {
      "title": "Do Something",
      "prompt": "Please do something useful.",
      "timeout": "10m"
    }
  ]
}
```

### 완전한 설정

```json
{
  "id": "comprehensive-feature-development",
  "title": "포괄적인 기능 개발",
  "repo_path": "/Users/dev/my-awesome-project",
  "webhook_url": "https://api.mycompany.com/webhooks/development-progress",
  "subtasks": [
    {
      "id": "backend-development", 
      "title": "백엔드 API 개발",
      "prompt": "사용자 인증과 프로필 관리를 위한 REST API를 개발해주세요. OpenAPI 문서도 포함해서 완료되면 '🚀 Backend API ready'를 출력해주세요.",
      "program": "claude",
      "completion_markers": [
        "🚀 Backend API ready",
        "OpenAPI documentation generated",
        "All API tests passing"
      ],
      "timeout": "45m",
      "webhook_payload": {
        "component": "backend",
        "priority": "high", 
        "team": "backend-team",
        "estimated_story_points": 8,
        "dependencies": [],
        "tags": ["api", "authentication", "profile"]
      }
    },
    {
      "id": "frontend-development",
      "title": "프론트엔드 UI 개발", 
      "prompt": "위에서 개발된 API를 사용하는 React 컴포넌트를 만들어주세요. TypeScript를 사용하고 단위테스트도 포함해주세요. 완료되면 '✨ Frontend components ready'를 출력해주세요.",
      "program": "claude",
      "completion_markers": [
        "✨ Frontend components ready",
        "TypeScript types defined",
        "Unit tests written",
        "Storybook stories created"
      ],
      "timeout": "40m",
      "webhook_payload": {
        "component": "frontend",
        "priority": "high",
        "team": "frontend-team", 
        "estimated_story_points": 5,
        "dependencies": ["backend-development"],
        "tags": ["react", "typescript", "ui"]
      }
    },
    {
      "id": "integration-testing",
      "title": "통합 테스트",
      "prompt": "전체 시스템의 end-to-end 테스트를 작성하고 실행해주세요. API와 UI가 올바르게 연동되는지 확인해주세요. 완료되면 '🎯 Integration tests passed'를 출력해주세요.",
      "program": "claude",
      "completion_markers": [
        "🎯 Integration tests passed",
        "All E2E scenarios covered", 
        "Test coverage above 90%"
      ],
      "timeout": "30m",
      "webhook_payload": {
        "component": "testing",
        "priority": "medium",
        "team": "qa-team",
        "estimated_story_points": 3,
        "dependencies": ["backend-development", "frontend-development"],
        "tags": ["e2e", "integration", "testing"]
      }
    },
    {
      "id": "documentation-update",
      "title": "문서 업데이트",
      "prompt": "개발된 기능에 대한 문서를 업데이트해주세요. README.md, API 문서, 사용자 가이드를 포함해주세요. 완료되면 '📚 Documentation updated'를 출력해주세요.",
      "program": "claude", 
      "completion_markers": [
        "📚 Documentation updated",
        "README.md updated",
        "API documentation complete",
        "User guide written"
      ],
      "timeout": "20m",
      "webhook_payload": {
        "component": "documentation",
        "priority": "low",
        "team": "tech-writing",
        "estimated_story_points": 2,
        "dependencies": ["integration-testing"],
        "tags": ["docs", "readme", "guide"]
      }
    }
  ]
}
```

## JSON Schema (Draft 7)

참고용 JSON Schema 정의:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "WTask MainTask Schema",
  "required": ["title", "subtasks"],
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^[a-zA-Z0-9_-]+$",
      "maxLength": 50
    },
    "title": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "repo_path": {
      "type": "string",
      "minLength": 1
    },
    "webhook_url": {
      "type": "string",
      "format": "uri",
      "pattern": "^https?://"
    },
    "subtasks": {
      "type": "array",
      "minItems": 1,
      "items": {
        "$ref": "#/definitions/subtask"
      }
    }
  },
  "definitions": {
    "subtask": {
      "type": "object", 
      "required": ["title", "prompt", "timeout"],
      "properties": {
        "id": {
          "type": "string",
          "pattern": "^[a-zA-Z0-9_-]+$",
          "maxLength": 50
        },
        "title": {
          "type": "string",
          "minLength": 1,
          "maxLength": 100
        },
        "prompt": {
          "type": "string",
          "minLength": 1,
          "maxLength": 10000
        },
        "program": {
          "type": "string",
          "minLength": 1
        },
        "completion_markers": {
          "type": "array",
          "items": {
            "type": "string",
            "minLength": 1
          }
        },
        "timeout": {
          "type": "string",
          "pattern": "^[0-9]+[smh]$|^[0-9]+h[0-9]+m$"
        },
        "webhook_payload": {
          "type": "object"
        }
      }
    }
  }
}
```

## Migration from Previous Versions

현재는 첫 번째 버전이므로 마이그레이션이 필요하지 않습니다. 향후 스키마 변경 시 이 섹션에서 마이그레이션 가이드를 제공합니다.

## Validation Tools

JSON 스키마 검증을 위한 도구들:

1. **온라인 검증**: [jsonschemavalidator.net](https://www.jsonschemavalidator.net)
2. **CLI 도구**: `ajv-cli`, `jsonschema`
3. **IDE 플러그인**: VSCode JSON Schema 확장
4. **내장 검증**: WTask CLI에서 자동 검증