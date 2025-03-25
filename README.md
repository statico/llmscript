# llmscript

llmscript is a shell script that uses a large language model (LLM) to build and test shell programs so that you can write scripts in natural language instead of bash or other shell scripting languages.

You can configure it to use any LLM, such as [Ollama](https://ollama.com/) or [Claude](https://www.anthropic.com/claude).

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

1. Building a testing plan
2. Writing a script and testing it against the plan
3. Making changes to the script until it passes the tests, or restarting from step 2 if there are too many mistakes
4. Checksumming the description and caching the test plan and script for future executions (into `~/.config/llmscript/cache`)
5. Running the script

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