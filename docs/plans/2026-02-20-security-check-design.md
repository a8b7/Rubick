# 敏感信息泄露检查与清理设计

## 概述

对已上传到公开 GitHub 仓库的 Rubick 项目进行敏感信息泄露检查，并修复发现的问题。

## 检查范围

- 硬编码密码/密钥/Token
- 加密密钥
- API Key
- 私钥文件
- 证书文件
- 配置文件中的凭据

## 发现的问题与修复

### 问题 1：硬编码默认加密密钥（严重）

**位置**: `internal/crypto/encryption.go:27`

**当前代码**:
```go
key := os.Getenv("RUBICK_ENCRYPTION_KEY")
if key == "" {
    key = "rubick-default-encryption-key-32b" // 硬编码的默认值
}
```

**修复后**:
```go
key := os.Getenv("RUBICK_ENCRYPTION_KEY")
if key == "" {
    log.Fatal("RUBICK_ENCRYPTION_KEY environment variable is required")
}
```

### 问题 2：.gitignore 不完善（中等）

添加以下模式：
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

# Database
data/
*.db
*.sqlite
*.sqlite3

# Logs
*.log
logs/
```

## 已确认安全的内容

- ✅ `config.yaml` 不含敏感信息
- ✅ 无 `.env` 文件被提交
- ✅ 无证书/私钥文件被提交
- ✅ Git 历史中无敏感文件

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/crypto/encryption.go` | 修改 | 移除硬编码密钥，强制环境变量 |
| `.gitignore` | 修改 | 添加敏感文件模式 |
| `README.md` | 修改 | 添加环境变量配置说明 |
