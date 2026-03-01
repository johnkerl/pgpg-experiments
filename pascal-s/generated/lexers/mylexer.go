package lexers

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

const MyLexerBufSize = 4096

type MyLexer struct {
	reader        *bufio.Reader
	buf           []byte
	tokenStart    int
	tokenLocation *tokens.TokenLocation
	atEOF         bool
}

var _ liblexers.AbstractLexer = (*MyLexer)(nil)

func NewMyLexer(r io.Reader) liblexers.AbstractLexer {
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}
	return &MyLexer{
		reader:        reader,
		buf:           make([]byte, 0, MyLexerBufSize),
		tokenLocation: tokens.NewTokenLocation(),
	}
}

// NewMyLexerFromString returns a lexer over s (convenience for tests and -e mode).
func NewMyLexerFromString(s string) liblexers.AbstractLexer {
	return NewMyLexer(strings.NewReader(s))
}

func (lexer *MyLexer) ensureFill(needBytes int) {
	for needBytes > len(lexer.buf) && !lexer.atEOF {
		chunk := make([]byte, MyLexerBufSize)
		n, err := lexer.reader.Read(chunk)
		if n > 0 {
			lexer.buf = append(lexer.buf, chunk[:n]...)
		}
		if err == io.EOF {
			lexer.atEOF = true
			return
		}
		if err != nil {
			lexer.atEOF = true
			return
		}
	}
}

func (lexer *MyLexer) peekRuneAt(byteOffset int) (rune, int) {
	lexer.ensureFill(byteOffset + utf8.UTFMax)
	if byteOffset >= len(lexer.buf) {
		return 0, 0
	}
	r, width := utf8.DecodeRune(lexer.buf[byteOffset:])
	if width == 0 {
		return 0, 0
	}
	return r, width
}

