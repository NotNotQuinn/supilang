package main

import (
	"log"
)

type (
	TokenType uint64
	Token     struct {
		Type  TokenType
		Value string
	}
)

const (
	TT_INVALID TokenType = iota
	TT_KEYWORD
	TT_WORD
	TT_STRING
)

var KEYWORDS []string = []string{
	"alias",
	"end",
	"exec",
	"pipe",
	"|",
}

type lexer struct {
	text    string
	pos     int
	tokens  []Token
	curchar byte
}

// advance one character
func (l *lexer) advance() {
	l.pos++
	if l.pos < len(l.text) {
		l.curchar = l.text[l.pos]
	} else {
		l.curchar = 0
	}
}

// Lex text into tokens
func (l *lexer) Lex(str string) []Token {
	// reset
	l.text = str
	l.pos = 0
	l.tokens = nil
	l.curchar = l.text[0]

	// then lex
	for l.curchar != 0 {
		if (l.curchar < ' ' || l.curchar >= 0x7F) && l.curchar != '\t' && l.curchar != '\r' && l.curchar != '\n' {
			log.Fatal("Invalid text in your file god daawm: ", l.curchar)
		} else if l.curchar == ' ' || l.curchar == '\t' || l.curchar == '\r' || l.curchar == '\n' {
			l.advance()
		} else if l.curchar == '"' {
			l.tokens = append(l.tokens, l.make_string())
		} else {
			// !@#$%^&*()_+-=[]{}\|'";:.>,</?`~abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
			l.tokens = append(l.tokens, l.make_ident())
		}
	}

	// and return
	return l.tokens
}

// Create a string, consuming characters as needed.
func (l *lexer) make_string() Token {
	l.advance()
	content := ""
	last := byte('"')
	for l.curchar != 0 && (l.curchar != '"' || last == '\\') {
		content += string(l.curchar)
		last = l.curchar
		l.advance()
	}
	l.advance()
	return Token{TT_STRING, content}
}

// Time to document so I dont forget everything tomorrow, actually so I dont have to remember everything for tomorrow.

// Create an identifier, consuming characters as needed.
func (l *lexer) make_ident() Token {
	ident := ""
	for l.curchar != 0 && !(l.curchar == ' ' || l.curchar == '\t' || l.curchar == '\r' || l.curchar == '\n') {
		ident += string(l.curchar)
		l.advance()
	}
	for _, v := range KEYWORDS {
		if v == ident {
			return Token{TT_KEYWORD, ident}
		}
	}
	return Token{TT_WORD, ident}
}
func NewLexer() lexer { return lexer{} }
