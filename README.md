# pgpg-experiments

This is a template package for experimenting with the [Pretty Good Parser Generator](https://github.com/johnkerl/pgpg). One Go module at the repo root depends on `github.com/johnkerl/pgpg/go`; each subdirectory is a separate example app.

## Build

From the repo root:

```bash
go build -o ast-print/ast-print ./ast-print
go build -o json-stream/json-stream ./json-stream
go build -o pascal-s/pascal-s ./pascal-s
go build -o pemdas-eval-gen/tryparse ./pemdas-eval-gen
```

See each subdirectory's `README.md` for how to run and regenerate.

## Updating the pgpg dependency

When new versions of pgpg are tagged (e.g. `go/v0.2.0`), from the repo root:

```bash
go get github.com/johnkerl/pgpg/go@go/v0.2.0
go mod tidy
```

## Using one app as a template

To start your own project from one of these apps: copy the app directory (e.g. `pemdas-eval-gen/`) into a new repo, add a `go.mod` at the root with `module <your-module>` and `require github.com/johnkerl/pgpg/go v0.1.0` (or the latest version), then change that app's imports from `github.com/johnkerl/pgpg-experiments/...` to `<your-module>/generated/...` so the single main package builds under your module.
