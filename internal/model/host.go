package model

import (
	"rubick/internal/crypto"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Host Docker 主机配置
type Host struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Type        string    `gorm:"not null" json:"type"` // local, tcp, ssh
	Host        string    `json:"host"`                 // 远程主机地址
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// SSH 配置
	SSHUser       string `gorm:"column:ssh_user" json:"ssh_user,omitempty"`
	SSHAuthType   string `gorm:"column:ssh_auth_type" json:"ssh_auth_type,omitempty"` // key, password
	SSHPrivateKey string `gorm:"column:ssh_private_key" json:"ssh_private_key,omitempty"`
	SSHPassword   string `gorm:"column:ssh_password" json:"ssh_password,omitempty"`
	SSHPort       int    `gorm:"column:ssh_port;default:22" json:"ssh_port,omitempty"`

	// TLS 配置
	TLSCertID     string `gorm:"column:tls_cert_id" json:"tls_cert_id,omitempty"`
	SkipTLSVerify bool   `gorm:"column:skip_tls_verify;default:false" json:"skip_tls_verify"`

	// Docker 端口
	DockerPort int `gorm:"column:docker_port;default:2375" json:"docker_port,omitempty"`

	// 未加密的临时字段（用于内部使用）
	sshPrivateKeyPlain string
	sshPasswordPlain   string
}

// ClearSensitiveFields 清除敏感字段（用于 API 响应）
func (h *Host) ClearSensitiveFields() {
	h.SSHPassword = ""
	h.SSHPrivateKey = ""
}

// BeforeCreate 创建前钩子
func (h *Host) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	return h.encryptSensitiveFields()
}

// BeforeUpdate 更新前钩子
func (h *Host) BeforeUpdate(tx *gorm.DB) error {
	return h.encryptSensitiveFields()
}

// AfterFind 查询后钩子
func (h *Host) AfterFind(tx *gorm.DB) error {
	return h.decryptSensitiveFields()
}

// encryptSensitiveFields 加密敏感字段
func (h *Host) encryptSensitiveFields() error {
	var err error

	if h.SSHPassword != "" && h.sshPasswordPlain == "" {
		h.SSHPassword, err = crypto.Encrypt(h.SSHPassword)
		if err != nil {
			return err
		}
	}

	if h.SSHPrivateKey != "" && h.sshPrivateKeyPlain == "" {
		h.SSHPrivateKey, err = crypto.Encrypt(h.SSHPrivateKey)
		if err != nil {
			return err
		}
	}

	return nil
}

// decryptSensitiveFields 解密敏感字段
func (h *Host) decryptSensitiveFields() error {
	var err error

	if h.SSHPassword != "" {
		h.sshPasswordPlain, err = crypto.Decrypt(h.SSHPassword)
		if err != nil {
			return err
		}
		h.SSHPassword = h.sshPasswordPlain
	}

	if h.SSHPrivateKey != "" {
		h.sshPrivateKeyPlain, err = crypto.Decrypt(h.SSHPrivateKey)
		if err != nil {
			return err
		}
		h.SSHPrivateKey = h.sshPrivateKeyPlain
	}

	return nil
}

// Certificate TLS 证书
type Certificate struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	CACert      string    `gorm:"column:ca_cert" json:"-"`     // 加密存储
	ClientCert  string    `gorm:"column:client_cert" json:"-"` // 加密存储
	ClientKey   string    `gorm:"column:client_key" json:"-"`  // 加密存储
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// BeforeCreate 创建前钩子
func (c *Certificate) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ComposeProject Compose 项目
type ComposeProject struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	HostID    string    `gorm:"not null;index" json:"host_id"`

	// 源类型: content (直接内容) 或 directory (指定目录)
	SourceType string `gorm:"not null;default:'content'" json:"source_type"`

	// Content 模式字段
	Content string `gorm:"type:text" json:"content,omitempty"` // YAML 内容

	// Directory 模式字段
	WorkDir     string `json:"work_dir,omitempty"`      // 服务器上的目录路径
	ComposeFile string `gorm:"default:'docker-compose.yml'" json:"compose_file,omitempty"`
	EnvFile     string `json:"env_file,omitempty"` // 可选环境变量文件

	Status    string    `gorm:"default:'stopped'" json:"status"` // stopped, running, error
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Host *Host `gorm:"foreignKey:HostID" json:"host,omitempty"`
}

// BeforeCreate 创建前钩子
func (p *ComposeProject) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
