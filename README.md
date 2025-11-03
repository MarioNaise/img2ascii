# img2ascii

`img2ascii` is a command-line tool that converts images into ASCII art. It supports various image formats,
and provides options for rendering animated GIFs as ASCII animations.

Supported formats: JPEG, PNG, GIF, BMP, TIFF, and WebP

## Features

- Convert images to ASCII art.
- Support for animated GIFs with frame-by-frame ASCII rendering.
- Support for colored and truecolor output.
- Configurable output dimensions and rendering options.
- Input via file paths or standard input (stdin).

## Installation

### Go

1. ```bash
   go install github.com/MarioNaise/img2ascii@latest
   ```

### Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/MarioNaise/img2ascii.git
   ```
2. Navigate to the project directory:
   ```bash
   cd img2ascii
   ```
3. Build the binary:
   ```bash
   go build -o img2ascii .
   ```

## Usage

### Basic Usage

Convert an image file to ASCII:

```bash
img2ascii path/to/image.jpg
```

### Animated GIFs

Render an animated GIF as ASCII:

```bash
img2ascii path/to/animation.gif
```

### Input from Stdin

Piped image data:

```bash
cat path/to/image.jpg | img2ascii
```

### Help

Display usage information:

```bash
img2ascii --help
```

## Flags

- `-map`: Characters to use for mapping brightness levels (default: `" .-:=+*#%@$"`)
- `-width`: Width of the output in characters
- `-height`: Height of the output in characters
- `-full`: Use full terminal dimensions.
  This is the default, if no other dimension options are provided.
  Defaults to either 'w' or 'h', depending on terminal and image size.
  Overrides `-width` and `-height` (Options: "w", "h", "term")
- `-color`: Enable colored output
- `-truecolor`: Use RGB truecolor for output (requires `-color`)
- `-transparent`: Treat transparent pixels as spaces
- `-bg`: Use background colors instead of foreground colors (requires `-color`)
- `-animate`: Animate GIF images (only supports a single input file)

## Examples

Convert a single image:

```bash
img2ascii example.png
```

Use custom mapping:

```bash
img2ascii -map " .oO@" example.png
```

Enable colored output:

```bash
img2ascii -color example.png
```

Convert multiple files:

```bash
img2ascii path/to/image.jpg path/to/image.png
img2ascii -animate=0 path/to/animations/*.gif
```

Write output to a file:

> Note: width or height needs to be specified when redirecting output to a file.

```bash
img2ascii -width 50 example.png > output.txt
```

More examples:

```bash
img2ascii -map " " -color -transparent -bg example.png
img2ascii -color -transparent -full term animation.gif
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
