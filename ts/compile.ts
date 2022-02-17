import type { SBLFile, Alias, AliasOptions, 
	GetCompiledAction, AliasBody, 
	AliasAction, ExecuteAction, RetrieveAction, 
	ExecuteActionSimple, ContinuedAction, 
	JSExecAction, CallAliasAction 
} from './types';
import type { Position } from './lexer';
// @ts-ignore xd no type definitions
import * as chiffon from 'chiffon';


function getOptions(alias: Alias): AliasOptions {
	let keyprefix = ""
	if (typeof alias.Keyprefix !== "undefined") {
		keyprefix = alias.Keyprefix
	}
	return {
		Aliasname:        alias.Name,
		Keyprefix:        keyprefix,
		JSForceErrorInfo: true,
		MinifyJS:         true,
		DisallowArgLiteral: false,
		ForcePipeCommand: false
	}
}

export class CompilerError extends Error {
	constructor(pos: Position, msg: string) {
		super(pos.filename+":"+pos.line+":"+pos.col+": " + msg)
	}
}

export function Compile(ast: SBLFile): string {
	let entry = ""
	let aliases: Record<string, Alias> = {}

	for (const d of ast.Declarations) {
		if (typeof d.Entrypoint !== "undefined" && entry === "") {
			entry = d.Entrypoint
		} else if (typeof d.Entrypoint !== "undefined") {
			throw new CompilerError(d.Pos, "only one entrypoint can be specified per file")
		} else if (typeof d.Alias !== "undefined") {
			if (typeof aliases[d.Alias.Name] !== "undefined") {
				throw new CompilerError(d.Pos, "duplicate alias definition: "+d.Alias.Name)
			}
			aliases[d.Alias.Name] = d.Alias
		} else {
			throw new CompilerError(d.Pos, "invalid declaration")
		}
	}
	if (entry === "" && Object.keys(aliases).length === 1) {
		return CompileAlias(ast.Declarations[0].Alias)
	} else if (typeof aliases[entry] !== "undefined") {
		return CompileAlias(aliases[entry])
	} else {
		throw new CompilerError(ast.Declarations[0]?.Pos ?? { filename: "unknown", col:1, line:1 }, "entrypoint can only be omitted if there is exactly one alias")
	}
}

function CompileAlias(a: Alias): string {
	let opts = getOptions(a)
	return "$alias addedit " + a.Name + " " + CompileAliasBody(a.Body, opts)
}
function CompileAliasBody(ab: AliasBody, a: AliasOptions) {
	let commands: string[] = []
	if (ab.Actions.length === 0) {
		throw new CompilerError(ab.Pos, "an alias must have at least one action")
	}

	// compile actions, add null command between them
	for (let i = 0; i < ab.Actions.length; i++) {
		let aa = ab.Actions[i]
		let cmds = CompileAliasAction(aa, a)
		commands.push(...cmds)
		if (i+1 !== ab.Actions.length && cmds.length > 0) {
			commands.push("abb say", "null")
		}
	}

	if (commands.length === 0) {
		throw new CompilerError(ab.Pos, "an alias must be equivelent to at least one command (your alias does nothing!)")
	}

	// used in "get compiled", must output a string ready to be input to $pipe no matter what
	if (a.ForcePipeCommand && commands.length === 1) {
		commands.unshift("null")
	} else if (commands.length === 1) {
		return commands[0]
	}

	let uniqueUsedPipeNums = new Set<number>()
	// Find used pipe characters (meaning they exist anywhere in the commands)
	// track the unique numbers they use
	let matches = [...commands.join(" ").matchAll(/\|(\d+)\||\|/g)].map(i => i[0])
	for (const v of matches) {
		if (v === "|") {
			// -1 represents "|"
			uniqueUsedPipeNums.add(-1)
		} else {
			let num = Number(v.slice(1, -1))
			uniqueUsedPipeNums.add(num)
		}
	}

	// pipe char is "|x|" where x is the lowest unused number
	let pipeChar = "|"
	if (uniqueUsedPipeNums.size > 0) {
		let num = 0
		let usedPipeNums = [...uniqueUsedPipeNums]
		usedPipeNums.sort((a,b) => a - b)
		for (let i = 0; i < usedPipeNums.length; i++) {
			let v = usedPipeNums[i]
			if (v !== i-1) {
				num = usedPipeNums[i-1] + 1
				break
			}
		}
		pipeChar = "|" + num + "|"
	}
	return "pipe _char:"+pipeChar+" "+ commands.join(" "+pipeChar+" ")
}

