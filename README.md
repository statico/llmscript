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