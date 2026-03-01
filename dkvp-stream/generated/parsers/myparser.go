package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

type MyParser struct {
	Trace            *MyParserTraceHooks
	stashedLookahead *tokens.Token
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
				var node *asts.ASTNode
				useFullTree := (astMode == "fullast")
				if !useFullTree && prod.hasPassthrough {
					node = rhsNodes[prod.passthroughIndex]
				} else if !useFullTree && prod.hasWithAppendedChildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					for _, ci := range prod.withAppendedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasWithPrependedChildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					for _, ci := range prod.withPrependedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasWithAdoptedGrandchildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					for _, ci := range prod.withAdoptedGrandchildren {
						childNode := rhsNodes[ci]
						if childNode != nil && childNode.Children != nil {
							newChildren = append(newChildren, childNode.Children...)
						}
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasHint {
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = prod.lhs
					}
					var parentToken *tokens.Token
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
					} else if prod.parentIndex >= 0 && prod.parentIndex < len(rhsNodes) {
						parentToken = rhsNodes[prod.parentIndex].Token
					}
					hintChildren := make([]*asts.ASTNode, len(prod.childIndices))
					for i, ci := range prod.childIndices {
						hintChildren[i] = rhsNodes[ci]
					}
					node = asts.NewASTNode(parentToken, nodeType, hintChildren)
				} else if prod.rhsCount == 1 {
					node = rhsNodes[0]
				} else if prod.rhsCount == 0 {
					node = asts.NewASTNode(nil, prod.lhs, []*asts.ASTNode{})
				} else {
					node = asts.NewASTNode(nil, prod.lhs, rhsNodes)
				}
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
		case MyParserActionAcceptAndYield:
			return nil, fmt.Errorf("parse error: multiple objects; use ParseOne for multi-object input")
		default:
			return nil, fmt.Errorf("parse error: no action")
		}
	}
}

