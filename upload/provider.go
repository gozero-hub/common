package upload

import (
	"context"
	"time"
)

// UploadProvider 上传服务提供者接口（策略模式）
type UploadProvider interface {
	// GetName 获取提供者名称
	GetName() string

	// GetStsToken 获取STS临时凭证
	GetStsToken(ctx context.Context) (*StsCredentials, error)

	// GetStsTokenWithPolicy 获取带自定义策略的STS凭证
	GetStsTokenWithPolicy(ctx context.Context, policy string) (*StsCredentials, error)

	// UploadFile 上传文件
	UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error)

	// UploadFileWithOptions 上传文件（带更多选项）
	UploadFileWithOptions(ctx context.Context, key string, data []byte, opts UploadOptions) (string, error)

	// DeleteFile 删除文件
	DeleteFile(ctx context.Context, key string) error

	// GetFileUrl 获取文件访问URL
	GetFileUrl(key string) string

	// GetPresignedUrl 获取预签名URL（用于临时访问私有文件）
	GetPresignedUrl(ctx context.Context, key string, expires time.Duration) (string, error)

	// GetPresignedUploadUrl 获取预签名上传URL（用于客户端直传）
	GetPresignedUploadUrl(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)

	// CheckFileExists 检查文件是否存在
	CheckFileExists(ctx context.Context, key string) (bool, error)

	// Close 关闭连接
	Close() error
}

// StsCredentials STS临时凭证信息
type StsCredentials struct {
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
	Expiration      int64  `json:"expiration"`
	Host            string `json:"host"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
}

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType string
	ACL         string
	Metadata    map[string]string
}
