package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

// type OpType uint64

// const (
// 	// execute a supibot command
// 	OP_EXEC OpType = iota
// 	// Pipe multiple commands together
// 	OP_PIPE

// 	// Store the next command execution into the key provided
// 	OP_STORE
// 	// Retrive the key and pipe it into the next command
// 	OP_RETRIEVE

// 	// // push string to a stack
// 	// OP_STACK_PUSH
// 	// // pop from a stack
// 	// OP_STACK_POP
// 	// // delete a stack
// 	// OP_STACK_DELETE

// 	// // Get a key from customData
// 	// OP_CUSTOM_DATA_GET
// 	// // Set a key to customData
// 	// OP_CUSTOM_DATA_SET
// 	// // Get all keys from customData
// 	// OP_CUSTOM_DATA_GET_KEYS
// )

// type (
// 	Operation struct {
// 		Type  OpType
// 		Value interface{}
// 	}

// 	Program struct {
// 		Ops []Operation
// 	}
// )

type AliasOptions struct {
	Keyprefix string
}

func (a *Alias) Getoptions() *AliasOptions {
	keyprefix := ""
	if a.Keyprefix != nil {
		keyprefix = *a.Keyprefix
	}
	return &AliasOptions{
		Keyprefix: keyprefix,
	}
}

// Compile the alias
func (a *Alias) Compile() (string, error) {
	out, err := a.Body.Compile(a.Getoptions())
	if err != nil {
		return "", fmt.Errorf("Alias: %w", err)
	}
	return "$alias add " + a.Name + " " + out, nil
}
func (ab *AliasBody) Compile(a *AliasOptions) (string, error) {
	commands := []string{}
	for i, aa := range ab.Actions {
		cmds, err := aa.Compile(a)
		if err != nil {
			return "", fmt.Errorf("AliasAction: %w", err)
		}
		commands = append(commands, cmds...)
		if i+1 != len(ab.Actions) {
			commands = append(commands, "abb say", "null")
		}
	}
	if len(commands) == 0 {
		return "", errors.New("an alias must have at least one action")
	} else if len(commands) == 1 {
		return commands[0], nil
	}
	return "pipe " + strings.Join(commands, " | "), nil
}
func (aa *AliasAction) Compile(a *AliasOptions) ([]string, error) {
	if aa.ExecuteAction != nil {
		return aa.ExecuteAction.Compile(a)
	}
	return nil, nil
}

func (aa *ExecuteAction) Compile(a *AliasOptions) ([]string, error) {
	commands := []string{}
	if aa.RetrieveKey != nil {
		escapedKey := strings.Replace(a.Keyprefix+*aa.RetrieveKey, `"`, `\\"`, -1)
		commands = append(commands, `js function:"customData.get(\"`+escapedKey+`\")"`)
	}
	if aa.SimpleAction != nil {
		out, err := aa.SimpleAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		commands = append(commands, out...)
	}
	if aa.ContinueAction != nil {
		cmds, err := aa.ContinueAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		commands = append(commands, cmds...)
	}
	return commands, nil
}
func (ca *ContinuedAction) Compile(a *AliasOptions) ([]string, error) {
	if ca.StoreKey != nil {
		escapedKey := strings.Replace(a.Keyprefix+*ca.StoreKey, `"`, `\\"`, -1)
		return []string{`js function:"customData.set(\"` + escapedKey + `\", args.join(' '))"`}, nil
	}
	return ca.NextAction.Compile(a)
}
func (ea *ExecuteActionSimple) Compile(a *AliasOptions) ([]string, error) {
	out := []string{}
	// if ea.ExecCommandLiteral != nil {
	// 	out = append(out, *ea.ExecCommandLiteral)
	if ea.PipeCommandLiterals != nil {
		out = append(out, ea.PipeCommandLiterals...)
	} else if ea.CallAlias != nil {
		code, err := ea.CallAlias.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		out = append(out, code)
	} else {
		return nil, errors.New("invalid ExecuteAction")
	}
	return out, nil
}
func (ca *CallAliasAction) Compile(a *AliasOptions) (string, error) {
	if ca.User != nil {
		return `alias try ` + *ca.User + ` ` + ca.AliasName, nil
	} else {
		return `alias run ` + ca.AliasName, nil
	}
}