function CompileAliasAction(aa: AliasAction, a: AliasOptions): string[] {
	let out: string[] = []
	if (typeof aa.ExecuteAction !== "undefined") {
		let cmds = CompileExecuteAction(aa.ExecuteAction, a)
		out.push(...cmds)
	} else if (typeof aa.GetCompiledAction !== "undefined") {
		let cmds = CompileGetCompiledAction(aa.GetCompiledAction, a)
		out.push(...cmds)
	}
	return out
}

function CompileExecuteAction(ea: ExecuteAction, a: AliasOptions): string[] {
	let out: string[] = []
	if (typeof ea.RetrieveAction !== "undefined") {
		let cmds = CompileRetrieveAction(ea.RetrieveAction, a)
		out.push(...cmds)
	}
	if (typeof ea.SimpleAction !== "undefined") {
		let cmds = CompileExecuteActionSimple(ea.SimpleAction, a)
		out.push(...cmds)
	}
	if (typeof ea.ContinueAction !== "undefined") {
		let cmds = CompileContinuedAction(ea.ContinueAction, a)
		out.push(...cmds)
	}
	return out
}

function CompileRetrieveAction(ra: RetrieveAction, a: AliasOptions): string[] {
	let commands: string[] = []
	if (typeof ra.RetrieveKey !== "undefined") {
		let key = ra.RetrieveKey
		if (ra.LocalRetrieveKey) {
			key = a.Keyprefix+ra.RetrieveKey
		}

		let escapedKey = key.replace(/"/g, `\\\\"`) 
		let errInfo = ""
		if (a.JSForceErrorInfo) {
			errInfo = "errorInfo:true "
		}
		commands.push(`js `+errInfo+`function:"customData.get(\\"`+escapedKey+`\\")"`)
	} else if (typeof ra.RetrieveArgs !== "undefined") {
		if (a.DisallowArgLiteral) {
			// can be disabled because "get compiled" will mess with them, making them unreliable
			throw new CompilerError(ra.Pos, "arg literals are not allowed in this context")
		}
		commands.push("abb say " + ra.RetrieveArgs)
	}
	return commands
}

function CompileGetCompiledAction(ca: GetCompiledAction, a: AliasOptions): string[] {
	let commands: string[] = []
	if (typeof ca.CompilationRoot !== "undefined") {
		let aliasOpts = Object.assign({}, a)
		aliasOpts.ForcePipeCommand = true
		aliasOpts.DisallowArgLiteral = true
		let execString = CompileAliasBody(ca.CompilationRoot, aliasOpts)

		let escapedString = execString.slice(5).replace(/\\/g, "\\\\")
		escapedString = escapedString.replace(/"/g, `\\"`)
		escapedString = escapedString.replace(/'/g, `\\'`)
		escapedString = escapedString.replace(/\n/g, ``)
		// escape params
		escapedString = escapedString.replace(/:/g, `'+':`)
		let errInfo = ""
		if (a.JSForceErrorInfo) {
			errInfo = "errorInfo:true "
		}
		if (typeof ca.ContinueAction !== "undefined" && typeof ca.ContinueAction.StoreKey !== "undefined") {
			let key = ca.ContinueAction.StoreKey
			if (ca.ContinueAction.StoreKeyLocal) {
				key = a.Keyprefix + key
			}

			ca.ContinueAction.StoreKey = undefined
			ca.ContinueAction.StoreKeyLocal = undefined
			let escapedKey = key.replace(/"/g, `\\\\"`)
			commands.push("js "+errInfo+"function:\" customData.set(\\\""+escapedKey+"\\\",'"+escapedString+"') \"")
		} else {
			commands.push("js "+errInfo+"function:\" '"+escapedString+"' \"")
		}
	}
	if (typeof ca.ContinueAction !== "undefined") {
		let cmds = CompileContinuedAction(ca.ContinueAction, a)
		commands.push(...cmds)
	}
	return commands
}

function CompileExecuteActionSimple(ea: ExecuteActionSimple, a: AliasOptions): string[] {
	let out: string[] = []
	if (typeof ea.JSExec !== "undefined") {
		let cmds = CompileJSExecAction(ea.JSExec, a)
		out.push(...cmds)
	} else if (typeof ea.PipeCommandLiterals !== "undefined") {
		out.push(...ea.PipeCommandLiterals)
	} else if (typeof ea.CallAlias !== "undefined") {
		let cmds = CompileCallAliasAction(ea.CallAlias, a)
		out.push(...cmds)
	} else if (ea.UseSayLiteral) {
		// since say will just output its input, we can optimize it out
		// as long as you dont append any text
		if (typeof ea.SayLiteral !== "undefined") {
			out.push("abb say "+ea.SayLiteral)
		}
	} else {
		throw new CompilerError(ea.Pos, "invalid ExecuteActionSimple")
	}
	return out
}

function CompileContinuedAction(ca: ContinuedAction, a: AliasOptions): string[] {
	let commands: string[] = []
	if (typeof ca.StoreKey !== "undefined") {
		let key = ca.StoreKey
		if (ca.StoreKeyLocal) {
			key = a.Keyprefix+key
		}
		// TODO: Check if store key is "temp" and add it to a seperate list of temp keys to be removed
		// after the alias ends (requires refactoring the return type from string[] to a custom type)

		let escapedKey = key.replace(/"/g, `\\\\"`) 
		let errInfo = ""
		if (a.JSForceErrorInfo) {
			errInfo = "errorInfo:true "
		}
		commands.push(`js `+errInfo+`function:"customData.set(\\"`+escapedKey+`\\", args.join(' '))"`)
	} else if (typeof ca.NextAction !== "undefined") {
		let cmds = CompileExecuteActionSimple(ca.NextAction, a)
		commands.push(...cmds)
	}
	if (typeof ca.SecondContinue !== "undefined") {
		let cmds = CompileContinuedAction(ca.SecondContinue, a)
		commands.push(...cmds)
	}
	return commands
}

function CompileCallAliasAction(ca: CallAliasAction, _a: AliasOptions): string[] {
	if (typeof ca.User !== "undefined") {
		return [`alias try ` + ca.User + ` ` + ca.AliasName]
	} else {
		return [`$ ` + ca.AliasName]
	}
}

function CompileJSExecAction(jsa: JSExecAction, a: AliasOptions): string[] {
	// unescaping is done in parsing step
	let unescapedCode = jsa.ExecString.RawString

	let escapedKeyprefix = a.Keyprefix.replace(/"/g, `\\"`)
	escapedKeyprefix = escapedKeyprefix.replace(/\n/g, "\\n")
	let injectedRuntimeFuncs = {
		getLocal: `
		// get the local value for the key
		function getLocal(key) {
			return customData.get("` + escapedKeyprefix + `"+key)
		}`,
		setLocal: `
		// set the local value for the key
		function setLocal(key, value) {
			return customData.set("` + escapedKeyprefix + `"+key, value)
		}`,
		getLocalPrefix: `
		// get the local key prefix
		function getLocalPrefix() {
			return "` + escapedKeyprefix + `"
		}`,
	};

	let injectedRuntime = "";

	for (const func in injectedRuntimeFuncs) {
		if ((new RegExp(`\\b${func}\\b`).test(unescapedCode))) 
			injectedRuntime+=injectedRuntimeFuncs[func as keyof typeof injectedRuntimeFuncs]
	}

	let inputCode = injectedRuntime + unescapedCode

	// minify the javascript and perform tree shaking
	// using: chiffon (everything else is WAY too big to be loaded into supibot)
	// This only parses the javascript, it doesnt even validate syntax, just parses it and
	// un-parses the AST
	let res = chiffon.minify(inputCode, {
		// ....this is the ONLY option
		maxLineLen: Infinity
	})

	let minifiedCode = `(()=>{`+res+`})();`

	let escapedMinifedCode = minifiedCode.replace(/\\/g, `\\\\`)
	escapedMinifedCode = escapedMinifedCode.replace(/"/g, `\\"`)
	// remove newlines (just in case there is any, for some reason)
	escapedMinifedCode = escapedMinifedCode.replace(/\n/g, "")
	let errInfo = ""
	if (a.JSForceErrorInfo) {
		errInfo = "errorInfo:true "
	}
	let importGist = ""
	if (typeof jsa.ImportedGist !== "undefined") {
		if (!jsa.ImportedGist.match(/^[0-9a-fA-F]*$/)) {
			throw new CompilerError(jsa.Pos, "gist ids can only contain hexadecimal characters (0123456789abcdefABCDEF)")
		}
		if (jsa.ImportedGist === "") {
			throw new CompilerError(jsa.Pos, "a gist id cannot be the empty string")
		}
		importGist = "importGist:"+jsa.ImportedGist+" "
	}
	return ["js "+ errInfo + importGist + "function:\"" + escapedMinifedCode + "\""]
}
