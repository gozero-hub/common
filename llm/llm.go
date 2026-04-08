package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// 配置写死在一个 map，key 即名称
var provMap = map[string]provider{
	"openai": {
		Name:    "openai",
		BaseURL: "https://api.openai.com/v1/chat/completions ",
		Model:   "gpt-3.5-turbo",
		Key:     os.Getenv("OPENAI_KEY"),
	},
	"kimi": {
		Name:    "kimi",
		BaseURL: "https://api.moonshot.cn/v1/chat/completions ",
		Model:   "moonshot-v1-128k",
		Key:     "",
	},
	//"deepseek": {
	//	Name:    "deepseek",
	//	BaseURL: "https://api.deepseek.com/v1/chat/completions ",
	//	Model:   "deepseek-chat",
	//	Key:     os.Getenv("DEEPSEEK_KEY"),
	//},
	"deepseek": {
		Name:    "deepseek",
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode ",
		Model:   "deepseek-v3",
		Key:     "",
	},
}

// 降级顺序
var fallbackOrder = []string{"deepseek", "openai", "kimi"}

type provider struct {
	Name    string
	BaseURL string
	Model   string
	Key     string
}

// 请求结构体（兼容 OpenAI / Kimi / DeepSeek）
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []messageItem `json:"messages"`
	Stream   bool          `json:"stream"`
}

type messageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CallOption 留空，后续可扩展
type CallOption func(*callParams)
type callParams struct{}

// ==================== 对外 API ====================

// Chat 一次性聊天，失败按 fallbackOrder 降级
func Chat(ctx context.Context, prompt string, _ ...CallOption) (string, error) {
	for _, name := range fallbackOrder {
		p, ok := provMap[name]
		if !ok || p.Key == "" {
			continue
		}
		text, err := callOnce(ctx, p, prompt)
		if err == nil {
			return text, nil
		}
		// 日志可在这里打印
	}
	return "", errors.New("llmext: all providers failed")
}

// Stream 流式聊天，失败自动降级
func Stream(ctx context.Context, prompt string, onChunk func(chunk string), _ ...CallOption) error {
	for _, name := range fallbackOrder {
		p, ok := provMap[name]
		if !ok || p.Key == "" {
			continue
		}
		err := streamOnce(ctx, p, prompt, onChunk)
		if err == nil {
			return nil
		}
	}
	return errors.New("llmext: all providers failed")
}

func getChatRequest(p provider, prompt string, stream bool) (io.Reader, error) {
	reqBody := chatRequest{
		Model: p.Model,
		Messages: []messageItem{{
			Role:    "user",
			Content: prompt,
		}},
		Stream: stream,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(bodyBytes), nil
}

// ==================== 底层 HTTP ====================

func callOnce(ctx context.Context, p provider, prompt string) (string, error) {
	body, err := getChatRequest(p, prompt, false)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.Key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status=%d", resp.StatusCode)
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", errors.New("empty choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func streamOnce(ctx context.Context, p provider, prompt string, onChunk func(chunk string)) error {
	body, err := getChatRequest(p, prompt, true)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.Key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		var delta struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(raw, &delta); err != nil {
			continue
		}
		if len(delta.Choices) > 0 {
			if chunk := delta.Choices[0].Delta.Content; chunk != "" {
				onChunk(chunk)
			}
		}
	}
	return nil
}
