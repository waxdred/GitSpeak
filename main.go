package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	PROMPT          = "Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n\nDiff:\n"
	customizeOption = "Customize commit message..."
	reloadOption    = "Reload suggestions..."
)

var (
	answerSize = flag.Int("max_length", 40, "The maximum size of each generated answer.")
	answer     = flag.Int("answer", 4, "The number of answers to generate.")
)

type GitCommenter struct {
	OpenAIKey    string
	Client       *openai.Client
	Instructions string
}

func NewGitCommenter(apiKey string) *GitCommenter {
	flag.Parse()
	return &GitCommenter{
		OpenAIKey:    apiKey,
		Client:       openai.NewClient(apiKey),
		Instructions: fmt.Sprintf("\nInstructions for the model:\n-Generate comments explaining why this change was made with a maximum of %d characters.\n-Comments must be concise, clear, and suited to a developer audience.\n-Generate at least %d different comments to provide a variety of perspectives on the changes.\n answer format - answer", *answerSize, *answer),
	}
}

func (gc *GitCommenter) GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	return files, nil
}

func (gc *GitCommenter) GetDiffForFile(filePath string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (gc *GitCommenter) GenerateSuggestions(filePath string) ([]string, error) {
	diff, err := gc.GetDiffForFile(filePath)
	if err != nil {
		return nil, err
	}

	resp, err := gc.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("%s%s%s", PROMPT, diff, gc.Instructions),
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	selection := strings.Split(resp.Choices[0].Message.Content, "\n")
	return selection, nil
}

func (gc *GitCommenter) RunFzf(selection []string) (string, error) {
	selection = append(selection, customizeOption, reloadOption)
	input := bytes.NewBufferString(strings.Join(selection, "\n"))

	cmd := exec.Command("fzf")
	cmd.Stdin = input
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (gc *GitCommenter) ProcessCommits() {
	files, err := gc.GetStagedFiles()
	if err != nil {
		fmt.Println("Error getting staged files:", err)
		return
	}

	for _, file := range files {
		for {
			suggestions, err := gc.GenerateSuggestions(file)
			if err != nil {
				fmt.Println("Error generating suggestions:", err)
				return
			}

			choice, err := gc.RunFzf(suggestions)
			if err != nil {
				fmt.Println("Error running fzf:", err)
				return
			}

			if choice == reloadOption {
				continue // Reload suggestions
			} else if choice == customizeOption {
				fmt.Println("Enter your custom commit message:")
				reader := bufio.NewReader(os.Stdin)
				customMessage, _ := reader.ReadString('\n')
				choice = strings.TrimSpace(customMessage)
			}

			if err := gc.GitCommit(choice); err != nil {
				fmt.Println("Error committing:", err)
				return
			}
			break
		}
	}
}

func (gc *GitCommenter) GitCommit(commitMessage string) error {
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing git commit: %w", err)
	}
	return nil
}

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable not set")
		return
	}

	commenter := NewGitCommenter(apiKey)
	commenter.ProcessCommits()
}
