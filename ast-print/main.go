package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"astprint/generated/lexers"
	"astprint/generated/parsers"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [-e | -l] [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -e: arguments are expressions to parse (at least one required).\n")
	fmt.Fprintf(os.Stderr, "  With -l and stdin a TTY, -p sets the prompt (default \"> \"); use -p \"\" to disable.\n")
	fmt.Fprintf(os.Stderr, "  Without -e/-l: zero arguments = read from stdin; one or more = read from those files.\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var verbose bool
	var exprMode bool
	flag.BoolVar(&verbose, "v", false, "Print AST before evaluation")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if exprMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "%s: -e requires at least one argument", os.Args[0])
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(arg, verbose); err != nil {
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
			if err := runParserOnce(string(content), verbose); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if err := runParserOnFiles(args, verbose); err != nil {

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

func runParserOnFiles(filenames []string, verbose bool) error {
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := runParserOnce(string(content), verbose); err != nil {
			return err
		}
	}
	return nil
}

func runParserOnce(input string, verbose bool) error {
	lexer := lexers.NewMyLexer(input)
	parser := parsers.NewMyParser()
	ast, err := parser.Parse(lexer, "")
	if err != nil {
		return err
	}
	ast.Print()
	return nil
}
