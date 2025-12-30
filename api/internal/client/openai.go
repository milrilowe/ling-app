package client

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// openaiClient implements OpenAIClient using the OpenAI API.
type openaiClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a new OpenAI client.
func NewOpenAIClient(apiKey string) OpenAIClient {
	client := openai.NewClient(apiKey)
	return &openaiClient{
		client: client,
	}
}

// Generate calls OpenAI to generate an AI response.
func (c *openaiClient) Generate(messages []ConversationMessage) (string, error) {
	// Convert our message format to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: openaiMessages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateTitle generates a short title (3-5 words) from conversation content.
func (c *openaiClient) GenerateTitle(content string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "Generate a very short title (3-5 words) that describes the topic of this conversation. Return only the title, no quotes or punctuation.",
				},
				{
					Role:    "user",
					Content: content,
				},
			},
			MaxTokens: 20,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate title: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
