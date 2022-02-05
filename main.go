package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
	"github.com/google/uuid"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

var m = minify.New()

func init() {
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
}

type AliasOptions struct {
	Keyprefix                 string
	RandomizePipeChar         bool
	AliasBodyForcePipeCommand bool
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
			return "", fmt.Errorf("AliasBody: %w", err)
		}
		commands = append(commands, cmds...)
		if i+1 != len(ab.Actions) && len(cmds) > 0 {
			commands = append(commands, "abb say", "null")
		}
	}

	if len(commands) == 0 {
		return "", errors.New("an alias must have at least one action")
	}
	if a.AliasBodyForcePipeCommand && len(commands) == 1 {
		commands = append([]string{"null"}, commands...)
	} else if len(commands) == 1 {
		return commands[0], nil
	}
	pipeChar := "|"
	if a.RandomizePipeChar {
		id, err := uuid.NewRandom()
		if err != nil {
			log.Fatal("could not generate uuid: ", err)
		}
		pipeChar = id.String()[:8]
	}
	return "pipe _char:" + pipeChar + " " + strings.Join(commands, " "+pipeChar+" "), nil
}
func (aa *AliasAction) Compile(a *AliasOptions) ([]string, error) {
	out := []string{}
	if aa.ExecuteAction != nil {
		cmds, err := aa.ExecuteAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		out = append(out, cmds...)
	} else if aa.GetCompiledAction != nil {
		cmds, err := aa.GetCompiledAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		out = append(out, cmds...)
	}
	if aa.ContinueAction != nil {
		cmds, err := aa.ContinueAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		out = append(out, cmds...)
	}
	return out, nil
}
func (ca *GetCompiledAction) Compile(a *AliasOptions) ([]string, error) {
	execString, err := ca.CompilationRoot.Compile(&AliasOptions{
		Keyprefix:                 a.Keyprefix,
		RandomizePipeChar:         true,
		AliasBodyForcePipeCommand: true,
	})
	if err != nil {
		return nil, fmt.Errorf("GetCompiledAction: %w", err)
	}
	// escape two sets of quotes, one for function param, one for javascript string literal
	// and remove "pipe " from the string
	escapedString := strings.Replace(execString[5:], `"`, `\"`, -1)
	escapedString = strings.Replace(escapedString, `'`, `\'`, -1)
	// escape _char: directive
	escapedString = strings.Replace(escapedString, `_char:`, `_char'+':`, -1)
	return []string{
		"js function:\" '" + escapedString + "' \"",
	}, nil
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
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands = append(commands, out...)
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
	if ea.JSExec != nil {
		// Unescape backtics from our parsing
		unescapedJSCode := strings.Replace(*ea.JSExec, "\\`", "`", -1)
		// Minify code so that it can fit on one line, because I'm not parsing that shit
		minifiedCodebuffer := new(bytes.Buffer)
		err := m.Minify("application/javascript", minifiedCodebuffer, bytes.NewBufferString(unescapedJSCode))
		if err != nil {
			return nil, participle.Errorf(ea.Pos, "ExecuteActionSimple: minify javascript: %w", err)
		}
		minifiedCode := minifiedCodebuffer.String()
		// Escape quote for funciton param
		escapedMinifiedCode := strings.Replace(minifiedCode, "\"", "\\\"", -1)
		// replace extra newlines with semicolon (idk should be fine, seems like it)
		escapedMinifiedCode = strings.Replace(escapedMinifiedCode, "\n", ";", -1)
		// leave spaces inside quotes because of a supibot bug
		out = append(out, "js function:\" "+escapedMinifiedCode+" \"")
	} else if ea.PipeCommandLiterals != nil {
		out = append(out, ea.PipeCommandLiterals...)
	} else if ea.CallAlias != nil {
		code, err := ea.CallAlias.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteActionSimple: %w", err)
		}
		out = append(out, code)
	} else {
		return nil, errors.New("invalid ExecuteActionSimple")
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

var processToken = func(trimleft, quotesize int, unquote bool, types ...string) func(p *participle.Parser) error {
	unquoteWithChar := func(s string) (string, error) {
		quote := s[trimleft]
		s = s[trimleft+quotesize : len(s)-quotesize]
		out := ""
		for s != "" {
			if unquote {
				value, _, tail, err := strconv.UnquoteChar(s, quote)
				if err != nil {
					return "", err
				}
				s = tail
				out += string(value)
			} else {
				out += s
				s = ""
			}
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

func main() {
	file := "test.supilang"
	// file := "something.supilang"
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Fatal("open: ", err)
	}
	alias := &Alias{}
	err = parser.ParseBytes(file, bytes, alias)
	if err != nil {
		log.Fatal(err)
	}
	repr.Println(alias, repr.Indent("  "), repr.Hide(lexer.Position{}), repr.OmitEmpty(true))
	code, err := alias.Compile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)
}
