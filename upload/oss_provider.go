package upload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"resyapi/internal/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSProvider 阿里云OSS上传服务提供者
type OSSProvider struct {
	cfg    config.OSSConfig
	client *oss.Client
	bucket *oss.Bucket
}

// NewOSSProvider 创建OSSProvider
func NewOSSProvider(cfg config.OSSConfig) (*OSSProvider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("oss provider is not enabled")
	}

	// 创建OSS客户端
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyId, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create oss client: %w", err)
	}

	// 获取bucket实例
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get oss bucket: %w", err)
	}

	return &OSSProvider{
		cfg:    cfg,
		client: client,
		bucket: bucket,
	}, nil
}

// GetName 获取提供者名称
func (o *OSSProvider) GetName() string {
	return "OSS"
}

// GetStsToken 获取STS临时凭证
func (o *OSSProvider) GetStsToken(ctx context.Context) (*StsCredentials, error) {
	return o.GetStsTokenWithPolicy(ctx, "")
}

// GetStsTokenWithPolicy 获取带自定义策略的STS凭证
func (o *OSSProvider) GetStsTokenWithPolicy(ctx context.Context, policy string) (*StsCredentials, error) {
	duration := o.cfg.DurationSeconds
	if duration == 0 {
		duration = 3600
	}

	// 构建AssumeRole请求
	reqBody := map[string]interface{}{
		"DurationSeconds": duration,
		"RoleArn":         o.cfg.RoleArn,
		"RoleSessionName": o.cfg.SessionName,
	}

	if policy != "" {
		reqBody["Policy"] = policy
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 调用阿里云STS服务
	stsEndpoint := fmt.Sprintf("https://sts.%s.aliyuncs.com/", o.cfg.Region)

	req, err := http.NewRequestWithContext(ctx, "POST", stsEndpoint, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 使用阿里云签名
	return o.getStsTokenViaSDK(ctx, policy)
}

// getStsTokenViaSDK 通过SDK获取STS Token（需要额外引入sts包）
func (o *OSSProvider) getStsTokenViaSDK(ctx context.Context, policy string) (*StsCredentials, error) {
	// 阿里云STS AssumeRole请求URL
	stsEndpoint := fmt.Sprintf("https://sts.%s.aliyuncs.com/", o.cfg.Region)

	// 使用临时方案：构建HTTP请求获取STS Token
	// 实际生产环境建议使用 aliyun-go-sdk-sts
	return o.assumeRole(ctx, stsEndpoint, policy)
}

// assumeRole 调用AssumeRole接口
func (o *OSSProvider) assumeRole(ctx context.Context, endpoint, policy string) (*StsCredentials, error) {
	// 构建请求参数
	params := map[string]string{
		"Action":          "AssumeRole",
		"RoleArn":         o.cfg.RoleArn,
		"RoleSessionName": o.cfg.SessionName,
		"Format":          "JSON",
		"Version":         "2015-04-01",
		"AccessKeyId":     o.cfg.AccessKeyId,
	}

	duration := o.cfg.DurationSeconds
	if duration == 0 {
		duration = 3600
	}
	params["DurationSeconds"] = fmt.Sprintf("%d", duration)

	if policy != "" {
		params["Policy"] = policy
	}

	// 签名请求（简化版，生产环境需要完整签名）
	// 这里使用一个通用的方法获取STS
	// 实际项目建议引入: github.com/aliyun/alibaba-cloud-sdk-go/services/sts

	// 简化实现：返回基本配置信息，实际Token需要通过SDK获取
	// 由于阿里云的签名比较复杂，这里提供一个框架实现

	// 临时返回（实际项目需要完整实现STS调用）
	return &StsCredentials{
		AccessKeyId:     o.cfg.AccessKeyId,
		AccessKeySecret: o.cfg.AccessKeySecret,
		SecurityToken:   "", // 需要通过STS服务获取
		Expiration:      time.Now().Add(time.Duration(duration) * time.Second).Unix(),
		Host:            o.getHost(),
		Region:          o.cfg.Region,
		Bucket:          o.cfg.Bucket,
		Endpoint:        o.cfg.Endpoint,
	}, fmt.Errorf("please implement STS call with alibaba-cloud-sdk-go/services/sts")
}

// UploadFile 上传文件
func (o *OSSProvider) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	opts := []oss.Option{
		oss.ContentType(contentType),
	}

	err := o.bucket.PutObject(key, bytes.NewReader(data), opts...)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to oss: %w", err)
	}

	return o.GetFileUrl(key), nil
}

