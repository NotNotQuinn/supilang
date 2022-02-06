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
	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/google/uuid"
)

type AliasOptions struct {
	Keyprefix          string
	RandomizePipeChar  bool
	ForcePipeCommand   bool
	JSForceErrorInfo   bool
	DisallowArgLiteral bool
	MinifyJS           bool
}

func (a AliasOptions) Copy() *AliasOptions {
	return &a
}

func (a *Alias) Getoptions() *AliasOptions {
	keyprefix := ""
	if a.Keyprefix != nil {
		keyprefix = *a.Keyprefix
	}
	return &AliasOptions{
		Keyprefix:         keyprefix,
		RandomizePipeChar: true,
		JSForceErrorInfo:  true,
		MinifyJS:          true,
	}
}

// Compile the alias
func (a *Alias) Compile() (string, error) {
	out, err := a.Body.Compile(a.Getoptions())
	if err != nil {
		return "", fmt.Errorf("Alias: %w", err)
	}
	return "$alias addedit " + a.Name + " " + out, nil
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
	if a.ForcePipeCommand && len(commands) == 1 {
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
	// if aa.ContinueAction != nil {
	// 	cmds, err := aa.ContinueAction.Compile(a)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("AliasAction: %w", err)
	// 	}
	// 	out = append(out, cmds...)
	// }
	return out, nil
}
func (ca *GetCompiledAction) Compile(a *AliasOptions) (commands []string, err error) {
	if ca.CompilationRoot != nil {
		aliasOpts := a.Copy()
		aliasOpts.ForcePipeCommand = true
		aliasOpts.DisallowArgLiteral = true
		aliasOpts.RandomizePipeChar = true
		execString, err := ca.CompilationRoot.Compile(aliasOpts)
		if err != nil {
			return nil, fmt.Errorf("GetCompiledAction: %w", err)
		}

		// escape two sets of quotes, one for function param, one for javascript string literal
		// and remove "pipe " from the string
		escapedString := strings.Replace(execString[5:], `\`, `\\`, -1)
		escapedString = strings.Replace(escapedString, `"`, `\"`, -1)
		escapedString = strings.Replace(escapedString, `'`, `\'`, -1)
		escapedString = strings.Replace(escapedString, "\n", "", -1)
		// escape params
		escapedString = strings.Replace(escapedString, `:`, `'+':`, -1)
		// escape arg literals
		// escapedString = strings.Replace(escapedString, `${`, `$'+'{`, -1)
		errInfo := ""
		if a.JSForceErrorInfo {
			errInfo = "errorInfo:true "
		}
		if ca.ContinueAction != nil && ca.ContinueAction.StoreKey != nil {
			key := *ca.ContinueAction.StoreKey
			ca.ContinueAction.StoreKey = nil
			escapedKey := strings.Replace(a.Keyprefix+key, `"`, `\\"`, -1)
			commands = append(commands, "js "+errInfo+"function:\" customData.set(\\\""+escapedKey+"\\\",'"+escapedString+"') \"")
		}
		commands = append(commands, "js "+errInfo+"function:\" '"+escapedString+"' \"")
	}
	if ca.ContinueAction != nil {
		cmds, err := ca.ContinueAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("GetCompiledAction: %w", err)
		}
		commands = append(commands, cmds...)
	}
	return
}
func (aa *ExecuteAction) Compile(a *AliasOptions) ([]string, error) {
	commands := []string{}
	if aa.RetrieveAction != nil {
		cmds, err := aa.RetrieveAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands = append(commands, cmds...)
	}
	if aa.SimpleAction != nil {
		out, err := aa.SimpleAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands = append(commands, out...)
	}
	if aa.ContinueAction != nil {
		out, err := aa.ContinueAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands = append(commands, out...)
	}
	return commands, nil
}
func (ra *RetrieveAction) Compile(a *AliasOptions) (commands []string, err error) {
	if ra.RetrieveKey != nil {
		escapedKey := strings.Replace(a.Keyprefix+*ra.RetrieveKey, `"`, `\\"`, -1)
		errInfo := ""
		if a.JSForceErrorInfo {
			errInfo = "errorInfo:true "
		}
		commands = append(commands, `js `+errInfo+`function:"customData.get(\"`+escapedKey+`\")"`)
	} else if ra.RetrieveArgs != nil {
		if a.DisallowArgLiteral {
			// can be disabled because "get compiled" will mess with them, making them unreliable
			return nil, participle.Errorf(ra.Pos, "arg literals are not allowed in this context")
		}
		commands = append(commands, `abb say `+*ra.RetrieveArgs)
	}
	return
}
func (ca *ContinuedAction) Compile(a *AliasOptions) (commands []string, err error) {
	if ca.StoreKey != nil {
		escapedKey := strings.Replace(a.Keyprefix+*ca.StoreKey, `"`, `\\"`, -1)
		errInfo := ""
		if a.JSForceErrorInfo {
			errInfo = "errorInfo:true "
		}
		commands = append(commands, `js `+errInfo+`function:"customData.set(\"`+escapedKey+`\", args.join(' '))"`)
	} else if ca.NextAction != nil {
		cmds, err := ca.NextAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ContinuedAction: %w", err)
		}
		commands = append(commands, cmds...)
	}
	if ca.ExtraAction != nil {
		cmds, err := ca.ExtraAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ContinuedAction: %w", err)
		}
		commands = append(commands, cmds...)
	}
	return
}
func (ea *ExecuteActionSimple) Compile(a *AliasOptions) ([]string, error) {
	out := []string{}
	if ea.JSExec != nil {
		cmds, err := ea.JSExec.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteActionSimple: %w", err)
		}
		out = append(out, cmds...)
	} else if ea.PipeCommandLiterals != nil {
		out = append(out, ea.PipeCommandLiterals...)
	} else if ea.CallAlias != nil {
		code, err := ea.CallAlias.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteActionSimple: %w", err)
		}
		out = append(out, code...)
	} else if ea.UseSayLiteral {
		// since say will just output its input, we can optimize it out
		// as long as you dont append any text
		if ea.SayLiteral != nil {
			out = append(out, "abb say "+*ea.SayLiteral)
		}
	} else {
		return nil, errors.New("invalid ExecuteActionSimple")
	}
	return out, nil
}
func (jsa *JSExecAction) Compile(a *AliasOptions) ([]string, error) {
	// Unescape backtics from our parsing
	unescapedJSCode := strings.Replace(jsa.ExecString, "\\`", "`", -1)
	// Minify code so that it can fit on one line, because I'm not parsing that shit
	res := esbuild.Transform(unescapedJSCode, esbuild.TransformOptions{
		Loader:            esbuild.LoaderJS,
		Drop:              esbuild.DropConsole,
		IgnoreAnnotations: true,
		// always minify whitespace to trim newlines
		MinifyWhitespace:  true,
		MinifyIdentifiers: a.MinifyJS,
		MinifySyntax:      a.MinifyJS,
	})

	if len(res.Errors) > 0 || len(res.Warnings) > 0 {
		locationToString := func(l lexer.Position, l2 esbuild.Location) string {
			// calculate the actual locaiton of l2 in our source file
			// based on where the js token started
			var loc lexer.Position
			if l2.Line == 1 {
				loc.Column = l.Column + len("```") + l2.Column
			} else {
				// add one for every backtic, because those are written as "\`"
				loc.Column = l2.Column + 1 + strings.Count(l2.LineText, "`")
			}
			loc.Filename = l.Filename
			loc.Line = l.Line + l2.Line - 1
			return loc.String()
		}
		for _, m := range res.Warnings {
			log.Printf("Minify JS (warning): %s: %s\n", locationToString(jsa.Pos, *m.Location), m.Text)
			for _, n := range m.Notes {
				log.Printf("Minify JS (warning): %s: Note: %s\n", locationToString(jsa.Pos, *n.Location), n.Text)
			}
		}
		for _, m := range res.Errors {
			log.Printf("Minify JS: %s: %s\n", locationToString(jsa.Pos, *m.Location), m.Text)
			for _, n := range m.Notes {
				log.Printf("Minify JS: %s: Note: %s\n", locationToString(jsa.Pos, *n.Location), n.Text)
			}
		}
		if len(res.Errors) > 0 {
			os.Exit(1)
		}
	}

	minifiedCode := `(()=>{` + string(res.Code) + `})();`

	// Escape quote for funciton param
	escapedMinifiedCode := strings.Replace(minifiedCode, `\`, `\\`, -1)
	escapedMinifiedCode = strings.Replace(escapedMinifiedCode, `"`, `\"`, -1)
	// remove newlines (just in case there is any, for some reason)
	escapedMinifiedCode = strings.Replace(escapedMinifiedCode, "\n", "", -1)
	errInfo := ""
	if a.JSForceErrorInfo {
		errInfo = "errorInfo:true "
	}
	// leave spaces inside quotes because of a supibot bug
	return []string{"js " + errInfo + "function:\" " + escapedMinifiedCode + " \""}, nil
}
func (ca *CallAliasAction) Compile(a *AliasOptions) ([]string, error) {
	if ca.User != nil {
		return []string{`alias try ` + *ca.User + ` ` + ca.AliasName}, nil
	} else {
		return []string{`$ ` + ca.AliasName}, nil
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

func (ast *SBLFile) Compile() (string, error) {
	var entry string
	aliases := make(map[string]*Alias)
	for _, d := range ast.Declarations {
		if d.Entrypoint != nil && entry == "" {
			entry = *d.Entrypoint
		} else if d.Entrypoint != nil {
			return "", participle.Errorf(d.Pos, "only one entrypoint can be specified per file")
		} else if d.Alias != nil {
			if aliases[d.Alias.Name] != nil {
				return "", participle.Errorf(d.Pos, "duplicate alias definition: %s", d.Alias.Name)
			}
			aliases[d.Alias.Name] = d.Alias
		} else {
			return "", participle.Errorf(d.Pos, "invalid declaration")
		}
	}
	if entry == "" && len(aliases) == 1 {
		return ast.Declarations[0].Alias.Compile()
	} else if aliases[entry] != nil {
		return aliases[entry].Compile()
	} else {
		return "", fmt.Errorf("entrypoint can only be omitted if there is one alias")
	}
}

func main() {
	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[1]
	} else {
		fmt.Printf("Usage: %s file\n", os.Args[0])
		os.Exit(1)
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("open: ", err)
	}
	fileAST := &SBLFile{}
	err = parser.ParseBytes(filename, bytes, fileAST)
	if err != nil {
		log.Fatal(err)
	}
	repr.Println(fileAST, repr.Indent("  "), repr.Hide(lexer.Position{}), repr.OmitEmpty(true))
	code, err := fileAST.Compile()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)

	os.WriteFile("out.alias", []byte(code), 0644)

}
