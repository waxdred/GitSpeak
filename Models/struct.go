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
	"unicode"
)

type Ollama struct {
	Model       string
	Url         string
	Apikey      string
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

func New(model, url, apikey, port string) *Ollama {
	u := fmt.Sprintf("%s:%s", url, port)
	if strings.Contains(u, "https") {
		u = url
	}
	return &Ollama{
		Model:  strings.ToLower(model),
		Apikey: apikey,
		Url:    fmt.Sprintf("%s/api/generate", u),
	}
}

func (o *Ollama) keepCommit() {
	var tmp []string
	re := regexp.MustCompile(`^\d+\. `)
	for i, _ := range o.Commit {
		if len(o.Commit[i]) > 0 && unicode.IsDigit(rune(o.Commit[i][0])) {
			result := re.ReplaceAllString(o.Commit[i], "")
			tmp = append(tmp, result)
		}
	}
	o.Commit = tmp
}

func (o *Ollama) FormatCommit() {
	var tmp string
	for _, r := range o.Response {
		tmp += r.Response
	}

	o.Commit = strings.Split(tmp, "\n")
	if len(o.Commit) > 0 {
		for i, _ := range o.Commit {
			if i == 0 && o.Commit[i][0] == ' ' {
				o.Commit[i] = o.Commit[i][1:]
			}
			regexPattern := `\s*\([^)]*\)$`
			re := regexp.MustCompile(regexPattern)

			o.Commit[i] = re.ReplaceAllString(o.Commit[i], "")
		}
	}
	o.keepCommit()
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
	fmt.Println(o.Url)
	fmt.Println(o.Apikey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.Apikey))
	req.Header.Set("Content-Type", "application/json")
	fmt.Println(req.Header.Get("Authorization"))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error: %s", resp.Status)
	}

	var fragments []Model
	decoder := json.NewDecoder(resp.Body)
	for {
		var m Model
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(m.Response)
		fragments = append(fragments, m)
		if m.Done {
			break
		}
	}
	o.Response = fragments
	o.FormatCommit()
	return nil
}