// UploadFileWithOptions 上传文件（带更多选项）
func (o *OSSProvider) UploadFileWithOptions(ctx context.Context, key string, data []byte, opts UploadOptions) (string, error) {
	var ossOpts []oss.Option

	if opts.ContentType != "" {
		ossOpts = append(ossOpts, oss.ContentType(opts.ContentType))
	}

	if opts.ACL != "" {
		// OSS ACL映射
		acl := oss.ACLPrivate
		switch opts.ACL {
		case "public-read":
			acl = oss.ACLPublicRead
		case "public-read-write":
			acl = oss.ACLPublicReadWrite
		}
		ossOpts = append(ossOpts, oss.ACL(acl))
	}

	if len(opts.Metadata) > 0 {
		for k, v := range opts.Metadata {
			ossOpts = append(ossOpts, oss.Meta(k, v))
		}
	}

	err := o.bucket.PutObject(key, bytes.NewReader(data), ossOpts...)
	if err != nil {
		return "", fmt.Errorf("failed to upload file with options: %w", err)
	}

	return o.GetFileUrl(key), nil
}

// DeleteFile 删除文件
func (o *OSSProvider) DeleteFile(ctx context.Context, key string) error {
	err := o.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("failed to delete oss file: %w", err)
	}
	return nil
}

// GetFileUrl 获取文件访问URL
func (o *OSSProvider) GetFileUrl(key string) string {
	host := o.getHost()
	if host != "" {
		return fmt.Sprintf("%s/%s", host, key)
	}
	// 使用默认的OSS访问URL
	return fmt.Sprintf("https://%s.%s/%s", o.cfg.Bucket, o.cfg.Endpoint, key)
}

// getHost 获取Host
func (o *OSSProvider) getHost() string {
	if o.cfg.Host != "" {
		return o.cfg.Host
	}
	if o.cfg.UseCname {
		return fmt.Sprintf("https://%s", o.cfg.Endpoint)
	}
	return ""
}

// GetPresignedUrl 获取预签名URL（用于临时访问私有文件）
func (o *OSSProvider) GetPresignedUrl(ctx context.Context, key string, expires time.Duration) (string, error) {
	// OSS生成签名URL
	signedURL, err := o.bucket.SignURL(key, oss.HTTPGet, int64(expires.Seconds()))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}
	return signedURL, nil
}

// GetPresignedUploadUrl 获取预签名上传URL（用于客户端直传）
func (o *OSSProvider) GetPresignedUploadUrl(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
	// OSS生成签名URL用于上传
	options := []oss.Option{
		oss.ContentType(contentType),
	}
	signedURL, err := o.bucket.SignURL(key, oss.HTTPPut, int64(expires.Seconds()), options...)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload url: %w", err)
	}
	return signedURL, nil
}

// CheckFileExists 检查文件是否存在
func (o *OSSProvider) CheckFileExists(ctx context.Context, key string) (bool, error) {
	exist, err := o.bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return exist, nil
}

// Close 关闭连接
func (o *OSSProvider) Close() error {
	// OSS客户端不需要显式关闭
	return nil
}

// 确保 OSSProvider 实现了 UploadProvider 接口
var _ UploadProvider = (*OSSProvider)(nil)
