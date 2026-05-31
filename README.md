# Grid

A small POSIX-friendly tool that arranges a set of images into a single PNG
collage.

## Build

```sh
go build -o grid ./cmd
```

## Usage

```
grid [-r rows] [-c cols] [-l mode] [-o name] file...
```

| Flag | Description                                       |
|------|---------------------------------------------------|
| `-r` | number of rows (`0` = auto)                       |
| `-c` | number of columns (`0` = auto)                    |
| `-l` | label mode: `none`, `grid`, `index`, `file`       |
| `-o` | output file prefix (default `merged`)             |
| `-h` | print help                                        |

`--` may be used to terminate option parsing when filenames begin with `-`.

If both `-r` and `-c` are given and their product is smaller than the number of
inputs, the output is split across `<prefix>_1.png`, `<prefix>_2.png`, ….
Otherwise a single `<prefix>.png` is written.

Supported input formats: PNG, JPEG, GIF.

## Configuration

Persistent visual settings are loaded from `$XDG_CONFIG_HOME/grid.toml`, or
`~/.config/grid.toml` if `XDG_CONFIG_HOME` is unset. All keys are optional.

```toml
background       = "#1e1e1e"   # canvas colour, or "transparent"
cell             = "#2d2d2d"   # empty cell / letterbox colour
label_bg         = "#101010"   # label strip background
label_fg         = "#cccccc"   # label text colour
font             = ""          # absolute path to a TTF/OTF (empty = built-in)
font_size        = 13.0        # label text size in points
cell_pad         = 6           # gap between cells in pixels
label_bar_height = 0           # 0 = derived from font_size
```

Hex colours accept `#rrggbb` or `#rrggbbaa`.

## Layout

```
.
├── cmd/                  program entry point
├── internal/collage/     composition, labels, scaling
└── internal/config/      grid.toml loader and colour parsing
```