func (lexer *MyLexer) Scan() *tokens.Token {
	lexer.ensureFill(lexer.tokenStart + 1)
	if lexer.tokenStart >= len(lexer.buf) && lexer.atEOF {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}

	for {
		if lexer.tokenStart >= len(lexer.buf) {
			if lexer.atEOF {
				return tokens.NewEOFToken(lexer.tokenLocation)
			}
			lexer.ensureFill(lexer.tokenStart + 1)
			if lexer.tokenStart >= len(lexer.buf) {
				return tokens.NewEOFToken(lexer.tokenLocation)
			}
		}

		startLocation := *lexer.tokenLocation
		scanOffset := lexer.tokenStart
		state := MyLexerStartState
		lastAcceptState := -1
		lastAcceptOffset := scanOffset

		for {
			if scanOffset >= len(lexer.buf) {
				if !lexer.atEOF {
					lexer.ensureFill(scanOffset + utf8.UTFMax)
				}
				if scanOffset >= len(lexer.buf) {
					break
				}
			}
			r, width := lexer.peekRuneAt(scanOffset)
			if width == 0 {
				break
			}
			nextState, ok := MyLexerLookupTransition(state, r)
			if !ok {
				break
			}
			scanOffset += width
			state = nextState
			if _, ok := MyLexerActions[state]; ok {
				lastAcceptState = state
				lastAcceptOffset = scanOffset
			}
		}

		if lastAcceptState < 0 {
			r, _ := lexer.peekRuneAt(lexer.tokenStart)
			return tokens.NewErrorToken(fmt.Sprintf("lexer: unrecognized input %q", r), lexer.tokenLocation)
		}

		lexemeText := string(lexer.buf[lexer.tokenStart:lastAcceptOffset])
		lexeme := []rune(lexemeText)
		for len(lexemeText) > 0 {
			r, w := utf8.DecodeRuneInString(lexemeText)
			lexer.tokenLocation.LocateRune(r, w)
			lexemeText = lexemeText[w:]
		}
		lexer.buf = lexer.buf[lastAcceptOffset:]
		lexer.tokenStart = 0
		tokenType := MyLexerActions[lastAcceptState]
		if MyLexerIsIgnoredToken(tokenType) {
			continue
		}
		return tokens.NewToken(lexeme, tokenType, &startLocation)
	}
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
		{from: 'A', to: 'Z', next: 11},
		{from: '_', to: '_', next: 12},
		{from: 'a', to: 'a', next: 13},
		{from: 'b', to: 'c', next: 14},
		{from: 'd', to: 'd', next: 15},
		{from: 'e', to: 'l', next: 14},
		{from: 'm', to: 'm', next: 16},
		{from: 'n', to: 'o', next: 14},
		{from: 'p', to: 'p', next: 17},
		{from: 'q', to: 'z', next: 14},
	},
	5: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	6: {
		{from: '*', to: '*', next: 27},
		{from: '+', to: '+', next: 28},
		{from: '-', to: '-', next: 29},
		{from: '.', to: '.', next: 30},
		{from: '/', to: '/', next: 31},
		{from: '0', to: '9', next: 32},
		{from: 'A', to: 'Z', next: 33},
		{from: '_', to: '_', next: 34},
		{from: 'a', to: 'z', next: 35},
	},
	7: {
		{from: '*', to: '*', next: 27},
		{from: '+', to: '+', next: 28},
		{from: '-', to: '-', next: 29},
		{from: '.', to: '.', next: 30},
		{from: '/', to: '/', next: 31},
		{from: '0', to: '9', next: 32},
		{from: 'A', to: 'Z', next: 33},
		{from: '_', to: '_', next: 34},
		{from: 'a', to: 'z', next: 35},
	},
	8: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	9: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	10: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 36},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	11: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	12: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	13: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'm', next: 26},
		{from: 'n', to: 'n', next: 37},
		{from: 'o', to: 'z', next: 26},
	},
	14: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	15: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'h', next: 26},
		{from: 'i', to: 'i', next: 38},
		{from: 'j', to: 'z', next: 26},
	},
	16: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'n', next: 26},
		{from: 'o', to: 'o', next: 39},
		{from: 'p', to: 'z', next: 26},
	},
	17: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'q', next: 26},
		{from: 'r', to: 'r', next: 40},
		{from: 's', to: 'z', next: 26},
	},
	18: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	19: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	20: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	21: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	22: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	23: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	24: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	25: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	26: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	27: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	28: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	29: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	30: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	31: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	32: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	33: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	34: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	35: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	36: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 36},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	37: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'c', next: 26},
		{from: 'd', to: 'd', next: 41},
		{from: 'e', to: 'z', next: 26},
	},
	38: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'u', next: 26},
		{from: 'v', to: 'v', next: 42},
		{from: 'w', to: 'z', next: 26},
	},
	39: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'c', next: 26},
		{from: 'd', to: 'd', next: 43},
		{from: 'e', to: 'z', next: 26},
	},
	40: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'n', next: 26},
		{from: 'o', to: 'o', next: 44},
		{from: 'p', to: 'z', next: 26},
	},
	41: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	42: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	43: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
	44: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'f', next: 26},
		{from: 'g', to: 'g', next: 45},
		{from: 'h', to: 'z', next: 26},
	},
	45: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'q', next: 26},
		{from: 'r', to: 'r', next: 46},
		{from: 's', to: 'z', next: 26},
	},
	46: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'a', next: 47},
		{from: 'b', to: 'z', next: 26},
	},
	47: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'l', next: 26},
		{from: 'm', to: 'm', next: 48},
		{from: 'n', to: 'z', next: 26},
	},
	48: {
		{from: '*', to: '*', next: 18},
		{from: '+', to: '+', next: 19},
		{from: '-', to: '-', next: 20},
		{from: '.', to: '.', next: 21},
		{from: '/', to: '/', next: 22},
		{from: '0', to: '9', next: 23},
		{from: 'A', to: 'Z', next: 24},
		{from: '_', to: '_', next: 25},
		{from: 'a', to: 'z', next: 26},
	},
}

var MyLexerActions = map[int]tokens.TokenType{
	1:  "!whitespace",
	2:  "!whitespace",
	3:  "!whitespace",
	4:  "!whitespace",
	5:  "multiplying_op",
	6:  "identifier",
	7:  "identifier",
	8:  "identifier",
	9:  "multiplying_op",
	10: "identifier",
	11: "identifier",
	12: "identifier",
	13: "identifier",
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
	31: "identifier",
	32: "identifier",
	33: "identifier",
	34: "identifier",
	35: "identifier",
	36: "identifier",
	37: "identifier",
	38: "identifier",
	39: "identifier",
	40: "identifier",
	41: "multiplying_op",
	42: "multiplying_op",
	43: "multiplying_op",
	44: "identifier",
	45: "identifier",
	46: "identifier",
	47: "identifier",
	48: "program",
}
