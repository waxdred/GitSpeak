package Models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var MODELLIST = []string{"llama2", "mistral"}

type Ollama struct {
	Model       string
	Url         string
	ModelAnswer Model
	Response    []Model
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

func New(model, url string) *Ollama {
	return &Ollama{
		Model: strings.ToLower(model),
		Url:   fmt.Sprintf("%s/api/generate", url),
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
	return nil
}
