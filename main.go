package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/waxdred/GitSpeak/Models"
	openai "github.com/waxdred/GitSpeak/Openai"
)

const PROMPT = "Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n\nDiff:\n"

var (
	semanticTerms = flag.String("semantic", "feat,fix,docs,style,refactor,perf,test,ci,chore,revert", "List of custom semantic commit terms separated by commas.")
	answerSize    = flag.Int("max_length", 60, "The maximum size of each generated answer.")
	answer        = flag.Int("answer", 4, "The number of answers to generate.")
	ollama        = flag.Bool("Ollama", false, "Run GitSpeak with your models Ollama")
	model         = flag.String("model", "llama2", "The Ollama model, by default llama2.")
	ollamaUrl     = flag.String("OllamaUrl", "http://localhost", "Url and port of your Ollama server by default http://localhost")
	port          = flag.String("port", "11434", "Port of your Ollama server by default 11434")
	comitAll      = flag.Bool("all", true, "Commit all files together or one by one")
)

type GitCommenter struct {
	OpenAi         *openai.OpenAI
	Instructions   string
	PathdirGit     string
	Semantic       string
	SemanticSelect string
	commitAll      bool
}

func NewGitCommenter(apiKey string, commitAll bool) *GitCommenter {
	g := &GitCommenter{
		Semantic:  strings.Replace(*semanticTerms, ",", ":\n", -1),
		OpenAi:    openai.NewOpenAI(apiKey),
		commitAll: commitAll,
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
	var cmd *exec.Cmd
	if gc.commitAll {
		cmd = exec.Command("git", "diff", "--cached")
	} else {
		cmd = exec.Command("git", "diff", "--cached", filePath)
	}
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
	input := bytes.NewBufferString(gc.Semantic + "\n ")
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
			selection, err = gc.OpenAi.ChatGpt(diff, PROMPT, gc.Instructions)
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
	var cmd *exec.Cmd
	if len(commitMessage) == 0 {
		return fmt.Errorf("commit message is empty")
	}
	if filePath == "all files" {
		cmd = exec.Command("git", "commit", "-a", "-m", commitMessage)
	} else {
		cmd = exec.Command("git", "commit", filePath, "-m", commitMessage)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing git commit: %w", err)
	}
	fmt.Println("Commit created:", commitMessage)

	return nil
}

func (gc *GitCommenter) GenPrompt(file string) error {
	var err error
	gc.SemanticSelect, err = gc.RunFzfSemantic(file)
	if err != nil {
		return err
	}
	if gc.SemanticSelect == "" || gc.SemanticSelect == " " {
		gc.Instructions = fmt.Sprintf("Instructions for the model:\n-Comments must be concise, clear, and suited to a developer audience.\n-Generate at least %d different comments to provide a variety of perspectives on the changes with a maximum of %d characters.\n-formating answer in the following way only:\n1. <Commit message>\n-Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n", *answer, *answerSize)
	} else {
		gc.Instructions = fmt.Sprintf("Instructions for the model:\n-Semantic Commit Messages %s\n-Comments must be concise, clear, and suited to a developer audience.\n-Generate at least %d different comments to provide a variety of perspectives on the changes with a maximum of %d characters.\n-formating answer in the following way only:\n1. <Commit message>\n-Based on the following diff, generate several informative commit comments that explain the changes made and their potential impact on the system. The changes are as follows\n", gc.SemanticSelect, *answer, *answerSize)
	}
	return nil
}

func (gc *GitCommenter) RunCommit(file string) {
	var prompt []string
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
		llm := Models.New(*model, *ollamaUrl, *port)
		err := llm.Generate(fmt.Sprintf("%s%s", gc.Instructions, diff))
		if err != nil {
			fmt.Println("Error generating comments:", err)
			return
		}
		prompt = llm.Commit
	} else {
		prompt, err = gc.OpenAi.ChatGpt(diff, PROMPT, gc.Instructions)
		if err != nil {
			fmt.Println("Error running chat gpt:", err)
			return
		}
	}
	commit, err := gc.RunFzf(prompt, diff, file)
	if err != nil {
		fmt.Println("Error running fzf:", err)
		return
	}
	if gc.SemanticSelect != "" || gc.SemanticSelect != " " {
		commit = fmt.Sprintf("%s %s", gc.SemanticSelect, commit)
	}
	err = gc.GitCommit(commit, file)
	if err != nil {
		fmt.Println("Error committing:", err)
		return
	}
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
	if gc.commitAll {
		gc.RunCommit("all files")
	}
	for _, file := range files {
		gc.RunCommit(file)
	}
}

func main() {
	flag.Parse()
	apiKey := os.Getenv("OPENAI_API_KEY")
	commenter := NewGitCommenter(apiKey, *comitAll)
	commenter.ProcessCommits()
}
