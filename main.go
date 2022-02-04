package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

type OpType uint64

const (
	// execute a supibot command
	OP_EXEC OpType = iota
	// Pipe multiple commands together
	OP_PIPE

	// // push string to a stack
	// OP_STACK_PUSH
	// // pop from a stack
	// OP_STACK_POP
	// // delete a stack
	// OP_STACK_DELETE

	// // Get a key from customData
	// OP_CUSTOM_DATA_GET
	// // Set a key to customData
	// OP_CUSTOM_DATA_SET
	// // Get all keys from customData
	// OP_CUSTOM_DATA_GET_KEYS
)

type (
	Operation struct {
		Type  OpType
		Value interface{}
	}

	Program struct {
		Ops []Operation
	}
)

func (p *Program) Compile() string {
	pipe := []string{}
	for _, op := range p.Ops {
		switch op.Type {
		case OP_EXEC:
			switch v := op.Value.(type) {
			case string:
				pipe = append(pipe, v)
			}
		case OP_PIPE:
			switch v := op.Value.(type) {
			case []string:
				pipe = append(pipe, strings.Join(v, " | "))
			}
		}
	}
	if len(pipe) == 1 {
		return pipe[0]
	} else {
		return "pipe " + strings.Join(pipe, " | abb say | null | ")
	}
}

// Custom lexer for Aliases
var aliasLexer = lexer.MustSimple([]lexer.Rule{
	{`Ident`, `[-a-zA-Z_0-9]{2,30}`, nil},
	// {`Word`, `[a-zA-Z_][a-zA-Z0-9_]`, nil},
	{`Keyword`, `alias|end|\||exec|pipe`, nil},
	{`CommandLiteral`, `"(?:\\.|[^"])*"`, nil},
	{"comment", `#[^\n]*`, nil},
	{"whitespace", `\s+`, nil},
})

var parser = participle.MustBuild(&Alias{},
	participle.Lexer(aliasLexer),
	participle.Unquote("CommandLiteral"),
)

type Alias struct {
	Name string              `"alias" @Ident`
	Body []*CommandExecution `@@* "end"`
}

type CommandExecution struct {
	ExecCommandLiteral  *string   `  "exec" @CommandLiteral`
	PipeCommandLiterals *[]string `| "pipe" @CommandLiteral { "|" @CommandLiteral }`
}

func (a *Alias) CompileToOps() []Operation {
	ops := []Operation{}
	for _, ce := range a.Body {
		if ce.PipeCommandLiterals != nil {
			ops = append(ops, Operation{OP_PIPE, *ce.PipeCommandLiterals})
		} else if ce.ExecCommandLiteral != nil {
			ops = append(ops, Operation{OP_EXEC, *ce.ExecCommandLiteral})
		}
	}
	return ops
}

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
	fmt.Println(alias.CompileToOps())
	p := (Program{alias.CompileToOps()})
	fmt.Println(p.Compile())
}