// ParseOne parses one record from the lexer. It is for multi-object input: call in a loop until done.
// Returns (ast, true, nil) on EOF after a record, (ast, false, nil) when more input follows, or (nil, false, err) on error.
func (parser *MyParser) ParseOne(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, bool, error) {
	if lexer == nil {
		return nil, false, fmt.Errorf("parser: nil lexer")
	}
	stateStack := []int{0}
	nodeStack := []*asts.ASTNode{}
	var lookahead *tokens.Token
	if parser.stashedLookahead != nil {
		lookahead = parser.stashedLookahead
		parser.stashedLookahead = nil
	} else {
		lookahead = lexer.Scan()
	}
	if parser.Trace != nil && parser.Trace.OnToken != nil {
		parser.Trace.OnToken(lookahead)
	}
	for {
		if lookahead == nil {
			return nil, false, fmt.Errorf("parser: lexer returned nil token")
		}
		if lookahead.Type == tokens.TokenTypeError {
			return nil, false, fmt.Errorf("lexer error: %s", string(lookahead.Lexeme))
		}
		state := stateStack[len(stateStack)-1]
		action, ok := MyParserActions[state][lookahead.Type]
		if !ok {
			return nil, false, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
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
				var node *asts.ASTNode
				useFullTree := (astMode == "fullast")
				if !useFullTree && prod.hasPassthrough {
					node = rhsNodes[prod.passthroughIndex]
				} else if !useFullTree && prod.hasWithAppendedChildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					for _, ci := range prod.withAppendedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasWithPrependedChildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					for _, ci := range prod.withPrependedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasWithAdoptedGrandchildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					for _, ci := range prod.withAdoptedGrandchildren {
						childNode := rhsNodes[ci]
						if childNode != nil && childNode.Children != nil {
							newChildren = append(newChildren, childNode.Children...)
						}
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasHint {
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = prod.lhs
					}
					var parentToken *tokens.Token
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
					} else if prod.parentIndex >= 0 && prod.parentIndex < len(rhsNodes) {
						parentToken = rhsNodes[prod.parentIndex].Token
					}
					hintChildren := make([]*asts.ASTNode, len(prod.childIndices))
					for i, ci := range prod.childIndices {
						hintChildren[i] = rhsNodes[ci]
					}
					node = asts.NewASTNode(parentToken, nodeType, hintChildren)
				} else if prod.rhsCount == 1 {
					node = rhsNodes[0]
				} else if prod.rhsCount == 0 {
					node = asts.NewASTNode(nil, prod.lhs, []*asts.ASTNode{})
				} else {
					node = asts.NewASTNode(nil, prod.lhs, rhsNodes)
				}
				nodeStack = append(nodeStack, node)
			}
			state = stateStack[len(stateStack)-1]
			nextState, ok := MyParserGotos[state][prod.lhs]
			if !ok {
				return nil, false, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case MyParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, false, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
			if astMode == "noast" {
				return nil, true, nil
			}
			return asts.NewAST(nodeStack[0]), true, nil
		case MyParserActionAcceptAndYield:
			if len(nodeStack) != 1 {
				return nil, false, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
			parser.stashedLookahead = lookahead
			if astMode == "noast" {
				return nil, false, nil
			}
			return asts.NewAST(nodeStack[0]), false, nil
		default:
			return nil, false, fmt.Errorf("parse error: no action")
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
	MyParserActionAcceptAndYield
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
	case MyParserActionAcceptAndYield:
		return "accept_and_yield"
	default:
		return "unknown"
	}
}

type MyParserProduction struct {
	lhs                         asts.NodeType
	rhsCount                    int
	hasHint                     bool
	hasPassthrough              bool
	hasParentLiteral            bool
	hasWithAppendedChildren     bool
	hasWithPrependedChildren    bool
	hasWithAdoptedGrandchildren bool
	parentIndex                 int
	passthroughIndex            int
	parentLiteral               string
	childIndices                []int
	withAppendedChildren        []int
	withPrependedChildren       []int
	withAdoptedGrandchildren    []int
	nodeType                    asts.NodeType
}

var MyParserActions = map[int]map[tokens.TokenType]MyParserAction{
	0: {
		tokens.TokenType("text"): {Kind: MyParserActionShift, Target: 4},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: MyParserActionReduce, Target: 2},
		tokens.TokenType("comma"): {Kind: MyParserActionReduce, Target: 2},
		tokens.TokenType("text"):  {Kind: MyParserActionReduce, Target: 2},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: MyParserActionReduce, Target: 1},
		tokens.TokenType("comma"):  {Kind: MyParserActionShift, Target: 5},
		tokens.TokenType("equals"): {Kind: MyParserActionAcceptAndYield},
		tokens.TokenType("text"):   {Kind: MyParserActionAcceptAndYield},
	},
	3: {
		tokens.TokenTypeEOF:        {Kind: MyParserActionAccept},
		tokens.TokenType("comma"):  {Kind: MyParserActionAcceptAndYield},
		tokens.TokenType("equals"): {Kind: MyParserActionAcceptAndYield},
		tokens.TokenType("text"):   {Kind: MyParserActionAcceptAndYield},
	},
	4: {
		tokens.TokenTypeEOF:        {Kind: MyParserActionReduce, Target: 6},
		tokens.TokenType("comma"):  {Kind: MyParserActionReduce, Target: 6},
		tokens.TokenType("equals"): {Kind: MyParserActionShift, Target: 6},
		tokens.TokenType("text"):   {Kind: MyParserActionReduce, Target: 6},
	},
	5: {
		tokens.TokenType("text"): {Kind: MyParserActionShift, Target: 4},
	},
	6: {
		tokens.TokenTypeEOF:       {Kind: MyParserActionReduce, Target: 5},
		tokens.TokenType("comma"): {Kind: MyParserActionReduce, Target: 5},
		tokens.TokenType("text"):  {Kind: MyParserActionShift, Target: 8},
	},
	7: {
		tokens.TokenTypeEOF:       {Kind: MyParserActionReduce, Target: 3},
		tokens.TokenType("comma"): {Kind: MyParserActionReduce, Target: 3},
		tokens.TokenType("text"):  {Kind: MyParserActionReduce, Target: 3},
	},
	8: {
		tokens.TokenTypeEOF:       {Kind: MyParserActionReduce, Target: 4},
		tokens.TokenType("comma"): {Kind: MyParserActionReduce, Target: 4},
		tokens.TokenType("text"):  {Kind: MyParserActionReduce, Target: 4},
	},
}

var MyParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Pair"):   1,
		asts.NodeType("Pairs"):  2,
		asts.NodeType("Record"): 3,
	},
	5: {
		asts.NodeType("Pair"): 7,
	},
}

var MyParserProductions = []MyParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Record"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Pairs"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "record", childIndices: []int{0}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("record")},
	{lhs: asts.NodeType("Pairs"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: true, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{2}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Pair"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("key_value")},
	{lhs: asts.NodeType("Pair"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("key_empty_value")},
	{lhs: asts.NodeType("Pair"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("value_only")},
}
