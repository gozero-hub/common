package upload

import (
	"context"
	"fmt"
	"sync"

	"resyapi/internal/config"
)

// Manager 上传管理器（工厂模式 + 单例模式）
type Manager struct {
	config    config.UploadConfig
	providers map[config.ProviderType]UploadProvider
	mu        sync.RWMutex
}

// 全局管理器实例
var (
	globalManager *Manager
	once          sync.Once
)

// NewManager 创建上传管理器
func NewManager(cfg config.UploadConfig) (*Manager, error) {
	m := &Manager{
		config:    cfg,
		providers: make(map[config.ProviderType]UploadProvider),
	}

	// 初始化默认的Provider
	if err := m.initProviders(); err != nil {
		return nil, err
	}

	return m, nil
}

// InitGlobalManager 初始化全局管理器
func InitGlobalManager(cfg config.UploadConfig) error {
	var err error
	once.Do(func() {
		globalManager, err = NewManager(cfg)
	})
	return err
}

// GetGlobalManager 获取全局管理器
func GetGlobalManager() *Manager {
	return globalManager
}

// initProviders 初始化所有启用的Provider
func (m *Manager) initProviders() error {
	// 初始化S3
	if m.config.S3.Enabled {
		provider, err := NewS3Provider(m.config.S3)
		if err != nil {
			return fmt.Errorf("failed to init s3 provider: %w", err)
		}
		m.providers[config.ProviderS3] = provider
	}

	// 初始化OSS
	if m.config.OSS.Enabled {
		provider, err := NewOSSProvider(m.config.OSS)
		if err != nil {
			return fmt.Errorf("failed to init oss provider: %w", err)
		}
		m.providers[config.ProviderOSS] = provider
	}

	return nil
}

// GetProvider 获取指定类型的Provider（策略选择）
func (m *Manager) GetProvider(ptype config.ProviderType) (UploadProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[ptype]
	if !ok {
		return nil, fmt.Errorf("provider %s not found or not enabled", ptype)
	}

	return provider, nil
}

// GetDefaultProvider 获取默认的Provider
func (m *Manager) GetDefaultProvider() (UploadProvider, error) {
	return m.GetProvider(m.config.DefaultType)
}

// GetProviderByRequest 根据请求获取Provider
// 如果requestType为空，使用默认类型
func (m *Manager) GetProviderByRequest(requestType config.ProviderType) (UploadProvider, error) {
	if requestType == "" {
		return m.GetDefaultProvider()
	}
	return m.GetProvider(requestType)
}

// RegisterProvider 动态注册Provider
func (m *Manager) RegisterProvider(ptype config.ProviderType, provider UploadProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[ptype] = provider
}

// UnregisterProvider 注销Provider
func (m *Manager) UnregisterProvider(ptype config.ProviderType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if provider, ok := m.providers[ptype]; ok {
		provider.Close()
		delete(m.providers, ptype)
	}
}

// ListProviders 列出所有可用的Provider
func (m *Manager) ListProviders() []config.ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]config.ProviderType, 0, len(m.providers))
	for t := range m.providers {
		types = append(types, t)
	}
	return types
}

// Close 关闭所有Provider
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, provider := range m.providers {
		provider.Close()
	}
	m.providers = make(map[config.ProviderType]UploadProvider)
}

// ============ 便捷方法 ============

// Upload 使用指定Provider上传文件
func (m *Manager) Upload(ctx context.Context, ptype config.ProviderType, key string, data []byte, contentType string) (string, error) {
	provider, err := m.GetProviderByRequest(ptype)
	if err != nil {
		return "", err
	}
	return provider.UploadFile(ctx, key, data, contentType)
}

// Delete 使用指定Provider删除文件
func (m *Manager) Delete(ctx context.Context, ptype config.ProviderType, key string) error {
	provider, err := m.GetProviderByRequest(ptype)
	if err != nil {
		return err
	}
	return provider.DeleteFile(ctx, key)
}

// GetUrl 使用指定Provider获取文件URL
func (m *Manager) GetUrl(ptype config.ProviderType, key string) (string, error) {
	provider, err := m.GetProviderByRequest(ptype)
	if err != nil {
		return "", err
	}
	return provider.GetFileUrl(key), nil
}

// GetStsToken 使用指定Provider获取STS Token
func (m *Manager) GetStsToken(ctx context.Context, ptype config.ProviderType) (*StsCredentials, error) {
	provider, err := m.GetProviderByRequest(ptype)
	if err != nil {
		return nil, err
	}
	return provider.GetStsToken(ctx)
}

// ============ 全局便捷方法 ============

// GUpload 全局上传方法
func GUpload(ctx context.Context, ptype config.ProviderType, key string, data []byte, contentType string) (string, error) {
	if globalManager == nil {
		return "", fmt.Errorf("global manager not initialized")
	}
	return globalManager.Upload(ctx, ptype, key, data, contentType)
}

// GDelete 全局删除方法
func GDelete(ctx context.Context, ptype config.ProviderType, key string) error {
	if globalManager == nil {
		return fmt.Errorf("global manager not initialized")
	}
	return globalManager.Delete(ctx, ptype, key)
}

// GGetUrl 全局获取URL方法
func GGetUrl(ptype config.ProviderType, key string) (string, error) {
	if globalManager == nil {
		return "", fmt.Errorf("global manager not initialized")
	}
	return globalManager.GetUrl(ptype, key)
}

// GGetStsToken 全局获取STS Token方法
func GGetStsToken(ctx context.Context, ptype config.ProviderType) (*StsCredentials, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("global manager not initialized")
	}
	return globalManager.GetStsToken(ctx, ptype)
}
