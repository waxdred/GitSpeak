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
- Fzf (Fuzzy Git) installed on your system. [fzf](https://github.com/junegunn/fzf)

## Installation

1. Clone this repository to your local machine.

```bash
git clone git@github.com:waxdred/GitSpeak.git
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

- `max_length`: Sets the maximum size of each generated answer. This controls the length of the commit messages that GitSpeak generates. The default value is 40. Example usage: GitSpeak -answer_size=50
- `answer`: Specifies the number of answers (commit messages) to generate. This allows you to control how many different commit message suggestions GitSpeak will offer you to choose from. The default value is 4. Example usage: GitSpeak -answer=3
- `semantic`: Allows specifying a custom list of semantic commit terms separated by commas. This lets you tailor the semantic categories to your project's needs, enhancing the structure and clarity of your commit history. The default terms are feat, fix, docs, style, refactor, perf, test, ci, chore, revert.

These flags provide you with flexibility in how GitSpeak generates commit messages, allowing you to tailor the output to your specific needs or preferences.

## Usages

With GitSpeak installed and configured, simply run the following command within your Git repository to generate a commit message for staged changes

```bash
GitSpeak
```

To customize the behavior of GitSpeak, you can use the flag options as shown in the examples above:

```bash
GitSpeak -max_length=60 -answer=5

GitSpeak -stage=true # Commit all files together or one by one"
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