var UnquoteWithChar = func(types ...string) func(p *participle.Parser) error {
	unquoteWithChar := func(s string) (string, error) {
		quote := s[1]
		s = s[2 : len(s)-1]
		out := ""
		for s != "" {
			value, _, tail, err := strconv.UnquoteChar(s, quote)
			if err != nil {
				return "", err
			}
			s = tail
			out += string(value)
		}
		return out, nil
	}
	if len(types) == 0 {
		return nil
	}
	return participle.Map(func(t lexer.Token) (lexer.Token, error) {
		value, err := unquoteWithChar(t.Value)
		if err != nil {
			return t, participle.Errorf(t.Pos, "invalid quoted string %q: %s", t.Value, err.Error())
		}
		t.Value = value
		return t, nil
	}, types...)
}

// Custom lexer for Aliases
var aliasLexer = lexer.MustSimple([]lexer.Rule{
	{`Ident`, `[-a-zA-Z_0-9]{2,30}`, nil},
	{`User`, `@[-a-zA-Z_0-9]*`, nil},
	// {`Word`, `[a-zA-Z_][a-zA-Z0-9_]`, nil},
	{`Keyword`, `alias|end|\||exec|pipe|->|prefixed`, nil},
	{`String`, `"(?:\\.|[^"])*"`, nil},
	{`StorageKey`, `s"(?:\\.|[^"])*"`, nil},
	{"comment", `#[^\n]*`, nil},
	{"whitespace", `\s+`, nil},
})

var parser = participle.MustBuild(&Alias{},
	participle.Lexer(aliasLexer),
	participle.Unquote("String"),
	UnquoteWithChar("StorageKey"),
)

type Alias struct {
	Name      string    `  "alias" @Ident`
	Keyprefix *string   `[ "prefixed" @String ]`
	Body      AliasBody `  @@ "end"`
}
type AliasBody struct {
	Actions []*AliasAction `@@*`
}

type AliasAction struct {
	ExecuteAction *ExecuteAction `@@`
}

// Execute a command, storing the output for later use
type ExecuteAction struct {
	RetrieveKey    *string              `[ "get" @String  "->" ]`
	SimpleAction   *ExecuteActionSimple `  @@`
	ContinueAction *ContinuedAction     `[ "->" @@ ]`
}

type ContinuedAction struct {
	StoreKey   *string        ` "set" @String`
	NextAction *ExecuteAction `| @@`
}

// Execute a command, voiding the output
type ExecuteActionSimple struct {
	PipeCommandLiterals []string         `   @String { "|" @String } `
	CallAlias           *CallAliasAction `|  @@`
}
type CallAliasAction struct {
	User      *string `"call" [ @User ]`
	AliasName string  `@Ident`
}

// func (a *Alias) CompileToOps() (ops []Operation) {
// 	for _, aa := range a.Body {
// 		ops = append(ops, aa.CompileToOps()...)
// 	}
// 	return ops
// }
// func (aa *AliasAction) CompileToOps() (ops []Operation) {
// 	if aa.ExecuteAction != nil {
// 		ops = append(ops, aa.ExecuteAction.CompileToOps()...)
// 	}
// 	return ops
// }
// func (ea *ExecuteAction) CompileToOps() (ops []Operation) {

// 	if ea.RetrieveKey != nil {
// 		ops = append(ops, Operation{OP_RETRIEVE, *ea.RetrieveKey})
// 	}
// 	if ea.StoreKey != nil {
// 		ops = append(ops, Operation{OP_STORE, *ea.StoreKey})
// 	}
// 	if ea.ExecCommandLiteral != nil {
// 		ops = append(ops, Operation{OP_EXEC, *ea.ExecCommandLiteral})
// 	} else if ea.PipeCommandLiterals != nil {
// 		ops = append(ops, Operation{OP_PIPE, *ea.PipeCommandLiterals})
// 	}
// 	return
// }
func main() {
	bytes, err := os.ReadFile("something.supilang")
	if err != nil {
		log.Fatal("open: ", err)
	}
	alias := &Alias{}
	err = parser.ParseBytes("something.supilang", bytes, alias)
	if err != nil {
		log.Fatal(err)
	}
	repr.Println(alias, repr.Indent("  "), repr.OmitEmpty(true))
	code, err := alias.Compile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)

}
