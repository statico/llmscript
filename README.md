# llmscript

llmscript lets you write shell scripts in natural language.

## Example

```sh
#!/usr/bin/env llmscript

Create an output directory, `output`.
For every PNG file in `input`:
  - Convert it to 256x256 with ImageMagick
  - Run pngcrush
```

Running it would look like this:

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
