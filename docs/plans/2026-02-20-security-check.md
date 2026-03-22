# 敏感信息泄露检查与清理 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复已上传到公开 GitHub 仓库中的敏感信息泄露问题，移除硬编码加密密钥，完善 .gitignore。

**Architecture:** 修改加密模块强制要求环境变量配置，增强 .gitignore 防止未来敏感文件被提交。

**Tech Stack:** Go, Git

---

### Task 1: 移除硬编码加密密钥

**Files:**
- Modify: `internal/crypto/encryption.go:21-42`

**Step 1: 修改 InitKey 函数，移除硬编码密钥**

将第 21-42 行的 `InitKey` 函数替换为：

```go
// InitKey 初始化加密密钥
func InitKey() error {
	var initErr error
	once.Do(func() {
		// 从环境变量获取密钥
		key := os.Getenv("RUBICK_ENCRYPTION_KEY")
		if key == "" {
			initErr = errors.New("RUBICK_ENCRYPTION_KEY environment variable is required")
			return
		}

		// 密钥必须是 16, 24 或 32 字节
		if len(key) < 32 {
			// 填充到 32 字节
			padded := make([]byte, 32)
			copy(padded, key)
			encryptionKey = padded
		} else {
			encryptionKey = []byte(key[:32])
		}
	})

	return initErr
}
```

**Step 2: 验证 Go 代码编译**

Run: `go build ./...`
Expected: 编译成功，无错误

**Step 3: Commit**

```bash
git add internal/crypto/encryption.go
git commit -m "security: remove hardcoded encryption key, require env variable"
```

---

### Task 2: 完善 .gitignore

**Files:**
- Modify: `.gitignore`

**Step 1: 在 .gitignore 末尾添加敏感文件模式**

在文件末尾添加：

```gitignore
# Secrets
*.pem
*.key
*.crt
*.p12
*.pfx
id_rsa*
*.env
*.env.local
.env.*
```

**Step 2: 验证文件更新**

Run: `cat .gitignore`
Expected: 文件包含新增的敏感文件模式

**Step 3: Commit**

```bash
git add .gitignore
git commit -m "security: add sensitive file patterns to .gitignore"
```

---

### Task 3: 更新 README 添加环境变量说明

**Files:**
- Modify: `README.md`（如果存在）或 `CLAUDE.md`

**Step 1: 检查是否有 README.md**

Run: `ls README.md 2>/dev/null || echo "not found"`

**Step 2: 添加环境变量配置说明**

如果 README.md 存在，在配置章节添加：

```markdown
## 环境变量

| 变量名 | 必需 | 说明 |
|--------|------|------|
| `RUBICK_ENCRYPTION_KEY` | 是 | 用于加密 SSH 密码和私钥，必须是 32 字节字符串 |

示例：
```bash
export RUBICK_ENCRYPTION_KEY="your-32-byte-encryption-key-here"
```
```

如果 README.md 不存在，跳过此步骤。

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add environment variable configuration guide"
```

---

### Task 4: 推送到 GitHub

**Step 1: 查看所有变更**

Run: `git log --oneline -5`
Expected: 显示最近 3 个安全相关的提交

**Step 2: 推送到远程仓库**

Run: `git push origin main`
Expected: 成功推送到 GitHub

---

## Summary

| Task | Description | Files Changed |
|------|-------------|---------------|
| 1 | 移除硬编码加密密钥 | 1 file modified |
| 2 | 完善 .gitignore | 1 file modified |
| 3 | 更新文档 | 1 file modified (optional) |
| 4 | 推送到 GitHub | - |
