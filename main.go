package main

import (
	"fmt"
	"log"
	"os"
	"strings"
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

func main() {
	bytes, err := os.ReadFile("something.supilang")
	if err != nil {
		log.Fatal("open: ", err)
	}

	lexer := NewLexer()
	tokens := lexer.Lex(string(bytes))
	fmt.Println(tokens)

	// p := Program{[]Operation{
	// 	{OP_EXEC, "ping"},
	// 	{OP_PIPE, []string{"ping", "remindme in 30s"}},
	// }}
	// fmt.Println(p.Compile())
}
