package upload

// ProviderType 上传服务提供商类型
type ProviderType string

const (
	ProviderS3  ProviderType = "s3"  // AWS S3
	ProviderOSS ProviderType = "oss" // 阿里云OSS
)

// UploadConfig 统一的上传配置
type UploadConfig struct {
	// 默认使用的上传类型: s3 或 oss
	DefaultType ProviderType

	// S3 配置
	S3 S3Config

	// OSS 配置
	OSS OSSConfig
}

// S3Config AWS S3配置
type S3Config struct {
	Enabled         bool   // 是否启用
	AccessKeyId     string // Access Key ID
	AccessKeySecret string // Access Key Secret
	Region          string // 区域，如 us-east-1
	Bucket          string // 存储桶名称
	Endpoint        string // 自定义端点（可选，用于兼容其他S3服务）
	RoleArn         string // STS Role ARN
	SessionName     string // STS会话名称
	DurationSeconds int32  // Token有效期（秒），默认3600
	Host            string // 对外访问的Host
}

// OSSConfig 阿里云OSS配置
type OSSConfig struct {
	Enabled         bool   // 是否启用
	AccessKeyId     string // Access Key ID
	AccessKeySecret string // Access Key Secret
	Endpoint        string // OSS端点，如 oss-cn-hangzhou.aliyuncs.com
	Bucket          string // 存储桶名称
	Region          string // 区域，如 cn-hangzhou
	RoleArn         string // RAM Role ARN
	SessionName     string // STS会话名称
	DurationSeconds int    // Token有效期（秒），默认3600
	Host            string // 对外访问的Host（CDN域名）
	UseCname        bool   // 是否使用自定义域名
}
