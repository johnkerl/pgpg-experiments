package lexers

import (
	"fmt"
	"strings"
	"unicode/utf8"

	liblexers "github.com/johnkerl/pgpg/lib/go/pkg/lexers"
	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
)

type MyLexer struct {
	inputText     string
	inputLength   int
	tokenLocation *tokens.TokenLocation
}

var _ liblexers.AbstractLexer = (*MyLexer)(nil)

func NewMyLexer(inputText string) liblexers.AbstractLexer {
	return &MyLexer{
		inputText:     inputText,
		inputLength:   len(inputText),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *MyLexer) Scan() *tokens.Token {
	for {
		if lexer.tokenLocation.ByteOffset >= lexer.inputLength {
			return tokens.NewEOFToken(lexer.tokenLocation)
		}

		startLocation := *lexer.tokenLocation
		scanLocation := *lexer.tokenLocation
		state := MyLexerStartState
		lastAcceptState := -1
		lastAcceptLocation := scanLocation

		for {
			if scanLocation.ByteOffset >= lexer.inputLength {
				break
			}
			r, width := lexer.peekRuneAt(scanLocation.ByteOffset)
			nextState, ok := MyLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanLocation.LocateRune(r, width)
			state = nextState
			if _, ok := MyLexerActions[state]; ok {
				lastAcceptState = state
				lastAcceptLocation = scanLocation
			}
		}

		if lastAcceptState < 0 {
			r, _ := lexer.peekRuneAt(lexer.tokenLocation.ByteOffset)
			return tokens.NewErrorToken(fmt.Sprintf("lexer: unrecognized input %q", r), lexer.tokenLocation)
		}

		lexemeText := lexer.inputText[lexer.tokenLocation.ByteOffset:lastAcceptLocation.ByteOffset]
		lexeme := []rune(lexemeText)
		*lexer.tokenLocation = lastAcceptLocation
		tokenType := MyLexerActions[lastAcceptState]
		if MyLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
}

func (lexer *MyLexer) peekRuneAt(byteOffset int) (rune, int) {
	r, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])
	return r, width
}

func MyLexerLookupTransition(state int, r rune) (int, bool) {
	transitionsForState, ok := MyLexerTransitions[state]
	if !ok {
		return 0, false
	}
	for _, tr := range transitionsForState {
		if r < tr.from {
			return 0, false
		}
		if r >= tr.from && r <= tr.to {
			return tr.next, true
		}
	}
	return 0, false
}
func MyLexerIsIgnoredToken(tokenType tokens.TokenType) bool {
	return strings.HasPrefix(string(tokenType), "!")
}

const MyLexerStartState = 0

type MyLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var MyLexerTransitions = map[int][]MyLexerRangeTransition{
	0: {
		{from: '\t', to: '\t', next: 1},
		{from: '\n', to: '\n', next: 2},
		{from: '\r', to: '\r', next: 3},
		{from: ' ', to: ' ', next: 4},
		{from: '*', to: '*', next: 5},
		{from: '+', to: '+', next: 6},
		{from: '-', to: '-', next: 7},
		{from: '.', to: '.', next: 8},
		{from: '/', to: '/', next: 9},
		{from: '0', to: '9', next: 10},
		{from: '<', to: '<', next: 11},
		{from: '=', to: '=', next: 12},
		{from: '>', to: '>', next: 13},
		{from: 'A', to: 'Z', next: 14},
		{from: '_', to: '_', next: 15},
		{from: 'a', to: 'a', next: 16},
		{from: 'b', to: 'c', next: 17},
		{from: 'd', to: 'd', next: 18},
		{from: 'e', to: 'l', next: 17},
		{from: 'm', to: 'm', next: 19},
		{from: 'n', to: 'n', next: 17},
		{from: 'o', to: 'o', next: 20},
		{from: 'p', to: 'p', next: 21},
		{from: 'q', to: 'z', next: 17},
	},
	5: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	6: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	7: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	8: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	9: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	10: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	11: {
		{from: '=', to: '=', next: 31},
		{from: '>', to: '>', next: 32},
	},
	13: {
		{from: '=', to: '=', next: 33},
	},
	14: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	15: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	16: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'm', next: 30},
		{from: 'n', to: 'n', next: 34},
		{from: 'o', to: 'z', next: 30},
	},
	17: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	18: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'h', next: 30},
		{from: 'i', to: 'i', next: 35},
		{from: 'j', to: 'z', next: 30},
	},
	19: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'n', next: 30},
		{from: 'o', to: 'o', next: 36},
		{from: 'p', to: 'z', next: 30},
	},
	20: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'q', next: 30},
		{from: 'r', to: 'r', next: 37},
		{from: 's', to: 'z', next: 30},
	},
	21: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'q', next: 30},
		{from: 'r', to: 'r', next: 38},
		{from: 's', to: 'z', next: 30},
	},
	22: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	23: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	24: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	25: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	26: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	27: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	28: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	29: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	30: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	34: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'c', next: 30},
		{from: 'd', to: 'd', next: 39},
		{from: 'e', to: 'z', next: 30},
	},
	35: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'u', next: 30},
		{from: 'v', to: 'v', next: 40},
		{from: 'w', to: 'z', next: 30},
	},
	36: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'c', next: 30},
		{from: 'd', to: 'd', next: 41},
		{from: 'e', to: 'z', next: 30},
	},
	37: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	38: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'n', next: 30},
		{from: 'o', to: 'o', next: 42},
		{from: 'p', to: 'z', next: 30},
	},
	39: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	40: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	41: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
	42: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'f', next: 30},
		{from: 'g', to: 'g', next: 43},
		{from: 'h', to: 'z', next: 30},
	},
	43: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'q', next: 30},
		{from: 'r', to: 'r', next: 44},
		{from: 's', to: 'z', next: 30},
	},
	44: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'a', next: 45},
		{from: 'b', to: 'z', next: 30},
	},
	45: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'l', next: 30},
		{from: 'm', to: 'm', next: 46},
		{from: 'n', to: 'z', next: 30},
	},
	46: {
		{from: '*', to: '*', next: 22},
		{from: '+', to: '+', next: 23},
		{from: '-', to: '-', next: 24},
		{from: '.', to: '.', next: 25},
		{from: '/', to: '/', next: 26},
		{from: '0', to: '9', next: 27},
		{from: 'A', to: 'Z', next: 28},
		{from: '_', to: '_', next: 29},
		{from: 'a', to: 'z', next: 30},
	},
}

var MyLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	5:  "multiplying_op",
	6:  "adding_op",
	7:  "adding_op",
	8:  "identifier",
	9:  "multiplying_op",
	10: "identifier",
	11: "relational_op",
	12: "relational_op",
	13: "relational_op",
	14: "identifier",
	15: "identifier",
	16: "identifier",
	17: "identifier",
	18: "identifier",
	19: "identifier",
	20: "identifier",
	21: "identifier",
	22: "identifier",
	23: "identifier",
	24: "identifier",
	25: "identifier",
	26: "identifier",
	27: "identifier",
	28: "identifier",
	29: "identifier",
	30: "identifier",
	31: "relational_op",
	32: "relational_op",
	33: "relational_op",
	34: "identifier",
	35: "identifier",
	36: "identifier",
	37: "adding_op",
	38: "identifier",
	39: "multiplying_op",
	40: "multiplying_op",
	41: "multiplying_op",
	42: "identifier",
	43: "identifier",
	44: "identifier",
	45: "identifier",
	46: "program",
}
