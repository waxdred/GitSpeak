package Models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var MODELLIST = []string{"llama2", "mistral"}

type Ollama struct {
	Model       string
	Url         string
	ModelAnswer Model
	Response    []Model
	Commit      []string
}

type Model struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func New(model, url, port string) *Ollama {
	return &Ollama{
		Model: strings.ToLower(model),
		Url:   fmt.Sprintf("%s/api/generate", fmt.Sprintf("%s:%s", url, port)),
	}
}

func (o *Ollama) FormatCommit() {
	var tmp string
	for _, r := range o.Response {
		tmp += r.Response
	}

	o.Commit = strings.Split(tmp, "\n")
	for i, _ := range o.Commit {
		if i == 0 && o.Commit[i][0] == ' ' {
			o.Commit[i] = o.Commit[i][1:]
		}
		regexPattern := `\s*\([^)]*\)$`
		re := regexp.MustCompile(regexPattern)

		o.Commit[i] = re.ReplaceAllString(o.Commit[i], "")
	}
}

func (o *Ollama) Generate(prompt string) error {
	r := Request{
		Model:  o.Model,
		Prompt: prompt,
	}
	rj, err := json.Marshal(r)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(rj)
	client := &http.Client{}
	req, err := http.NewRequest("POST", o.Url, b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var fragments []Model
	decoder := json.NewDecoder(resp.Body)
	for {
		var m Model
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fragments = append(fragments, m)
		if m.Done {
			break
		}
	}
	o.Response = fragments
	o.FormatCommit()
	return nil
}
