package openai

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	OpenAIKey string
	Client    *openai.Client
}

func NewOpenAI(apiKey string) *OpenAI {
	if apiKey == "" {
		return nil
	}
	return &OpenAI{
		OpenAIKey: apiKey,
		Client:    openai.NewClient(apiKey),
	}
}

func (o *OpenAI) ChatGpt(diff, Sendprompt, Instructions string) ([]string, error) {
	resp, err := o.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("%s%s%s", Sendprompt, diff, Instructions),
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}
	prompt := strings.Split(resp.Choices[0].Message.Content, "\n")
	var p []string
	for i, _ := range prompt {
		prompt[i] = strings.TrimPrefix(prompt[i], "- ")
		prompt[i] = strings.TrimLeft(prompt[i], " \t")
		if len(prompt[i]) > 0 && unicode.IsDigit(rune(prompt[i][0])) {
			index := strings.Index(prompt[i], " ")
			if index != -1 {
				prompt[i] = prompt[i][index+1:]
			}
		}
		index := strings.Index(prompt[i], ".")
		if index != -1 {
			prompt[i] = prompt[i][:index]
		}
		prompt[i] = strings.Replace(prompt[i], "\n", "", -1)
		if prompt[i] != "" {
			p = append(p, prompt[i])
		}
	}
	return p, nil
}
