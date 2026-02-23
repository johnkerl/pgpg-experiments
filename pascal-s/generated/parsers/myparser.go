package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/lib/go/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/lib/go/pkg/lexers"
	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
)

type MyParser struct {
	Trace *MyParserTraceHooks
}

type MyParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action MyParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewMyParser() *MyParser { return &MyParser{} }

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var MyParserNoASTSentinel = &asts.ASTNode{}

func (parser *MyParser) Parse(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, error) {
	if lexer == nil {
		return nil, fmt.Errorf("parser: nil lexer")
	}
	stateStack := []int{0}
	nodeStack := []*asts.ASTNode{}
	lookahead := lexer.Scan()
	if parser.Trace != nil && parser.Trace.OnToken != nil {
		parser.Trace.OnToken(lookahead)
	}
	for {
		if lookahead == nil {
			return nil, fmt.Errorf("parser: lexer returned nil token")
		}
		if lookahead.Type == tokens.TokenTypeError {
			return nil, fmt.Errorf("lexer error: %s", string(lookahead.Lexeme))
		}
		state := stateStack[len(stateStack)-1]
		action, ok := MyParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case MyParserActionShift:
			if astMode == "noast" {
				nodeStack = append(nodeStack, MyParserNoASTSentinel)
			} else {
				nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			}
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case MyParserActionReduce:
			prod := MyParserProductions[action.Target]
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if astMode == "noast" {
				nodeStack = append(nodeStack, MyParserNoASTSentinel)
			} else {
				if prod.rhsCount == 0 {
					rhsNodes = []*asts.ASTNode{}
				}
				node := asts.NewASTNode(nil, prod.lhs, rhsNodes)
				nodeStack = append(nodeStack, node)
			}
			state = stateStack[len(stateStack)-1]
			nextState, ok := MyParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case MyParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
			if astMode == "noast" {
				return nil, nil
			}
			return asts.NewAST(nodeStack[0]), nil
		default:
			return nil, fmt.Errorf("parse error: no action")
		}
	}
}

// AttachCLITrace installs tracing hooks for CLI debugging.
func (parser *MyParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &MyParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatMyParserToken(tok))
		},
		OnAction: func(state int, action MyParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatMyParserAction(action), tokenTypeNameMyParser(lookahead), tokenLexemeMyParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatMyParserIntStack(stateStack), formatMyParserNodeStack(nodeStack))
		},
	}
}

type MyParserActionKind int

const (
	MyParserActionShift MyParserActionKind = iota
	MyParserActionReduce
	MyParserActionAccept
)

type MyParserAction struct {
	Kind   MyParserActionKind
	Target int
}

func formatMyParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNameMyParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemeMyParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatMyParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatMyParserNodeStack(stack []*asts.ASTNode) string {
	parts := make([]string, len(stack))
	for i, node := range stack {
		if node == nil {
			parts[i] = "<nil>"
			continue
		}
		parts[i] = string(node.Type)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatMyParserAction(action MyParserAction) string {
	switch action.Kind {
	case MyParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case MyParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case MyParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

type MyParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var MyParserActions = map[int]map[tokens.TokenType]MyParserAction{
	0: {
		tokens.TokenType("program"): {Kind: MyParserActionShift, Target: 2},
	},
	1: {
		tokens.TokenTypeEOF: {Kind: MyParserActionAccept},
	},
	2: {
		tokens.TokenType("identifier"): {Kind: MyParserActionShift, Target: 3},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: MyParserActionReduce, Target: 1},
	},
}

var MyParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Program"): 1,
	},
}

var MyParserProductions = []MyParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1},
	{lhs: asts.NodeType("Program"), rhsCount: 2},
}
