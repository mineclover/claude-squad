# Claude Squad with WTask - Build and Install Guide

이 가이드는 WTask 기능이 포함된 Claude Squad를 로컬에서 빌드하고 설치하는 방법을 안내합니다.

## Prerequisites

### 필수 요구사항

1. **Go 1.21 이상**
   ```bash
   go version  # Go 버전 확인
   ```

2. **Git**
   ```bash
   git --version  # Git 버전 확인
   ```

3. **tmux** (Claude Squad 실행에 필요)
   ```bash
   # macOS (Homebrew)
   brew install tmux
   
   # Ubuntu/Debian
   sudo apt-get install tmux
   
   # CentOS/RHEL
   sudo yum install tmux
   ```

4. **GitHub CLI (gh)** (PR 생성 등에 필요)
   ```bash
   # macOS (Homebrew)
   brew install gh
   
   # Ubuntu/Debian
   sudo apt-get install gh
   ```

## Build Process

### 1. 저장소 클론 및 이동

```bash
git clone https://github.com/smtg-ai/claude-squad.git
cd claude-squad
```

### 2. 의존성 설치

```bash
go mod tidy
```

### 3. 빌드

```bash
# 개발용 빌드 (현재 플랫폼용)
go build -o cs

# 릴리즈용 빌드 (최적화)
go build -ldflags "-s -w" -o cs

# 크로스 플랫폼 빌드 예시
GOOS=linux GOARCH=amd64 go build -o cs-linux-amd64
GOOS=windows GOARCH=amd64 go build -o cs-windows-amd64.exe
GOOS=darwin GOARCH=arm64 go build -o cs-darwin-arm64
```

### 4. 빌드 확인

```bash
./cs version
./cs wtask --help
```

## Installation Methods

### Method 1: Local Installation (권장)

로컬 사용자 디렉토리에 설치하는 방법입니다.

```bash
# 1. 로컬 bin 디렉토리 생성
mkdir -p ~/.local/bin

# 2. 빌드된 바이너리 복사
cp ./cs ~/.local/bin/claude-squad
chmod +x ~/.local/bin/claude-squad

# 3. PATH에 추가 (셸에 따라 다름)
# Bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Fish
fish_add_path ~/.local/bin
```

### Method 2: Global Installation

시스템 전체에서 사용할 수 있도록 설치하는 방법입니다.

```bash
# 시스템 bin 디렉토리에 설치 (관리자 권한 필요)
sudo cp ./cs /usr/local/bin/claude-squad
sudo chmod +x /usr/local/bin/claude-squad

# 또는 기존 Claude Squad 교체 (Homebrew 설치된 경우)
sudo cp ./cs /opt/homebrew/bin/claude-squad
```

### Method 3: 개발 환경에서 직접 실행

빌드 후 바로 실행하는 방법입니다.

```bash
# 현재 디렉토리에서 직접 실행
./cs wtask examples/simple-task.json

# 또는 go run으로 직접 실행
go run . wtask examples/simple-task.json
```

## Installation Verification

설치가 완료되면 다음 명령으로 확인할 수 있습니다:

```bash
# 버전 확인
claude-squad version

# 또는 (별칭 설정한 경우)
cs version

# WTask 명령 확인
claude-squad wtask --help

# 예시 실행 테스트
claude-squad wtask examples/simple-task.json --webhook https://httpbin.org/post
```

## Shell Aliases Setup

편의를 위해 별칭을 설정할 수 있습니다:

```bash
# Bash/Zsh
echo 'alias cs="claude-squad"' >> ~/.bashrc  # 또는 ~/.zshrc
source ~/.bashrc  # 또는 source ~/.zshrc

# Fish
alias cs="claude-squad"
funcsave cs
```

## Development Workflow

개발 중인 경우 다음 워크플로우를 사용하세요:

### 1. 코드 수정 후 빌드 및 테스트

```bash
# 빌드
go build -o cs

# 단위 테스트 실행
go test ./...

# WTask 기능 테스트
./cs wtask examples/simple-task.json
```

### 2. 실시간 개발

```bash
# 파일 변경 감지하여 자동 빌드 (air 설치 필요)
go install github.com/cosmtrek/air@latest
air

# 또는 직접 go run 사용
go run . wtask examples/simple-task.json
```

### 3. 디버깅

```bash
# 로그 레벨 증가 (환경변수)
export LOG_LEVEL=debug
./cs wtask examples/simple-task.json

# 로그 파일 확인
tail -f /tmp/claudesquad.log
```

## 다중 버전 관리

여러 버전을 관리해야 하는 경우:

```bash
# 버전별 설치
cp ./cs ~/.local/bin/claude-squad-wtask
cp ./cs ~/.local/bin/claude-squad-dev

# 심볼릭 링크로 현재 버전 지정
ln -sf ~/.local/bin/claude-squad-wtask ~/.local/bin/claude-squad

# 버전 전환
ln -sf ~/.local/bin/claude-squad-dev ~/.local/bin/claude-squad
```

## Troubleshooting

### 일반적인 문제들

#### 1. `command not found: claude-squad`

**해결방법:**
```bash
# PATH 확인
echo $PATH

# PATH에 설치 디렉토리 추가
export PATH="$HOME/.local/bin:$PATH"

# 또는 설치 위치 확인
which claude-squad
```

#### 2. `permission denied`

**해결방법:**
```bash
# 실행 권한 부여
chmod +x ~/.local/bin/claude-squad

# 또는 소유권 확인
ls -la ~/.local/bin/claude-squad
```

#### 3. `unknown command "wtask"`

이는 오래된 버전의 Claude Squad가 설치되어 있는 경우입니다.

**해결방법:**
```bash
# 현재 사용 중인 바이너리 확인
which claude-squad

# WTask 기능이 있는지 확인
claude-squad wtask --help

# 없다면 우리가 빌드한 버전으로 교체
cp ./cs $(which claude-squad)
```

#### 4. Build 오류

**해결방법:**
```bash
# Go 모듈 정리
go mod tidy

# 캐시 정리
go clean -modcache

# 의존성 다시 다운로드
go mod download
```

#### 5. WTask 실행 오류

**해결방법:**
```bash
# Git 저장소에서 실행하는지 확인
git status

# tmux 설치 확인
tmux -V

# 로그 파일 확인
tail -f /tmp/claudesquad.log
```

## Configuration

### 설정 파일 위치

Claude Squad는 다음 위치에서 설정을 찾습니다:

```bash
# 사용자별 설정
~/.config/claude-squad/config.yaml

# 프로젝트별 설정
./.claude-squad/config.yaml

# 환경 변수
export CLAUDE_SQUAD_CONFIG_DIR=/path/to/config
```

### WTask 기본 설정

WTask 관련 기본 설정:

```yaml
# ~/.config/claude-squad/config.yaml
wtask:
  default_program: "claude"
  default_timeout: "30m"
  webhook_retry_count: 3
  worktree_cleanup: true
```

## Advanced Installation

### Docker를 사용한 설치

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -ldflags "-s -w" -o claude-squad

FROM alpine:latest
RUN apk --no-cache add git tmux
COPY --from=builder /app/claude-squad /usr/local/bin/
ENTRYPOINT ["claude-squad"]
```

```bash
# Docker 빌드 및 실행
docker build -t claude-squad-wtask .
docker run -v $(pwd):/workspace claude-squad-wtask wtask examples/simple-task.json
```

### CI/CD 환경에서 설치

```yaml
# GitHub Actions 예시
- name: Install Claude Squad with WTask
  run: |
    git clone https://github.com/smtg-ai/claude-squad.git
    cd claude-squad
    go build -o claude-squad
    sudo mv claude-squad /usr/local/bin/
    claude-squad wtask --help
```

## Update Process

새로운 기능이 추가되거나 버그가 수정된 경우:

```bash
# 1. 최신 코드 가져오기
git pull origin main

# 2. 리빌드
go build -o cs

# 3. 재설치
cp ./cs ~/.local/bin/claude-squad

# 4. 확인
claude-squad version
claude-squad wtask --help
```

## Uninstall

Claude Squad with WTask를 완전히 제거하려면:

```bash
# 바이너리 제거
rm ~/.local/bin/claude-squad
# 또는
sudo rm /usr/local/bin/claude-squad

# 설정 파일 제거 (선택사항)
rm -rf ~/.config/claude-squad

# PATH에서 제거 (해당하는 경우)
# ~/.bashrc, ~/.zshrc 등에서 PATH 설정 라인 제거

# 별칭 제거 (해당하는 경우)
unalias cs
```

## Additional Resources

- [WTask Documentation](WTASK_DOCUMENTATION.md)
- [Task Schema Reference](TASK_SCHEMA.md)
- [Examples](examples/)
- [GitHub Repository](https://github.com/smtg-ai/claude-squad)

## Support

문제가 발생하면:

1. 이 가이드의 Troubleshooting 섹션 확인
2. [GitHub Issues](https://github.com/smtg-ai/claude-squad/issues) 검색
3. 새로운 이슈 생성 시 다음 정보 포함:
   - OS 및 버전
   - Go 버전
   - 실행한 명령
   - 전체 에러 메시지
   - 로그 파일 내용