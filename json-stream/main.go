package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/johnkerl/pgpg-experiments/json-stream/generated/lexers"
	"github.com/johnkerl/pgpg-experiments/json-stream/generated/parsers"
)

type traceOptions struct {
	tokens   bool
	states   bool
	stack    bool
	printAst bool
	astMode  string // "", "noast", or "fullast"
}

// lineBufReader implements io.Reader and delivers stdin one line at a time so the
// lexer sees input as soon as the user presses Enter (or the pipe sends a newline).
type lineBufReader struct {
	br  *bufio.Reader
	buf []byte
}

func (l *lineBufReader) Read(p []byte) (n int, err error) {
	for len(p) > 0 {
		if len(l.buf) > 0 {
			n = copy(p, l.buf)
			l.buf = l.buf[n:]
			return n, nil
		}
		l.buf, err = l.br.ReadBytes('\n')
		if len(l.buf) > 0 {
			continue
		}
		return 0, err
	}
	return 0, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [-e] [-multi] [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -e: arguments are expressions to parse (at least one required).\n")
	fmt.Fprintf(os.Stderr, "  -multi: parse multiple JSON values from one stream (one per line).\n")
	fmt.Fprintf(os.Stderr, "  Without -e: zero args = read from stdin; one or more = read from those files.\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var printAst bool
	var exprMode bool
	var multi bool
	var noast bool
	var fullast bool
	var traceTokens bool
	var traceStates bool
	var traceStack bool
	flag.BoolVar(&printAst, "v", false, "Print AST")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.BoolVar(&multi, "multi", false, "Parse multiple top-level objects from one stream")
	flag.BoolVar(&noast, "noast", false, "Syntax-only: do not build or print AST")
	flag.BoolVar(&fullast, "fullast", false, "Ignore AST hints and build full parse tree")
	flag.BoolVar(&traceTokens, "tokens", false, "Print tokens as they're read")
	flag.BoolVar(&traceStates, "states", false, "Show parser state transitions")
	flag.BoolVar(&traceStack, "stack", false, "Show parser stack after each action")
	flag.Usage = usage
	flag.Parse()

	if noast && fullast {
		fmt.Fprintln(os.Stderr, "cannot use -noast and -fullast together")
		os.Exit(1)
	}
	astMode := ""
	if noast {
		astMode = "noast"
	} else if fullast {
		astMode = "fullast"
	}

	args := flag.Args()

	opts := traceOptions{
		tokens:   traceTokens,
		states:   traceStates,
		stack:    traceStack,
		printAst: printAst,
		astMode:  astMode,
	}

	if multi {
		var r io.Reader
		if exprMode {
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "%s: -e requires at least one argument\n", os.Args[0])
				os.Exit(1)
			}
			r = strings.NewReader(strings.Join(args, "\n"))
		} else if len(args) == 0 {
			r = &lineBufReader{br: bufio.NewReaderSize(os.Stdin, 64)}
		} else if len(args) == 1 {
			f, err := os.Open(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer f.Close()
			r = f
		} else {
			fmt.Fprintln(os.Stderr, "with -multi use stdin or a single file")
			os.Exit(1)
		}
		if err := runMulti(r, opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if exprMode {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "%s: -e requires at least one argument\n", os.Args[0])
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(arg, opts); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return
	}

	if len(args) == 0 {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := runParserOnce(string(content), opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else if err := runParserOnFiles(args, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParserOnFiles(filenames []string, opts traceOptions) error {
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := runParserOnce(string(content), opts); err != nil {
			return err
		}
	}
	return nil
}

// runMulti reads input line by line and parses each line as one JSON value (e.g. NDJSON).
func runMulti(r io.Reader, opts traceOptions) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if err := runParserOnce(line, opts); err != nil {
			return err
		}
	}
	return sc.Err()
}

func runParserOnce(input string, opts traceOptions) error {
	lexer := lexers.NewMyLexerFromString(input)
	parser := parsers.NewMyParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	ast, err := parser.Parse(lexer, opts.astMode)
	if err != nil {
		return err
	}
	if ast != nil && opts.astMode != "noast" {
		ast.Print()
	}
	return nil
}
