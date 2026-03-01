# pgpg-experiments

This is a template package for experimenting with the [Pretty Good Parser
Generator](https://github.com/johnkerl/pgpg). One Go module at the repo root depends on
`github.com/johnkerl/pgpg/go`; each subdirectory is a separate example app.

## Build

From the repo root:

```bash
go  build -o ast-print/ast-print      ./ast-print
go  build -o json-stream/json-stream  ./json-stream
go  build -o pascal-s/pascal-s        ./pascal-s
go  build -o pemdas-eval-gen/tryparse ./pemdas-eval-gen
```

See each subdirectory's `README.md` for how to run and regenerate.

## Updating the pgpg dependency

When new versions of PGPG are tagged (e.g. `go/v0.2.0`), from the repo root:

```bash
go get github.com/johnkerl/pgpg/go@v0.2.0
go mod tidy
```

## Using one app as a template

To start your own project from one of these apps: copy the app directory (e.g. `pemdas-eval-gen/`)
into a new repo, add a `go.mod` at the root with `module your-module-name-goes-here`, and `require
github.com/johnkerl/pgpg/go v0.1.0` (or the latest version). Then change that app's imports from
`github.com/johnkerl/pgpg-experiments/...` to `<your-module>/generated/...` so the single main
package builds under your module.

## Co-developing with PGPG

To work on `pgpg-experiments` and `pgpg` in parallel (e.g. developing/tweaking the `pgpg` library
while testing in these apps), use a `go.work` file so the Go toolchain uses your local `pgpg` checkout
instead of the published module.

1. Clone both repos as siblings:
   ```
   git clone https://github.com/johnkerl/pgpg
   git clone https://github.com/johnkerl/pgpg-experiments
   ```
   so you have `pgpg/` and `pgpg-experiments/` next to each other.

2. Create `go.work` in the pgpg-experiments root:
   ```bash
   cd pgpg-experiments
   go work init .
   go work use ../pgpg/go
   ```
   Or create it manually:
   ```
   go 1.25

   use (
       .
       ../pgpg/go
   )
   ```

3. Build and run as usual. The workspace makes `go build` and `go test` use your local pgpg code.

4. Do not check `go.work` or `go.work.sum` in to source control. It's listed in `.gitignore`. Each developer creates their own when needed.

5. To switch back to the published module, remove or rename `go.work`. To temporarily suppress, `export GOWORK=off` for your current terminal session.
