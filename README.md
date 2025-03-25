# llmscript

llmscript is a shell script that uses a large language model (LLM) to build and test shell programs so that you can write scripts in natural language instead of bash or other shell scripting languages.

You can configure it to use any LLM, such as [Ollama](https://ollama.com/) or [Claude](https://www.anthropic.com/claude).

> [!NOTE]
> Most of this project was written by an LLM, so it's not perfect. I can't actually claim that I "wrote" any of the source code.

## Example

```sh
#!/usr/bin/env llmscript

Create an output directory, `output`.
For every PNG file in `input`:
  - Convert it to 256x256 with ImageMagick
  - Run pngcrush
```

Running it with a directory of PNG files would look like this:

```shell
$ ./convert-pngs
Creating output directory
Convering input/1.png
Convering input/2.png
Convering input/3.png
Running pngcrush on output/1.png
Running pngcrush on output/2.png
Running pngcrush on output/3.png
Done!
```

<details>
<summary>Show intermediate steps</summary>

# TODO

</details>

## How it works

Given a script description written in natural language, llmscript works by:

1. Generating two scripts:
   - The main script that does the actual work
   - A test script that verifies the main script works correctly
2. Making both scripts executable
3. Running the test script, which will:
   - Set up the test environment
   - Run the main script
   - Verify the output and state
   - Exit with success/failure status
4. If the test fails, the LLM will fix both scripts and try again
5. Once tests pass, the scripts are cached for future use (in `~/.config/llmscript/cache`)

For example, if you write "Print hello world", llmscript might generate:

```sh
# script.sh
MESSAGE=${MESSAGE:-"Hello, world!"}
echo $MESSAGE

# test.sh
#!/bin/sh
set -e
[ "$(./script.sh)" = "Hello, world!" ] || exit 1
```

## Configuration

llmscript can be configured using a YAML file located at `~/.config/llmscript/config.yaml`. You can auto-generate a configuration file using the `llmscript --write-config` command.

Here's an example configuration:

```yaml
# LLM configuration
llm:
  # The LLM provider to use (required)
  provider: "ollama"  # or "claude", "openai", etc.

  # Provider-specific settings
  ollama:
    model: "llama3.2"  # The model to use
    host: "http://localhost:11434"  # Optional: Ollama host URL

  claude:
    api_key: "${CLAUDE_API_KEY}"  # Environment variable reference
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