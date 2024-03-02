package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

func main() {
	api_key := os.Getenv("OPENAI_API_KEY")
	fmt.Println("API Key: ", api_key)
	files, err := getStageFiles()
	if err != nil {
		fmt.Println("Error getting stage files: ", err)
		return
	}
	fmt.Println("Stage files: ", files)
}
