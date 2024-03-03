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
	"github.com/waxdred/GitSpeak/Models"
)

const PROMPT = "Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n\nDiff:\n"

var (
	semanticTerms = flag.String("semantic", "feat,fix,docs,style,refactor,perf,test,ci,chore,revert", "List of custom semantic commit terms separated by commas.")
	answerSize    = flag.Int("max_length", 60, "The maximum size of each generated answer.")
	answer        = flag.Int("answer", 4, "The number of answers to generate.")
	ollama        = flag.Bool("Ollama", false, "Run GitSpeak with your models Ollama")
	model         = flag.String("model", "llama2", "The Ollama model, by default llama2.")
	ollamaUrl     = flag.String("OllamaUrl", "http://localhost:11434", "Url of your Ollama server by default http://localhost:11434")
)

type GitCommenter struct {
	OpenAIKey    string
	Client       *openai.Client
	Instructions string
	PathdirGit   string
	Semantic     string
}

func NewGitCommenter(apiKey string) *GitCommenter {
	g := &GitCommenter{
		OpenAIKey: apiKey,
		Client:    openai.NewClient(apiKey),
		Semantic:  strings.Replace(*semanticTerms, ",", ":\n", -1),
	}
	g.PathDirGit()
	return g
}

func (gc *GitCommenter) PathDirGit() error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		fmt.Println("Erreur lors de l'exécution de la commande git:", err)
		return err
	}
	gitDirPath := strings.TrimSpace(out.String())
	gc.PathdirGit = strings.TrimSuffix(gitDirPath, "/.git")
	return nil
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
			cleanedFiles = append(cleanedFiles, fmt.Sprintf("%s/%s", gc.PathdirGit, file))
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

func (gc *GitCommenter) RunFzfSemantic(file string) (string, error) {
	if gc.Semantic == "" {
		return "", nil
	}
	input := bytes.NewBufferString(gc.Semantic)
	cmd := exec.Command("fzf", "--header", "Semantic for "+file)
	var stdout bytes.Buffer
	cmd.Stdin = input
	cmd.Stderr = os.Stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running fzf: %v\n", err)
		return "", err
	}
	selected := strings.TrimSpace(stdout.String())
	if selected == "none:" {
		return "", nil
	}
	return selected, nil
}

func (gc *GitCommenter) RunFzf(selection []string, diff string, file string) (string, error) {
	var err error
	customOption := "Customize commit message..."
	reloadOption := "Reload suggestions..."

	if err != nil {
		fmt.Println("Error running fzf semantic:", err)
		return "", err
	}

	for {
		selection = append(selection, customOption)
		selection = append(selection, reloadOption)
		input := bytes.NewBufferString(strings.Join(selection, "\n"))
		cmd := exec.Command("fzf", "--header", "commit: "+file)
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
		} else if selected == reloadOption {
			selection, err = gc.ChatGpt(diff)
			if err != nil {
				fmt.Println("Error running chat gpt:", err)
				return "", err
			}
			continue
		}
		return selected, nil
	}
}

func (gc *GitCommenter) GitCommit(commitMessage, filePath string) error {
	if len(commitMessage) == 0 {
		return fmt.Errorf("commit message is empty")
	}
	cmd := exec.Command("git", "commit", filePath, "-m", commitMessage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing git commit: %w", err)
	}
	return nil
}

func (gc *GitCommenter) GenPrompt(file string) error {
	var err error
	gc.Semantic, err = gc.RunFzfSemantic(file)
	if err != nil {
		return err
	}
	gc.Instructions = fmt.Sprintf("\nInstructions for the model:\n-Semantic Commit Messages%s\n-Generate comments explaining why this change was made with a maximum of %d characters.\n-Comments must be concise, clear, and suited to a developer audience.\n-Generate at least %d different comments to provide a variety of perspectives on the changes.\n answer format - answer", gc.Semantic, *answerSize, *answer)
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
	var p []string
	for i, s := range prompt {
		prompt[i] = strings.TrimPrefix(s, "- ")
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
		err = gc.GenPrompt(file)
		if err != nil {
			fmt.Println("Error generating prompt:", err)
			return
		}
		if *ollama {
			llm := Models.New(*model, *ollamaUrl)
			llm.Generate(fmt.Sprintf("%s%s%s", PROMPT, diff, gc.Instructions))
			for _, fragment := range llm.Response {
				fmt.Print(fragment.Response)
			}
		} else {
			prompt, err := gc.ChatGpt(diff)
			if err != nil {
				fmt.Println("Error running chat gpt:", err)
				return
			}
			commit, err := gc.RunFzf(prompt, diff, file)
			if err != nil {
				fmt.Println("Error running fzf:", err)
				return
			}
			if gc.Semantic != "" {
				commit = fmt.Sprintf("%s %s", gc.Semantic, commit)
			}
			err = gc.GitCommit(commit, file)
			if err != nil {
				fmt.Println("Error committing:", err)
				return
			}
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
	commenter.ProcessCommits()
}
