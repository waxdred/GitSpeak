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
	"unicode"

	openai "github.com/sashabaranov/go-openai"
)

const PROMPT = "Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n\nDiff:\n"

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
	return &GitCommenter{
		OpenAIKey: apiKey,
		Client:    openai.NewClient(apiKey),
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
	files := strings.Split(out.String(), "\n")
	var cleanedFiles []string
	for _, file := range files {
		if file != "" {
			cleanedFiles = append(cleanedFiles, file)
		}
	}
	return cleanedFiles, nil
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

func (gc *GitCommenter) RunFzf(selection []string) (string, error) {
	customOption := "Customize commit message..."
	selection = append(selection, customOption)
	input := bytes.NewBufferString(strings.Join(selection, "\n"))
	cmd := exec.Command("fzf")
	var stdout bytes.Buffer
	cmd.Stdin = input
	cmd.Stderr = os.Stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running fzf: %v\n", err)
		return "", err
	}
	selected := strings.TrimSpace(stdout.String())
	if selected == customOption {
		fmt.Print("Enter your custom commit message: ")
		reader := bufio.NewReader(os.Stdin)
		customMessage, _ := reader.ReadString('\n')
		return strings.TrimSpace(customMessage), nil
	}

	return selected, nil
}

func (gc *GitCommenter) GitCommit(commitMessage string) error {
	if len(commitMessage) == 0 {
		return fmt.Errorf("commit message is empty")
	}
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing git commit: %w", err)
	}
	return nil
}

func (gc *GitCommenter) ChatGpt(diff string) ([]string, error) {
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
		fmt.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}
	prompt := strings.Split(resp.Choices[0].Message.Content, "\n")
	for i, s := range prompt {
		prompt[i] = strings.TrimPrefix(s, "- ")
		prompt[i] = strings.Replace(prompt[i], "\n", "", -1)
		if unicode.IsDigit(rune(prompt[i][0])) {
			index := strings.Index(prompt[i], " ")
			prompt[i] = prompt[i][index+1:]
		}
	}
	return prompt, nil
}

func (gc *GitCommenter) ProcessCommits() {
	files, err := gc.GetStagedFiles()
	if err != nil {
		fmt.Println("Error getting staged files:", err)
		return
	}
	if len(files) == 0 {
		fmt.Println("No staged files found")
		return
	}
	for _, file := range files {
		diff, err := gc.GetDiffForFile(file)
		if err != nil {
			fmt.Println("Error getting diff for file:", file, err)
			return
		}
		prompt, err := gc.ChatGpt(diff)
		if err != nil {
			fmt.Println("Error running chat gpt:", err)
			return
		}
		commit, err := gc.RunFzf(prompt)
		if err != nil {
			fmt.Println("Error running fzf:", err)
			return
		}
		err = gc.GitCommit(commit)
		if err != nil {
			fmt.Println("Error committing:", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable not set")
		return
	}
	commenter := NewGitCommenter(apiKey)
	commenter.Instructions = fmt.Sprintf("\nInstructions for the model:\n-Generate comments explaining why this change was made with a maximum of %d characters.\n-Comments must be concise, clear, and suited to a developer audience.\n-Generate at least %d different comments to provide a variety of perspectives on the changes.\n answer format - answer", *answerSize, *answer)
	commenter.ProcessCommits()
}
