# Top-level Makefile for pgpg-experiments. Run from repo root.

APPS := ast-print dkvp-stream json-stream pascal-s pemdas-eval-gen

# Binary paths (app-dir/binary-name)
ast_print_bin       := ast-print/ast-print
dkvp_stream_bin     := dkvp-stream/dkvp-stream
json_stream_bin     := json-stream/json-stream
pascal_s_bin        := pascal-s/pascal-s
pemdas_eval_gen_bin := pemdas-eval-gen/tryparse

BINARIES := \
  $(ast_print_bin) \
  $(dkvp_stream_bin) \
  $(json_stream_bin) \
  $(pascal_s_bin) \
  $(pemdas_eval_gen_bin)

.PHONY: build generate binaries test fmt clean

# Generate lexers/parsers in each app (run each app's generate.sh).
generate:
	$(foreach app,$(APPS),(cd $(app) && ./generate.sh) &&) true

# Build binaries only (assumes generate has been run).
binaries:
	go build -o $(ast_print_bin)       ./ast-print
	go build -o $(dkvp_stream_bin)     ./dkvp-stream
	go build -o $(json_stream_bin)     ./json-stream
	go build -o $(pascal_s_bin)        ./pascal-s
	go build -o $(pemdas_eval_gen_bin) ./pemdas-eval-gen

# Full build: generate then build binaries.
build: generate
	$(MAKE) binaries

# No *_test.go in this repo; go test ./... still runs (may report no tests or build failures).
test:
	go test ./...

# Format Go code in all app directories.
fmt:
	go fmt ./...

# Remove built executables.
clean:
	rm -f $(BINARIES)
