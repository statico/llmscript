# llmscript

[![CI](https://github.com/statico/llmscript/actions/workflows/ci.yml/badge.svg)](https://github.com/statico/llmscript/actions/workflows/ci.yml)

llmscript is a shell script that uses a large language model (LLM) to build and test shell programs so that you can write scripts in natural language instead of bash or other shell scripting languages.

You can configure it to use [Ollama](https://ollama.com/) (free and local), [Claude](https://www.anthropic.com/claude) (paid), or [OpenAI](https://openai.com/) (paid).

> [!NOTE]
> Does this actually work? Yeah, somewhat! Could it create scripts that erase your drive? Maybe! Good luck!
>
> Most of this project was written by [Claude](https://www.anthropic.com/claude) with [Cursor](https://www.cursor.com). I can't actually claim that I "wrote" any of the source code. I barely know Go.

## Example

```
#!/usr/bin/env llmscript

Create an output directory, `output`.
For every PNG file in `input`:
  - Convert it to 256x256 with ImageMagick
  - Run pngcrush
```

Running it with a directory of PNG files would look like this:

```
$ ./convert-pngs
✓ Script generated successfully!
Creating output directory
Convering input/1.png
Convering input/2.png
Convering input/3.png
Running pngcrush on output/1.png
Running pngcrush on output/2.png
Running pngcrush on output/3.png
Done!
```

Running it again will use the cache and not generate any new scripts:

```
$ ./convert-pngs
✓ Cached script found
Creating output directory
Convering input/1.png
...
```

If you want to generate a new script, use the `--no-cache` flag.

## Prerequisites

- [Go](https://go.dev/) (1.22 or later)
- One of:
  - [Ollama](https://ollama.com/) running locally
  - A [Claude](https://www.anthropic.com/claude) API key
  - An [OpenAI](https://openai.com/) API key

## Installation

```
go install github.com/statico/llmscript/cmd/llmscript@latest
```

(Can't find it? Check `~/go/bin`.)

Or, if you're spooked by running LLM-generated shell scripts (good for you!), consider running llmscript via Docker:

```
docker run --network host -it -v "$(pwd):/data" -w /data ghcr.io/statico/llmscript --verbose examples/hello-world
```

## Usage

Create a script file like the above example, or check out the [examples](examples) directory for more. You can use a shebang like:

```
#!/usr/bin/env llmscript
```

or run it directly like:

```
$ llmscript hello-world
```

By default, llmscript will use Ollama with the `llama3.2` model. You can configure this by creating a config file with the `llmscript --write-config` command to create a config file in `~/.config/llmscript/config.yaml` which you can edit. You can also use command-line args (see below).

## How it works

Want to see it all in action? Run `llmscript --verbose examples/hello-world`

Given a script description written in natural language, llmscript works by:

1. Generating a feature script that implements the functionality
2. Generating a test script that verifies the feature script works correctly
3. Running the test script to verify the feature script works correctly, fixing the feature script if necessary, possibly going back to step 1 if the test script fails too many times
4. Caching the scripts for future use
5. Running the feature script with any additional arguments you provide

For example, given a simple hello world script:

```
#!/usr/bin/env llmscript

Print hello world
```

llmscript might generate the following feature script:

```bash
#!/bin/bash
echo "Hello, world!"
```

...and the following test script to test it:

```bash
#!/bin/bash
[ "$(./script.sh)" = "Hello, world!" ] || exit 1
```

## Configuration

llmscript can be configured using a YAML file located at `~/.config/llmscript/config.yaml`. You can auto-generate a configuration file using the `llmscript --write-config` command.

Here's an example configuration:

```yaml
# LLM configuration
llm:
  # The LLM provider to use (required)
  provider: "ollama" # or "claude", "openai", etc.

  # Provider-specific settings
  ollama:
    model: "llama3.2" # The model to use
    host: "http://localhost:11434" # Optional: Ollama host URL

  claude:
    api_key: "${CLAUDE_API_KEY}" # Environment variable reference
    model: "claude-3-opus-20240229"

  openai:
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-4-turbo-preview"

# Maximum number of attempts to fix the script allowed before restarting from step 2
max_fixes: 10

# Maximum number of attempts to generate a working script before giving up completely
max_attempts: 3

# Timeout for script execution during testing (in seconds)
timeout: 30

# Additional prompt to provide to the LLM
additional_prompt: |
  Use ANSI color codes to make the output more readable.
```

### Environment Variables

You can use environment variables in the configuration file using the `${VAR_NAME}` syntax. This is particularly useful for API keys and sensitive information.

### Configuration Precedence

1. Command line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

### Command Line Flags

You can override configuration settings using command line flags:

```shell
llmscript --llm.provider=claude --timeout=10 script.txt
```

## Caveats

> [!WARNING]
> This is an experimental project. It generates and executes shell scripts, which could be dangerous if the LLM generates malicious code. Use at your own risk and always review generated scripts before running them.
