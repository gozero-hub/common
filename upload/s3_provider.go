package upload

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"resyapi/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// S3Provider S3上传服务提供者
type S3Provider struct {
	cfg       config.S3Config
	client    *s3.Client
	stsClient *sts.Client
}

// NewS3Provider 创建S3Provider
func NewS3Provider(cfg config.S3Config) (*S3Provider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("s3 provider is not enabled")
	}

	client, err := createS3Client(cfg)
	if err != nil {
		return nil, err
	}

	stsClient, err := createStsClient(cfg)
	if err != nil {
		return nil, err
	}

	return &S3Provider{
		cfg:       cfg,
		client:    client,
		stsClient: stsClient,
	}, nil
}

// createS3Client 创建S3客户端
func createS3Client(cfg config.S3Config) (*s3.Client, error) {
	var optFns []func(*awsConfig.LoadOptions) error

	if cfg.AccessKeyId != "" && cfg.AccessKeySecret != "" {
		optFns = append(optFns, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyId, cfg.AccessKeySecret, ""),
		))
	}

	if cfg.Region != "" {
		optFns = append(optFns, awsConfig.WithRegion(cfg.Region))
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = true
		},
	}

	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	return s3.NewFromConfig(awsCfg, s3Opts...), nil
}

// createStsClient 创建STS客户端
func createStsClient(cfg config.S3Config) (*sts.Client, error) {
	var optFns []func(*awsConfig.LoadOptions) error

	if cfg.AccessKeyId != "" && cfg.AccessKeySecret != "" {
		optFns = append(optFns, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyId, cfg.AccessKeySecret, ""),
		))
	}

	if cfg.Region != "" {
		optFns = append(optFns, awsConfig.WithRegion(cfg.Region))
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config for sts: %w", err)
	}

	return sts.NewFromConfig(awsCfg), nil
}

// GetName 获取提供者名称
func (s *S3Provider) GetName() string {
	return "S3"
}

// GetStsToken 获取STS临时凭证
func (s *S3Provider) GetStsToken(ctx context.Context) (*StsCredentials, error) {
	duration := s.cfg.DurationSeconds
	if duration == 0 {
		duration = 3600
	}

	// input := &sts.AssumeRoleInput{
	// 	// RoleArn:         aws.String(s.cfg.RoleArn),
	// 	// RoleSessionName: aws.String(s.cfg.SessionName),
	// 	DurationSeconds: aws.Int32(duration),
	// }

	// result, err := s.stsClient.AssumeRole(ctx, nil)
	result, err := s.stsClient.GetSessionToken(ctx, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(3600),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assume role: %w", err)
	}

	return &StsCredentials{
		AccessKeyId:     aws.ToString(result.Credentials.AccessKeyId),
		AccessKeySecret: aws.ToString(result.Credentials.SecretAccessKey),
		SecurityToken:   aws.ToString(result.Credentials.SessionToken),
		Expiration:      result.Credentials.Expiration.Unix(),
		Host:            s.cfg.Host,
		Region:          s.cfg.Region,
		Bucket:          s.cfg.Bucket,
		Endpoint:        s.cfg.Endpoint,
	}, nil
}

// GetStsTokenWithPolicy 获取带自定义策略的STS凭证
func (s *S3Provider) GetStsTokenWithPolicy(ctx context.Context, policy string) (*StsCredentials, error) {
	duration := s.cfg.DurationSeconds
	if duration == 0 {
		duration = 3600
	}

	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(s.cfg.RoleArn),
		RoleSessionName: aws.String(s.cfg.SessionName),
		DurationSeconds: aws.Int32(duration),
		Policy:          aws.String(policy),
	}

	result, err := s.stsClient.AssumeRole(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role with policy: %w", err)
	}

	return &StsCredentials{
		AccessKeyId:     aws.ToString(result.Credentials.AccessKeyId),
		AccessKeySecret: aws.ToString(result.Credentials.SecretAccessKey),
		SecurityToken:   aws.ToString(result.Credentials.SessionToken),
		Expiration:      result.Credentials.Expiration.Unix(),
		Host:            s.cfg.Host,
		Region:          s.cfg.Region,
		Bucket:          s.cfg.Bucket,
		Endpoint:        s.cfg.Endpoint,
	}, nil
}

// UploadFile 上传文件
func (s *S3Provider) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return s.GetFileUrl(key), nil
}

// UploadFileWithOptions 上传文件（带更多选项）
func (s *S3Provider) UploadFileWithOptions(ctx context.Context, key string, data []byte, opts UploadOptions) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	if opts.ACL != "" {
		input.ACL = s3Types.ObjectCannedACL(opts.ACL)
	}

	if len(opts.Metadata) > 0 {
		input.Metadata = opts.Metadata
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file with options: %w", err)
	}

	return s.GetFileUrl(key), nil
}

// DeleteFile 删除文件
func (s *S3Provider) DeleteFile(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileUrl 获取文件访问URL
func (s *S3Provider) GetFileUrl(key string) string {
	if s.cfg.Host != "" {
		return fmt.Sprintf("%s/%s/%s", s.cfg.Host, s.cfg.Bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.Bucket, s.cfg.Region, key)
}

// GetPresignedUrl 获取预签名URL
func (s *S3Provider) GetPresignedUrl(ctx context.Context, key string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}

	result, err := presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}

	return result.URL, nil
}

// GetPresignedUploadUrl 获取预签名上传URL
func (s *S3Provider) GetPresignedUploadUrl(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	result, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload url: %w", err)
	}

	return result.URL, nil
}

// CheckFileExists 检查文件是否存在
func (s *S3Provider) CheckFileExists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Close 关闭连接
func (s *S3Provider) Close() error {
	// AWS SDK 不需要显式关闭
	return nil
}

// 确保 S3Provider 实现了 UploadProvider 接口
var _ UploadProvider = (*S3Provider)(nil)
