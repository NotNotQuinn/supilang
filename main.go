package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
	esbuild "github.com/evanw/esbuild/pkg/api"
)

type AliasOptions struct {
	// Name of the alias for this scope
	Aliasname string
	// Key prefix used for this scope
	Keyprefix string
	// Force AliasBody to output $pipe command for this scope
	ForcePipeCommand bool
	// Force all instances of $js to use errorInfo:true for this scope
	JSForceErrorInfo bool
	// Disallow argument literals (${0+}, etc) for this scope
	DisallowArgLiteral bool
	// Minify javascript for this scope (will still be preprocessed, but not minified)
	MinifyJS bool
	// Do not remove temporary keys
	KeepTempkeys bool
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
		Aliasname:          a.Name,
		Keyprefix:          keyprefix,
		ForcePipeCommand:   false,
		JSForceErrorInfo:   true,
		DisallowArgLiteral: false,
		MinifyJS:           true,
		KeepTempkeys:       true,
	}
}

type Commands struct {
	aliasCommands []string
	// tempKeys is an array of keys that will be unset after the alias has finished
	tempKeys []string
}

// appends c2 to the end of c1
func (c1 *Commands) append(c2 *Commands) {
	c1.aliasCommands = append(c1.aliasCommands, c2.aliasCommands...)
	c1.tempKeys = append(c1.tempKeys, c2.tempKeys...)
}

// Add a single command to the commands
func (c *Commands) add(commandString string) {
	c.aliasCommands = append(c.aliasCommands, commandString)
}

type CompiledAliasBody struct {
	// alias txt
	bodyText string
	// keys to be unset at the end of alias execution (that havent already been unset)
	tempKeys []string
}

// Compile the alias
func (a *Alias) Compile() (string, error) {
	out, err := a.Body.Compile(a.Getoptions())
	if err != nil {
		return "", fmt.Errorf("Alias: %w", err)
	}
	if len(out.tempKeys) > 0 {
		return "", fmt.Errorf("uncleared tempKeys: %v", out.tempKeys)
	}
	return "$alias addedit " + a.Name + " " + out.bodyText, nil
}

// eg. "|123|" or "|"
var pipechar = regexp.MustCompile(`\|(\d+)\||\|`)

