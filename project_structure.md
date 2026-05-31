# Project Structure

## Top-Level Layout

```
.
├── cmd/
│   ├── main.go            program entry, flag parsing, run orchestration
│   └── io.go              image decoding, PNG writing, multi-page output
├── internal/
│   ├── collage/
│   │   ├── collage.go     Make composes a grid onto a single canvas
│   │   ├── dims.go        SquareDimensions chooses an auto layout
│   │   ├── labels.go      LabelMode, label text, and label rendering
│   │   ├── render.go      maxDims and aspect-preserving scaling helpers
│   │   └── style.go       Style struct, colour and font resolution
│   └── config/
│       └── config.go      loader for ~/.config/grid.toml and colour parser
├── go.mod
├── README.md
└── LICENSE
```

## Module responsibilities

- **cmd**: thin layer that parses POSIX-style short flags, loads the config,
  decodes input images, and dispatches to the collage package. No image logic
  lives here.
- **internal/collage**: pure rendering. Given a `Style`, a label mode, and a
  list of images, produces an `*image.RGBA`. Knows nothing about flags or
  filesystem layout.
- **internal/config**: TOML loader plus hex-colour parser. Returns sensible
  defaults when the config file is absent.
