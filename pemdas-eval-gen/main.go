package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"tryparse/generated/lexers"
	"tryparse/generated/parsers"
)

type traceOptions struct {
	tokens   bool
	states   bool
	stack    bool
	printAst bool
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [-e | -l] [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -v: Print the AST parse tree.\n")
	fmt.Fprintf(os.Stderr, "  -e: arguments are expressions to parse (at least one required).\n")
	fmt.Fprintf(os.Stderr, "  -l: read stdin line-by-line, evaluate each line, print result (REPL mode).\n")
	fmt.Fprintf(os.Stderr, "  With -l and stdin a TTY, -p sets the prompt (default \"> \"); use -p \"\" to disable.\n")
	fmt.Fprintf(os.Stderr, "  Without -e/-l: zero arguments = read from stdin; one or more = read from those files.\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var printAst bool
	var exprMode bool
	var lineMode bool
	var prompt string
	var traceTokens bool
	var traceStates bool
	var traceStack bool
	flag.BoolVar(&printAst, "v", false, "Print AST before evaluation")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.BoolVar(&lineMode, "l", false, "Read stdin line-by-line, evaluate each, print result (REPL)")
	flag.StringVar(&prompt, "p", "> ", "In -l mode with TTY stdin, prompt string (default \"> \"; use \"\" to disable)")
	flag.BoolVar(&traceTokens, "tokens", false, "Print tokens as they're read")
	flag.BoolVar(&traceStates, "states", false, "Show parser state transitions")
	flag.BoolVar(&traceStack, "stack", false, "Show parser stack after each action")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	opts := traceOptions{
		tokens:   traceTokens,
		states:   traceStates,
		stack:    traceStack,
		printAst: printAst,
	}

	if lineMode {
		if exprMode {
			fmt.Fprintf(os.Stderr, "%s: -e and -l are mutually exclusive", os.Args[0])
			os.Exit(1)
		}
		runREPL(opts, prompt)

	} else if exprMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "%s: -e requires at least one argument", os.Args[0])
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(arg, opts); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return

	} else {
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
}

func stdinIsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func runREPL(opts traceOptions, prompt string) {
	usePrompt := stdinIsTTY() && prompt != ""
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if usePrompt {
			fmt.Fprint(os.Stdout, prompt)
			os.Stdout.Sync()
		}
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := runParserOnce(line, opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
	if err := scanner.Err(); err != nil {
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

func runParserOnce(input string, opts traceOptions) error {
	lexer := lexers.NewMyLexer(input)
	parser := parsers.NewMyParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	ast, err := parser.Parse(lexer, "")
	if err != nil {
		return err
	}

	result, err := evaluateAST(ast, opts.printAst)
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", result)
	return nil
}
