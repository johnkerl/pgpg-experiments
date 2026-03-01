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

const MyLexerStartState = 0

type MyLexerRangeTransition struct {
	from rune
	to   rune
	next int
}

var MyLexerTransitions = map[int][]MyLexerRangeTransition{
	0: {
		{from: '\x00', to: '+', next: 1},
		{from: ',', to: ',', next: 2},
		{from: '-', to: '<', next: 3},
		{from: '=', to: '=', next: 4},
		{from: '>', to: '\uffff', next: 5},
	},
	1: {
		{from: '\x00', to: '+', next: 6},
		{from: '-', to: '<', next: 7},
		{from: '>', to: '\uffff', next: 8},
	},
	3: {
		{from: '\x00', to: '+', next: 6},
		{from: '-', to: '<', next: 7},
		{from: '>', to: '\uffff', next: 8},
	},
	5: {
		{from: '\x00', to: '+', next: 6},
		{from: '-', to: '<', next: 7},
		{from: '>', to: '\uffff', next: 8},
	},
	6: {
		{from: '\x00', to: '+', next: 6},
		{from: '-', to: '<', next: 7},
		{from: '>', to: '\uffff', next: 8},
	},
	7: {
		{from: '\x00', to: '+', next: 6},
		{from: '-', to: '<', next: 7},
		{from: '>', to: '\uffff', next: 8},
	},
	8: {
		{from: '\x00', to: '+', next: 6},
		{from: '-', to: '<', next: 7},
		{from: '>', to: '\uffff', next: 8},
	},
}

var MyLexerActions = map[int]tokens.TokenType{
	1: "text",
	2: "comma",
	3: "text",
	4: "equals",
	5: "text",
	6: "text",
	7: "text",
	8: "text",
}
