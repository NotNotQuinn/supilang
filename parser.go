package main

import (
	"fmt"
	"strings"
)

type parser struct {
	tokens   []Token
	pos      int
	rawInput string
	curTok   *Token
}

func NewParser(s string) parser {
	lexer := NewLexer()

	return parser{
		tokens:   lexer.Lex(s),
		pos:      -1,
		rawInput: s,
	}
}

func (p *parser) Parse() (TreeNode, error) {
	p.scanToken()
	return p.parseAlias()
}

func (p *parser) scanToken() {
	p.pos++
	p.curTok = &p.tokens[p.pos]
}

func (p *parser) parseAlias() (*AliasNode, error) {
	_, err := p.parseKeyword("alias")
	if err != nil {
		return nil, err
	}
	ident, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	body, err := p.parseAliasBody()
	if err != nil {
		return nil, err
	}
	_, err = p.parseKeyword("end")
	if err != nil {
		return nil, err
	}
	return &AliasNode{
		Identifier: ident,
		Body:       body,
	}, nil
}
func (p *parser) parseKeyword(keywords ...string) (string, error) {
	for _, v := range keywords {
		if p.curTok.Value == v {
			p.scanToken()
			return v, nil
		}
	}
	keywordsString := "'" + strings.Join(keywords, "', or '") + "'"

	return "", fmt.Errorf("expected keyword %s found '%v'", keywordsString, p.curTok)
}
func (p *parser) parseIdent() (string, error) {
	if p.curTok.Type != TT_WORD {
		return "", fmt.Errorf("expected ident, found '%v'", p.curTok)
	}
	p.scanToken()
	return p.curTok.Value, nil
}
func (p *parser) parseAliasBody() (*AliasBodyNode, error) {
	statements := []*StatementNode{}
	if p.curTok.Type == TT_NEWLINE {
		p.scanToken()
	} else {
		return nil, fmt.Errorf("expected statement, found '%v'", p.curTok)
	}
	statement, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	statements = append(statements, statement)
	for {
		if p.curTok.Type == TT_NEWLINE {
			p.scanToken()
		} else {
			break
		}
		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	return &AliasBodyNode{
		Statements: statements,
	}, nil
}
func (p *parser) parseStatement() (*StatementNode, error) {
	keyword, err := p.parseKeyword("exec", "pipe")
	if err != nil {
		return nil, err
	}
	switch keyword {
	case "exec":
		value, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return &StatementNode{
			Command:             keyword,
			CommandSpecificPart: value,
		}, nil
	case "pipe":
		pipelist, err := p.parsePipeList()
		if err != nil {
			return nil, err
		}
		return &StatementNode{
			Command:             keyword,
			CommandSpecificPart: pipelist,
		}, nil
	}
	return nil, fmt.Errorf("unreachable end of parseStatement")
}
func (p *parser) parseString() (string, error) {
	if p.curTok.Type != TT_STRING {
		return "", fmt.Errorf("expected string literal, found '%v'", p.curTok)
	}
	// the only supported escapes are quotes
	value := strings.Replace(p.curTok.Value, `\"`, `"`, -1)
	p.scanToken()
	return value, nil
}
func (p *parser) parsePipeList() (*PipeListNode, error) {
	strings := []string{}
	str, err := p.parseString()
	if err != nil {
		return nil, err
	}
	strings = append(strings, str)
	if p.curTok.Type == TT_NEWLINE {
		p.scanToken()
	}
	_, err = p.parseKeyword("|")
	if err != nil {
		return nil, fmt.Errorf("must pipe at least 2 commands together: %w", err)
	}
	str, err = p.parseString()
	strings = append(strings, str)
	if err != nil {
		return nil, err
	}
	for {
		if p.curTok.Type == TT_NEWLINE {
			p.scanToken()
		}
		_, err = p.parseKeyword("|")
		if err != nil {
			break
		}
		str, err = p.parseString()
		strings = append(strings, str)
		if err != nil {
			return nil, err
		}
	}
	return &PipeListNode{
		Strings: strings,
	}, nil
}

type (
	AliasNode struct {
		Identifier string
		Body       *AliasBodyNode
	}
	AliasBodyNode struct {
		Statements []*StatementNode
	}
	StatementNode struct {
		Command             string
		CommandSpecificPart interface{}
	}
	PipeListNode struct {
		Strings []string
	}
)

func (a *AliasNode) Traverse() {}

type TreeNode interface {
	Traverse()
}
