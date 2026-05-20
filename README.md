# Grid

tool to create image collage



## Setup

### Requirements

* [Go](https://go.dev/dl/) 1.22+


### Install dependencies

```bash
$ go mod download
```

### Build

```bash
$ make build
```

### Run

```bash
$ make run
```




## Testing

### Run all tests

```bash
$ go test ./...
```



## Documentation

* [Project Structure](docs/project_structure.md): repository layout and module 
  responsibilities




## Release

Add details on how a release is produced. Typically done in CI, therefore
descibe process by which a release build is triggered, alternatively if no CI,
describe manual steps to build and release.



## Copier Update

The project uses [Copier](https://copier.readthedocs.io/en/stable/) to manage
its template. When the template is updated, the project can be easily brought
up to date by running the following command from the project root:
```bash
$ copier update --defaults --answers-file .copier-go-answers.yml
```
When the `--defaults` option is used, copier update skips all previously
answered questions and prompts only for newly introduced ones.
