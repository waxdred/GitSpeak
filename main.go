package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const ANSWER_SIZE = 40
const PROMPT = "Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n\nDiff:\n"

var INSTRUCTIONS = fmt.Sprintf("\nInstructions for the model:\n-Generate comments explaining why this change was made with a maximum of %d characters.\n-Comments must be concise, clear, and suited to a developer audience.\n-Generate at least three different comments to provide a variety of perspectives on the changes.\n answer format - answer", ANSWER_SIZE)

func getStageFiles() ([]string, error) {
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

func getDiffForFile(filePath string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func runFzf(selection []string) (string, error) {
	input := bytes.NewBufferString(strings.Join(selection, "\n"))
	cmd := exec.Command("fzf")
	var stdout bytes.Buffer
	cmd.Stdin = input      // Définissez le `stdin` de `fzf` pour lire vos réponses
	cmd.Stderr = os.Stderr // Redirigez également le `stderr` pour capturer les erreurs potentielles
	cmd.Stdout = &stdout

	// Démarrez `fzf` et attendez qu'il se termine
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors de l'exécution de fzf: %v\n", err)
		return "", err
	}
	return stdout.String(), nil
}

func GitCommit(commitMessage string) error {
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("erreur lors de l'exécution du commit git : %w", err)
	}
	return nil
}

func main() {
	api_key := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(api_key)
	files, err := getStageFiles()
	if err != nil {
		fmt.Println("Error getting stage files: ", err)
		return
	}
	for _, file := range files {
		diff, err := getDiffForFile(file)
		if err != nil {
			fmt.Println("Error getting diff for file: ", file, err)
			return
		}
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("%s%s%s", PROMPT, diff, INSTRUCTIONS),
					},
				},
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}
		selection := strings.Split(resp.Choices[0].Message.Content, "\n")
		for i, s := range selection {
			selection[i] = strings.TrimPrefix(s, "- ")
		}
		commit, err := runFzf(selection)
		if err != nil {
			fmt.Println("Error running fzf: ", err)
			return
		}
		err = GitCommit(commit[3:])
		if err != nil {
			fmt.Println("Error committing: ", err)
			return
		}
	}
}
