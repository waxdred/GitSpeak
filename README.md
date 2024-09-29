# GitSpeak

GitSpeak is an innovative tool designed to enhance your Git workflow by automatically generating insightful commit messages based on changes made in your code. Leveraging the power of OpenAI's GPT-3.5, GitSpeak analyses your code diffs and crafts descriptive, meaningful commit messages that save you time and improve your project's documentation.

## Table of Contents

- [Features](#Features)
- [Prerequisites](#Prerequisites)
- [Installation](#Installation)
- [Configuration](#Configuration)
- [Options](#Options)
- [Usages](#Usages)

## Features

- Automatic Commit Message Generation: Generate commit messages that describe your code changes and their implications.
- Integration with OpenAI's GPT-3.5: Utilizes the latest AI models for accurate and context-aware message generation.
- Easy Installation: Simple setup process that integrates smoothly with your existing Git workflow.

![](https://i.imgur.com/W5hUk29.gif)

## Prerequisites

- Git installed on your system.
- Go (Golang) installed on your system (version 1.16 or newer recommended).
- An OpenAI API key. You can obtain one by signing up at [OpenAI](https://platform.openai.com/docs/overview).
- Fzg (Fuzzy Git) installed on your system. [fzf](https://github.com/junegunn/fzf)

## Installation

1. Clone this repository to your local machine.

```bash
git clone https://your-repository-url.git
cd GitSpeak
```

2. Compile and install GitSpeak.

```bash
make
```

This command builds the GitSpeak tool and places the binary in ~/.GitSpeak/bin.

3.  Add GitSpeak to your `PATH`.

```bash
echo 'export PATH=$PATH:~/.GitSpeak/bin' >> ~/.bashrc
source ~/.bashrc
```

## Configuration

To use GitSpeak, you must set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="your_openai_api_key_here"
```

Add this line to your .bashrc, .zshrc, or equivalent shell configuration file to make it permanent.

## Options

GitSpeak supports a number of command-line flag options that allow you to customize its behavior. Below are the available flags and their descriptions:

- `stage`: Commit all files together or one by one (default: true)
  Example usage: ./GitSpeak -stage=false
- `semantic`: List of custom semantic commit terms separated by commas (default:            
 "feat,fix,docs,style,refactor,perf,test,ci,chore,revert")
Example usage: ./GitSpeak -semantic "custom_term,another_term"
- `max_length`: The maximum size of each generated answer (default: 60)
Example usage: ./GitSpeak -max_length=80
- `answer`: The number of answers to generate (default: 4)
Example usage: ./GitSpeak -answer=3
- `Ollama`: Enable or disable Ollama-based commit message generation (default: true)
Example usage: ./GitSpeak -Ollama=false
- `model`: Select the OpenAI model to use for commit message generation (default: "gpt-3.5"). Available models include "gpt-2", "gpt-3", and "gpt-3.5"
Example usage: ./GitSpeak -model gpt-3
- `OllamaUrl`: Specify the URL of the Ollama model to use for commit message generation (default: https://api.openai.com/v1/models/gpt-3-5).
Example usage: ./GitSpeak -OllamaUrl https://api.openai.com/v1/models/gpt-2
- `OllaApiKey`: Specify the API key to use for Ollama-based commit message generation.
Example usage: ./GitSpeak -OllaApiKey "your_ollaki_api_key_here"

These flags provide you with flexibility in how GitSpeak generates commit messages, allowing you to tailor the output to your specific needs or preferences.

## Usages

With GitSpeak installed and configured, simply run the following command within your Git repository to generate a commit message for staged changes

```bash
GitSpeak
```

To customize the behavior of GitSpeak, you can use the flag options as shown in the examples above:

```bash
GitSpeak -max_length=60 -answer=5
```

This command will generate 5 different commit message suggestions, each with a maximum length of 60 characters.

Follow the interactive prompt provided by fzf to select the most appropriate commit message generated based on your code changes.

Customizing Semantic Terms
GitSpeak supports custom semantic commit terms, allowing you to tailor the semantic categories to your project's needs. Use the -semantic flag to specify your custom terms separated by commas:

```bash
./GitSpeak -semantic "feat,fix,docs,customCategory"
```

The default semantic terms are: feat, fix, docs, style, refactor, perf, test, ci, chore, revert. Use the -semantic flag to override these defaults with your preferred terms.

## Contributing

Contributions to GitSpeak are welcome!
