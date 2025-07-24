# Claude Squad Path Management Analysis

## 기존 Claude Squad의 Path 관리 방식

### 1. 기본 Path 구조

**실행 위치 검증:**
- `main.go:41`: CLI 실행 시 현재 디렉토리를 절대 경로로 변환
- `main.go:46-47`: `git.IsGitRepo(currentDir)`로 Git 저장소인지 검증
- **Claude Squad는 반드시 Git 저장소 내에서만 실행 가능**

**세션 생성 시 Path:**
```go
// app/app.go:476, 498
instance, err := session.NewInstance(session.InstanceOptions{
    Title:   "",
    Path:    ".",        // 항상 현재 디렉토리(".")를 사용
    Program: m.program,
})
```

### 2. Path 변환 과정

**NewInstance에서의 처리:**
```go
// session/instance.go:158-162
absPath, err := filepath.Abs(opts.Path)  // "." → 절대경로 변환
if err != nil {
    return nil, fmt.Errorf("failed to get absolute path: %w", err)
}
```

**Git Worktree 생성:**
```go
// session/git/worktree.go:51-56
absPath, err := filepath.Abs(repoPath)   // 다시 한번 절대경로 변환
repoPath, err = findGitRepoRoot(absPath) // Git 저장소 루트 찾기
```

### 3. Worktree 저장 위치

**Worktree 디렉토리:**
```go
// session/git/worktree.go:11-17
func getWorktreeDirectory() (string, error) {
    configDir, err := config.GetConfigDir()  // ~/.claude-squad/
    return filepath.Join(configDir, "worktrees"), nil
}
```

**실제 Worktree 경로:**
```go
// session/git/worktree.go:68-69
worktreePath := filepath.Join(worktreeDir, sanitizedName)
worktreePath = worktreePath + "_" + fmt.Sprintf("%x", time.Now().UnixNano())
// 결과: ~/.claude-squad/worktrees/session-name_1234567890abcdef
```

## repo_path의 의미

### 현재 구현에서의 repo_path:

1. **CLI 실행 위치 기준**
   - `repo_path: "."` → CLI가 실행된 현재 디렉토리
   - 상대경로도 가능하지만, 내부적으로 절대경로로 변환됨

2. **Git 저장소 루트 자동 탐지**
   - `findGitRepoRoot()` 함수가 주어진 경로에서 Git 저장소 루트를 찾음
   - 예: `/path/to/repo/subdir/` → `/path/to/repo/` (Git 루트)

3. **Worktree 생성 위치**
   - 모든 worktree는 `~/.claude-squad/worktrees/` 하위에 생성
   - 원본 저장소 위치와 무관하게 격리된 공간 사용

## wtask의 Path 처리

### 현재 wtask 구현:
```go
// cmd/wtask.go:195-206
if mainTask.RepoPath == "" {
    cwd, err := os.Getwd()           // CLI 실행 디렉토리
    mainTask.RepoPath = cwd
} else {
    absPath, err := filepath.Abs(mainTask.RepoPath)  // 절대경로 변환
    mainTask.RepoPath = absPath
}
```

### Path 사용 예시:

**JSON 설정에서:**
```json
{
  "repo_path": ".",                    // CLI 실행 위치
  "repo_path": "/absolute/path/to/repo", // 절대경로 
  "repo_path": "../other-repo",         // 상대경로 (절대경로로 변환됨)
}
```

## 멀티 세션 Path 관리 방식

### 1. 세션 격리
- 각 세션은 독립된 Git worktree에서 실행
- Worktree 위치: `~/.claude-squad/worktrees/{session-name}_{timestamp}/`
- 모든 세션이 같은 Git 저장소를 기반으로 하지만 독립된 브랜치에서 작업

### 2. 동일한 저장소, 다른 브랜치
```
원본 저장소: /Users/dev/my-project/
├── main 브랜치 (원본)
├── ~/.claude-squad/worktrees/session-1_abc123/ → branch: user/session-1  
├── ~/.claude-squad/worktrees/session-2_def456/ → branch: user/session-2
└── ~/.claude-squad/worktrees/task-xyz_789abc/ → branch: user/task-xyz
```

### 3. Branch 명명 규칙
```go
// session/git/worktree.go:48
branchName := fmt.Sprintf("%s%s", cfg.BranchPrefix, sanitizedName)
// 결과: "username/session-name" 형태
```

## 권장사항

### wtask 사용 시:
1. **현재 디렉토리에서 실행** (가장 간단)
   ```bash
   cd /path/to/your/repo
   cs wtask task.json
   ```

2. **JSON에서 절대경로 지정**
   ```json
   {
     "repo_path": "/Users/dev/my-project"
   }
   ```

3. **CLI 플래그로 경로 오버라이드** (향후 구현 가능)
   ```bash
   cs wtask task.json --repo-path /path/to/repo
   ```

### 주의사항:
- **반드시 Git 저장소 내에서 실행해야 함**
- 상대경로는 CLI 실행 위치 기준으로 해석됨  
- 각 MainTask는 독립된 worktree에서 실행되므로 서로 간섭하지 않음