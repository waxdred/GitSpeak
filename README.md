# GitSpeak

GitSpeak is an innovative tool designed to enhance your Git workflow by automatically generating insightful commit messages based on changes made in your code. Leveraging the power of OpenAI's GPT-3.5, GitSpeak analyses your code diffs and crafts descriptive, meaningful commit messages that save you time and improve your project's documentation.

## Features

- Automatic Commit Message Generation: Generate commit messages that describe your code changes and their implications.
- Integration with OpenAI's GPT-3.5: Utilizes the latest AI models for accurate and context-aware message generation.
- Easy Installation: Simple setup process that integrates smoothly with your existing Git workflow.

## Prerequisites

- Git installed on your system.
- Go (Golang) installed on your system (version 1.16 or newer recommended).
- An OpenAI API key. You can obtain one by signing up at [OpenAI](https://platform.openai.com/docs/overview).
- Fzg (Fuzzy Git) installed on your system. [fzf](https://github.com/junegunn/fzf).

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

3.  Add GitSpeak to your PATH.

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

## Usage

With GitSpeak installed and configured, simply run the following command within your Git repository to generate a commit message for staged changes:

```bash
GitSpeak
```

Follow the interactive prompt provided by fzf to select the most appropriate commit message generated based on your code changes.

## Contributing

Contributions to GitSpeak are welcome! Please refer to the CONTRIBUTING.md file for more information on how to contribute to this project.
