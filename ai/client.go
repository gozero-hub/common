package ai

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

func Client() {
	client := openai.NewClientWithConfig(openai.DefaultConfig("YOUR_KIMI_KEY"))
	//client.BaseURL = "https://api.moonshot.cn/v1"
	ctx := context.Background()

	title := "Book Review: Frederick E. Greenspahn, Ed. The State of American Jewry: New Insights and Scholarship"

	prompt := fmt.Sprintf(`请根据以下论文标题或摘要，判断其:
1. 文章类型（例如：Research Article, Review, Case Report, Book Review 等）
2. 研究关键词（3~5 个）
3. 研究领域（尽量精确到一级或二级学科）

输出 JSON 格式:
{"type": "...", "keywords": ["..."], "research_field": "..."}

论文标题/摘要：
%s`, title)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4o-mini", // 可以换成 kimi 或 deepseek 的模型
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
