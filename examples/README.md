# WTask Examples

이 디렉토리는 다양한 시나리오에서 WTask를 사용하는 예시들을 포함합니다.

## 예시 파일들

### 1. `simple-task.json` - 초보자용
- **난이도**: ⭐
- **소요시간**: ~5분  
- **설명**: 간단한 Hello World 프로그램 생성
- **학습 목표**: 기본적인 WTask 구조 이해

```bash
cs wtask examples/simple-task.json
```

### 2. `api-development.json` - 백엔드 개발
- **난이도**: ⭐⭐⭐
- **소요시간**: ~2시간
- **설명**: Node.js Express REST API 풀스택 개발
- **포함 기능**: 
  - 프로젝트 초기화
  - MongoDB 모델링
  - JWT 인증
  - CRUD API
  - 유효성 검사
  - 테스트 작성
  - API 문서화

```bash
cs wtask examples/api-development.json --timeout 45m
```

### 3. `frontend-development.json` - 프론트엔드 개발
- **난이도**: ⭐⭐⭐
- **소요시간**: ~2.5시간
- **설명**: React TypeScript 프론트엔드 개발
- **포함 기능**:
  - React 프로젝트 설정
  - 컴포넌트 개발
  - 라우팅 설정
  - 폼 처리
  - API 연동
  - 스타일링
  - 테스트 작성
  - 빌드 최적화

```bash
cs wtask examples/frontend-development.json --webhook YOUR_SLACK_WEBHOOK
```

### 4. `go-api-example.json` - Go API 개발
- **난이도**: ⭐⭐
- **소요시간**: ~37분
- **설명**: Go를 사용한 간단한 REST API 개발
- **포함 기능**:
  - Go REST API 서버 생성
  - JSON 응답 처리
  - 에러 처리
  - 단위 테스트 작성
  - API 문서화

```bash
cs wtask examples/go-api-example.json
```

### 5. `gemini-task.json` - Google Gemini 활용
- **난이도**: ⭐⭐⭐
- **소요시간**: ~1.5시간
- **설명**: Google Gemini를 사용한 AI 모델 성능 비교 및 분석
- **포함 기능**:
  - AI 모델 연구 및 분석
  - 복잡한 알고리즘 설계
  - 데이터 분석 및 시각화

```bash
cs wtask examples/gemini-task.json --program gemini
```

### 6. `multi-ai-task.json` - 멀티 AI 협업
- **난이도**: ⭐⭐⭐⭐
- **소요시간**: ~2.5시간  
- **설명**: 여러 AI 에이전트가 협업하는 복합 프로젝트
- **포함 기능**:
  - Claude: 프로젝트 계획 및 백엔드 구현
  - Gemini: 시스템 아키텍처 설계 및 테스트
  - Aider: 코드 리팩토링 및 최적화

```bash
cs wtask examples/multi-ai-task.json
```

### 7. `devops-deployment.json` - DevOps 파이프라인
- **난이도**: ⭐⭐⭐⭐
- **소요시간**: ~3시간
- **설명**: 완전한 CI/CD 파이프라인 구축
- **포함 기능**:
  - Docker 컨테이너화
  - GitHub Actions CI/CD
  - Kubernetes 배포
  - 모니터링 설정
  - 시크릿 관리
  - 백업 전략
  - 부하 테스트
  - 운영 문서화

```bash
cs wtask examples/devops-deployment.json --timeout 1h
```

## 사용 방법

### 1. 예시 복사 및 수정

```bash
# 예시 파일을 복사해서 수정
cp examples/simple-task.json my-task.json

# 웹훅 URL 수정
# JSON 파일에서 webhook_url을 본인의 URL로 변경
```

### 2. CLI 옵션 활용

```bash
# 웹훅 URL 오버라이드
cs wtask examples/api-development.json --webhook https://your-webhook.com

# 타임아웃 조정
cs wtask examples/frontend-development.json --timeout 2h

# 프로그램 변경
cs wtask examples/simple-task.json --program aider
```

### 3. 상황별 예시 선택

| 시나리오 | 추천 예시 | 설명 |
|----------|-----------|------|
| WTask 학습 | `simple-task.json` | 기본 기능 이해 |
| Go API 개발 | `go-api-example.json` | 간단한 Go REST API |
| Node.js API 개발 | `api-development.json` | 백엔드 개발 워크플로우 |
| 웹 앱 개발 | `frontend-development.json` | React 프론트엔드 개발 |
| AI 연구/분석 | `gemini-task.json` | Gemini를 활용한 고급 분석 |
| 멀티 AI 협업 | `multi-ai-task.json` | 여러 AI 에이전트 협력 |
| 인프라 구축 | `devops-deployment.json` | DevOps 파이프라인 |

## 커스터마이징 가이드

### Webhook URL 설정

```json
{
  "webhook_url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}
```

**인기 있는 웹훅 서비스:**
- Slack Webhooks
- Discord Webhooks  
- Microsoft Teams
- 커스텀 API 엔드포인트

### 완료 마커 커스터마이징

```json
{
  "completion_markers": [
    "✅ 원하는 완료 메시지",
    "SUCCESS: 빌드 완료",
    "🎉 배포 성공"
  ]
}
```

### 타임아웃 조정

```json
{
  "timeout": "45m"    // 복잡한 작업
  "timeout": "10m"    // 간단한 작업  
  "timeout": "2h"     // 매우 복잡한 작업
}
```

### 커스텀 프로그램 사용

```json
{
  "program": "aider --model gpt-4"  // Aider with GPT-4
  "program": "claude"               // Claude Code (기본값)
  "program": "codex"                // OpenAI Codex
}
```

## 트러블슈팅

### 일반적인 문제들

1. **태스크가 완료되지 않음**
   - `completion_markers` 확인
   - `timeout` 값 증가
   - 프롬프트 명확성 개선

2. **웹훅 전송 실패**
   - URL 유효성 확인
   - 네트워크 연결 확인
   - 웹훅 서비스 상태 확인

3. **Git 저장소 오류**
   - Git 저장소에서 실행하는지 확인
   - `repo_path` 설정 확인

### 디버깅 팁

```bash
# 로그 확인
tail -f /tmp/claudesquad.log

# 간단한 태스크로 먼저 테스트
cs wtask examples/simple-task.json

# 웹훅 없이 테스트 
# JSON에서 webhook_url 제거 또는 빈 문자열로 설정
```

## 기여하기

새로운 예시를 추가하고 싶다면:

1. 명확한 제목과 설명
2. 적절한 난이도 표시
3. 예상 소요 시간
4. 완료 마커 설정
5. README 업데이트

## 피드백

예시 개선을 위한 피드백은 언제든 환영합니다:
- 이슈 생성
- Pull Request
- 토론 참여