func (ab *AliasBody) Compile(a *AliasOptions) (*CompiledAliasBody, error) {
	// Commands are strings to be piped together
	commands := []string{}
	// keys to be removed after the alias finishes (completely)
	tempKeys := []string{}
	if len(ab.Actions) == 0 {
		return nil, errors.New("an alias must have at least one action")
	}

	// compile actions, adding null command between them
	for i, aa := range ab.Actions {
		cmds, err := aa.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasBody: %w", err)
		}
		commands = append(commands, cmds.aliasCommands...)
		tempKeys = append(tempKeys, cmds.tempKeys...)
		if i+1 != len(ab.Actions) && len(cmds.aliasCommands) > 0 {
			commands = append(commands, "abb say", "null")
		}
	}

	// TODO: Nothing is stopping "recursed" alias bodies from resolving these
	// they should only be "removed" by the root alias body
	if a.KeepTempkeys {
		tempKeys = nil
	}
	if len(tempKeys) > 0 {
		deleteKeysJS := "let k = ["
		for i, key := range tempKeys {
			escapedKey := strings.ReplaceAll(key, `"`, `\\"`)
			deleteKeysJS += `\"` + escapedKey + `\"`
			if i+1 != len(tempKeys) {
				deleteKeysJS += ","
			}
		}
		deleteKeysJS += `];for(let i=0;i<k.length;i++)customData.set(k[i],undefined);`
		// args.join(' ') must be last   // TODO: Some way to passthrough text without removing params
		deleteKeysJS += "args.join(' ');"
		errInfo := ""
		if a.JSForceErrorInfo {
			errInfo = "errorInfo:true "
		}
		deleteKeysCommand := `js ` + errInfo + `function:"` + deleteKeysJS + `"`
		commands = append(commands, deleteKeysCommand)
		tempKeys = nil
	}

	if len(commands) == 0 {
		return nil, errors.New("an alias must be equivelent to at least one command (your alias does nothing!)")
	}

	// used in "get compiled", must output a string ready to be input to $pipe no matter what
	if a.ForcePipeCommand && len(commands) == 1 {
		commands = append([]string{"null"}, commands...)
	} else if len(commands) == 1 {
		return &CompiledAliasBody{commands[0], tempKeys}, nil
	}

	uniqueUsedPipeNums := make(map[int]bool)
	// Find used pipe characters (meaning they exist anywhere in the commands)
	// track the unique numbers they use
	matches := pipechar.FindAllString(strings.Join(commands, " "), -1)
	for _, v := range matches {
		if v == "|" {
			// -1 represents "|"
			uniqueUsedPipeNums[-1] = true
		} else {
			numInChar, err := strconv.Atoi(v[1 : len(v)-1])
			if err != nil {
				return nil, fmt.Errorf("parse int in pipe char: %w", err)
			}
			uniqueUsedPipeNums[numInChar] = true
		}
	}

	// pipe char is "|x|" where x is the lowest unused number
	pipeChar := "|"
	if len(uniqueUsedPipeNums) > 0 {
		num := 0
		usedPipeNums := make([]int, 0, len(uniqueUsedPipeNums))
		for k := range uniqueUsedPipeNums {
			usedPipeNums = append(usedPipeNums, k)
		}

		// find lowest missing int
		sort.Slice(usedPipeNums, func(i, j int) bool {
			return usedPipeNums[i] < usedPipeNums[j]
		})
		for i, v := range usedPipeNums {
			// list    [-1 0 1 2]
			// indexes [ 0 1 2 3]
			// v should equal the index - 1, if nothing is missing
			if v != i-1 {
				num = usedPipeNums[i-1] + 1
				break
			} else if i == len(usedPipeNums)-1 {
				// nothing is missing, so increase by one
				num = v + 1
				break
			}
		}
		pipeChar = "|" + fmt.Sprint(num) + "|"
	}
	return &CompiledAliasBody{"pipe _char:" + pipeChar + " " + strings.Join(commands, pipeChar), tempKeys}, nil
}
func (aa *AliasAction) Compile(a *AliasOptions) (*Commands, error) {
	out := &Commands{}
	if aa.ExecuteAction != nil {
		cmds, err := aa.ExecuteAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		out.append(cmds)
	} else if aa.GetCompiledAction != nil {
		cmds, err := aa.GetCompiledAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("AliasAction: %w", err)
		}
		out.append(cmds)
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
func (ca *GetCompiledAction) Compile(a *AliasOptions) (commands *Commands, err error) {
	commands = &Commands{}
	if ca.CompilationRoot != nil {
		aliasOpts := a.Copy()
		aliasOpts.ForcePipeCommand = true
		aliasOpts.DisallowArgLiteral = true
		result, err := ca.CompilationRoot.Compile(aliasOpts)
		if err != nil {
			return nil, err
		}
		commands.tempKeys = append(commands.tempKeys, result.tempKeys...)
		execString := result.bodyText
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
			if ca.ContinueAction.StoreKeyLocal {
				key = a.Keyprefix + key
			}
			if ca.ContinueAction.StoreKeyTemp {
				commands.tempKeys = append(commands.tempKeys, key)
			}

			ca.ContinueAction.StoreKey = nil
			ca.ContinueAction.StoreKeyLocal = false
			ca.ContinueAction.StoreKeyTemp = false
			escapedKey := strings.Replace(key, `"`, `\\"`, -1)
			commands.add("js " + errInfo + "function:\" customData.set(\\\"" + escapedKey + "\\\",'" + escapedString + "') \"")
		} else {
			commands.add("js " + errInfo + "function:\" '" + escapedString + "' \"")
		}
	}
	if ca.ContinueAction != nil {
		cmds, err := ca.ContinueAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("GetCompiledAction: %w", err)
		}
		commands.append(cmds)
	}
	return
}
func (ea *ExecuteAction) Compile(a *AliasOptions) (*Commands, error) {
	commands := Commands{}
	if ea.RetrieveAction != nil {
		cmds, err := ea.RetrieveAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands.append(cmds)
	}
	if ea.SimpleAction != nil {
		cmds, err := ea.SimpleAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands.append(cmds)
	}
	if ea.ContinueAction != nil {
		cmds, err := ea.ContinueAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteAction: %w", err)
		}
		commands.append(cmds)
	}
	return &commands, nil
}
func (ra *RetrieveAction) Compile(a *AliasOptions) (commands *Commands, err error) {
	commands = &Commands{}
	if ra.RetrieveKey != nil {
		key := *ra.RetrieveKey
		if ra.LocalRetrieveKey {
			key = a.Keyprefix + *ra.RetrieveKey
		}

		escapedKey := strings.Replace(key, `"`, `\\"`, -1)
		errInfo := ""
		if a.JSForceErrorInfo {
			errInfo = "errorInfo:true "
		}
		commands.add(`js ` + errInfo + `function:"customData.get(\"` + escapedKey + `\")"`)
	} else if ra.RetrieveArgs != nil {
		if a.DisallowArgLiteral {
			// can be disabled because "get compiled" will mess with them, making them unreliable
			return nil, participle.Errorf(ra.Pos, "arg literals are not allowed in this context")
		}
		commands.add(`abb say ` + *ra.RetrieveArgs)
	}
	return
}
func (ca *ContinuedAction) Compile(a *AliasOptions) (commands *Commands, err error) {
	commands = &Commands{}
	if ca.StoreKey != nil {
		key := *ca.StoreKey
		if ca.StoreKeyLocal {
			key = a.Keyprefix + *ca.StoreKey
		}
		fmt.Printf("ca: %v\n", ca)
		if ca.StoreKeyTemp {
			fmt.Printf("key: %v\n", key)
			commands.tempKeys = append(commands.tempKeys, key)
		}

		escapedKey := strings.Replace(key, `"`, `\\"`, -1)
		errInfo := ""
		if a.JSForceErrorInfo {
			errInfo = "errorInfo:true "
		}
		commands.add(`js ` + errInfo + `function:"customData.set(\"` + escapedKey + `\", args.join(' '))"`)
	} else if ca.NextAction != nil {
		cmds, err := ca.NextAction.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ContinuedAction: %w", err)
		}
		commands.append(cmds)
	}
	if ca.SecondContinue != nil {
		cmds, err := ca.SecondContinue.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ContinuedAction: %w", err)
		}
		commands.append(cmds)
	}
	return
}
func (ea *ExecuteActionSimple) Compile(a *AliasOptions) (*Commands, error) {
	out := &Commands{}
	if ea.JSExec != nil {
		cmds, err := ea.JSExec.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteActionSimple: %w", err)
		}
		out.append(cmds)
	} else if ea.PipeCommandLiterals != nil {
		out.aliasCommands = append(out.aliasCommands, ea.PipeCommandLiterals...)
	} else if ea.CallAlias != nil {
		cmds, err := ea.CallAlias.Compile(a)
		if err != nil {
			return nil, fmt.Errorf("ExecuteActionSimple: %w", err)
		}
		out.append(cmds)
	} else if ea.UseSayLiteral {
		// since say will just output its input, we can optimize it out
		// as long as you dont append any text
		if ea.SayLiteral != nil {
			out.add("abb say " + *ea.SayLiteral)
		}
	} else {
		return nil, errors.New("invalid ExecuteActionSimple")
	}
	return out, nil
}
func (jsa *JSExecAction) Compile(a *AliasOptions) (commands *Commands, err error) {
	commands = &Commands{}
	// Calculate injected Javascript
	unescapedJSCode := strings.Replace(jsa.ExecString.RawString, "\\`", "`", -1)

	escapedKeyprefix := strings.Replace(a.Keyprefix, `"`, `\"`, -1)
	escapedKeyprefix = strings.Replace(escapedKeyprefix, "\n", "\\n", -1)
	// runtime functions to interact with local keys
	injectedRuntime := `
		// get the local value for the key
		function getLocal(key) {
			return customData.get("` + escapedKeyprefix + `"+key)
		}
		// set the local value for the key
		function setLocal(key, value) {
			return customData.set("` + escapedKeyprefix + `"+key, value)
		}
		// get the local key prefix
		function getLocalPrefix() {
			return "` + escapedKeyprefix + `"
		}
`
	// injected gists
	injectedGistsContent := []string{}
	for _, id := range jsa.InjectedGists {
		content, err := getGistContent(id)
		if err != nil {
			return nil, participle.Errorf(jsa.Pos, "get gist content: %s", err.Error())
		}
		injectedGistsContent = append(injectedGistsContent, content)
	}
	injectedGistJS := strings.Join(injectedGistsContent, "\n\n")
	// Final injected code
	injectedCode := injectedGistJS + "\n\n" + injectedRuntime + "\n\n"
	// Minify code so that it can fit on one line, because I'm not parsing that shit
	res := esbuild.Transform(injectedCode+unescapedJSCode, esbuild.TransformOptions{
		Loader:            esbuild.LoaderJS,
		Drop:              esbuild.DropConsole, // console doesnt even exist in $js
		IgnoreAnnotations: true,
		// Tree shake to remove our runtime if it doesnt get used
		TreeShaking: esbuild.TreeShakingTrue,
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
			injectedLines := strings.Split(injectedCode, "\n")
			// If its the first line of where the user wrote....
			if l2.Line-len(injectedLines)+1 == 1 {
				lenLastInjectedLine := len(injectedLines[len(injectedLines)-1])
				loc.Column =
					//  text within source file, before js starts
					l.Column + len("```") +
						//  text written after the backtics
						(l2.Column - lenLastInjectedLine) +
						//  offset for escaping the backtic character with backslash
						strings.Count(l2.LineText[lenLastInjectedLine:], "`")
			} else {
				// add one for every backtic, because those are written as "\`"
				loc.Column = l2.Column + 1 + strings.Count(l2.LineText, "`")
			}
			loc.Filename = l.Filename
			loc.Line = l.Line + l2.Line - len(injectedLines)
			return loc.String()
		}
		for _, m := range res.Warnings {
			log.Printf("Minify JS (warning): %s: %s\n", locationToString(jsa.ExecString.Pos, *m.Location), m.Text)
			for _, n := range m.Notes {
				log.Printf("Minify JS (warning): %s: Note: %s\n", locationToString(jsa.ExecString.Pos, *n.Location), n.Text)
			}
		}
		for _, m := range res.Errors {
			log.Printf("Minify JS: %s: %s\n", locationToString(jsa.ExecString.Pos, *m.Location), m.Text)
			for _, n := range m.Notes {
				log.Printf("Minify JS: %s: Note: %s\n", locationToString(jsa.ExecString.Pos, *n.Location), n.Text)
			}
		}
		if len(res.Errors) > 0 {
			os.Exit(1)
		}
	}

	// this string must not start or end with a double quote
	// supibot trims them, thinking they are part of the parameter.
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
	importGist := ""
	if jsa.ImportedGist != nil {
		if match, err := regexp.MatchString("^[0-9a-fA-F]*$", *jsa.ImportedGist); !match || err != nil {
			if err != nil {
				return nil, err
			}
			return nil, participle.Errorf(jsa.Pos, "gist ids can only contain hexadecimal characters (0123456789abcdefABCDEF)")
		}
		if *jsa.ImportedGist == "" {
			return nil, participle.Errorf(jsa.Pos, "a gist id cannot be the empty string")
		}
		importGist = "importGist:" + *jsa.ImportedGist + " "
	}
	commands = &Commands{}
	commands.add("js " + errInfo + importGist + "function:\"" + escapedMinifiedCode + "\"")
	return commands, nil
}
func (ca *CallAliasAction) Compile(a *AliasOptions) (commands *Commands, err error) {
	commands = &Commands{}
	if ca.User != nil {
		commands.add(`alias try ` + *ca.User + ` ` + ca.AliasName)
	} else {
		commands.add(`$ ` + ca.AliasName)
	}
	return
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
