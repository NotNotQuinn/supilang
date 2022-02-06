package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type SBLFile struct {
	Declarations []Declaration `@@*`
}

type Declaration struct {
	Pos        lexer.Position
	Entrypoint *string `  "entry" @Ident`
	Alias      *Alias  `|  @@`
}

type Alias struct {
	Name      string     `  "alias" @Ident`
	Keyprefix *string    `[ "prefixed" @String ]`
	Body      *AliasBody `   @@ "end"`
}

type AliasBody struct {
	Actions []*AliasAction `@@*`
}

type AliasAction struct {
	ExecuteAction     *ExecuteAction     `  @@`
	GetCompiledAction *GetCompiledAction `| @@`
}

// Execute a command, storing the output for later use
type ExecuteAction struct {
	RetrieveAction *RetrieveAction      `[ @@ "->" ]`
	SimpleAction   *ExecuteActionSimple `  @@`
	ContinueAction *ContinuedAction     `[ "->" @@ ]`
}

type RetrieveAction struct {
	Pos          lexer.Position
	RetrieveKey  *string `  "get" @String`
	RetrieveArgs *string `|  @ArgLiteral`
}

type ContinuedAction struct {
	StoreKey    *string              `  "set" @String`
	NextAction  *ExecuteActionSimple `|  @@`
	ExtraAction *ExecuteAction       `[ "->" @@ ]`
}

// Execute a command, voiding the output
type ExecuteActionSimple struct {
	Pos                 lexer.Position
	JSExec              *JSExecAction    `  "js" @@`
	PipeCommandLiterals []string         `|  ("exec" | "pipe") @String { "|" @String } `
	UseSayLiteral       bool             `|  @"say" `
	SayLiteral          *string          `   [ @String ] `
	CallAlias           *CallAliasAction `|  @@`
}

type JSExecAction struct {
	Pos        lexer.Position
	ExecString string `@JSExecString`
}

type GetCompiledAction struct {
	CompilationRoot *AliasBody       `"get" "compiled" @@ "end"`
	ContinueAction  *ContinuedAction `[ "->" @@ ]`
}

type CallAliasAction struct {
	User      *string `"call" [ @User ]`
	AliasName string  `@Ident`
}

// Custom lexer for Aliases
var aliasLexer = lexer.MustSimple([]lexer.Rule{
	// identifiers can "overwrite" keywords, otherwise keywords are priorotized
	{`Ident`, `[-a-zA-Z_0-9]{2,30}`, nil},
	{`Keyword`, `alias|end|\||exec|pipe|->|prefixed|say`, nil},
	{`User`, `@[-a-zA-Z_0-9]*`, nil},
	{`ArgLiteral`, `\$\{\d+(-\d+|\+|)\}`, nil},
	{`JSExecString`, `(\x60{3})(?:\\.|[^\x60])*(\x60{3})`, nil},
	// {`Word`, `[a-zA-Z_][a-zA-Z0-9_]`, nil},
	{`String`, `"(?:\\.|[^"])*"`, nil},
	{"comment", `#[^\n]*`, nil},
	{"whitespace", `\s+`, nil},
})

var parser = participle.MustBuild(&SBLFile{},
	participle.Lexer(aliasLexer),
	participle.Unquote("String"),
	processToken(0, 3, false, "JSExecString"),
)

// What do I even want in this language?
// ...

// Define aliases - abstract some things away and make it easier to design them

// What will help me with this goal?

// Abstract COMMON paterns in aliases, make them easier to edit and read
