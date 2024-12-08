package depl

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"
)

type itemType int

var (
	BuildNrName = "Buildnr"
)

const (
	eof                = -1
	itemError itemType = iota
	itemEOF
	itemBuildNr
	itemVerString
	itemText
)

type item struct {
	typ itemType

	val string
}

type lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	state stateFn
	items chan item
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	ru, s := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = s
	l.pos += l.width
	return ru
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *lexer) peek() rune {
	ru := l.next()
	l.backup()
	return ru
}

func (l *lexer) nextItem() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
}

func lexQuoteInBn(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("string error")
		case r == '"':
			l.backup()
			l.emit(itemVerString)
			return nil
		}
	}
}

func lexInsideBuildNr(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("buildnr format error")
		case unicode.IsSpace(r):
			l.ignore()
		case r == '"':
			l.ignore()
			return lexQuoteInBn
		}
	}
}

func lexBuildNr(l *lexer) stateFn {
	l.pos += len(BuildNrName)
	l.emit(itemBuildNr)
	return lexInsideBuildNr
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], BuildNrName) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBuildNr // Next state
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	return nil
}

func lexCtor(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		state: lexText,
		items: make(chan item, 2),
	}
	return l
}

func GetBuildVersionNr(str string, tokenBNR string) string {
	if tokenBNR != "" {
		BuildNrName = tokenBNR
	}

	l := lexCtor("Text lex", str)
	rr := ""
	for {
		item := l.nextItem()
		//fmt.Printf("type %v, val %q\n", item.typ, item.val)
		if item.typ == itemVerString {
			rr = item.val
		}
		if l.state == nil {
			break
		}
	}
	return rr
}

func GetVersionNrFromFile(filename string, tokenBNR string) string {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("Cannot read input file", err)
	}
	s := string(buf)
	//fmt.Println(s)
	vn := GetBuildVersionNr(s, tokenBNR)
	if vn == "" {
		log.Fatalln("Version not found")
	}

	return vn
}

func TestLexer() {
	fmt.Println("Oh my Lexer, find the version number in this string")

	str := `Hello this a cod
	and here {
		Buildnr = "1.3.5"
		other intersting stuff}..`

	rr := GetBuildVersionNr(str, "")
	if rr != "" {
		fmt.Println("Found version string: ", rr)
	} else {
		fmt.Println("Build version not found")
	}
	fmt.Println("That's all folks!")
}
