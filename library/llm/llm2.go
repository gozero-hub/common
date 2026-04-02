package llm

//
//import (
//	"context"
//	"gopkg.in/yaml.v3"
//	"sync"
//)
//
//var (
//	once sync.Once
//)
//
//type Provider struct {
//	Name    string `yaml:"name"`
//	BaseURL string `yaml:"baseurl"`
//	Model   string `yaml:"model"`
//	Key     string `yaml:"key"`
//}
//
//type Config struct {
//	LLM struct {
//		Providers     []Provider `yaml:"providers"`
//		FallbackOrder []string   `yaml:"fallback_order"`
//	} `yaml:"llm"`
//}
//
//type callParams struct{ provider string }
//type CallOption func(*callParams)
//
//type LLM interface {
//	Chat(ctx context.Context, prompt string, opts ...CallOption) (string, error)
//	Stream(ctx context.Context, prompt string, onChunk func(chunk string), opts ...CallOption) error
//}
//
//func Init() error {
//	once.Do(func() {
//		var c Config
//		cfgFile := "./llm.yaml"
//		_ = yaml.Unmarshal([]byte(cfgFile), &c)
//	})
//	return nil
//}

//func WithProvider(name string) CallOption { return func(o *callParams) { o.provider = name } }
