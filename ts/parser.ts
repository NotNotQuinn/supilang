import { lexFile, TokenType, Token } from './lexer';
import { SBLFile, Declaration, Alias, 
	AliasBody, AliasAction, ExecuteAction, 
	GetCompiledAction, RetrieveAction, 
	ExecuteActionSimple, ContinuedAction, 
	JSExecAction, CallAliasAction 
} from './types';

// Error thrown when parsing couldnt be done correctly,
// NOT all errors from the parser are this type of error
export class ParsingError extends Error {}

// internal parser expectation
export type Expectation = {
	// expected keyword
	keyword: string
	// error message
	message: string
}

export class Parser {
	tokens: Token[] = [];
	expectations: Expectation[] = [];
	filename: string;
	curtokIndex: any = 0;
	// current token
	private tok(): Token {
		return this.tokens[this.curtokIndex]
	}
	// move past the current token
	private scanTok() {
		this.curtokIndex++
		if (this.curtokIndex > this.tokens.length) {
			throw new Error("token overflow")
		}
	}
	// raise an error because the current token is unexpected
	private unexpectedToken(msg: string = ""): never {
		console.error("Unexpected token: "+msg, this.tok())
		if (this.tok().type === TokenType.EOF) {
			if (this.expectations.length > 0) {
				msg = this.expectations[this.expectations.length-1].message
			}
			throw new ParsingError("unexpected EOF: " + msg)
		}
		throw new ParsingError("unexpected token: " + msg)
	}
	// Try the callback, if it fails due to a parsing error (other errors are re-thrown)
	// roll back the token index and return false,
	// if the callback is performed without any error return true
	private try(cb: () => void): boolean {
		let startTokIndex = this.curtokIndex
		try {
			cb()
			return true
		} catch (e) {
			if (!(e instanceof ParsingError)) {
				throw e
			}
			this.curtokIndex = startTokIndex
			return false
		}
	}
	constructor(filename: string, contents: string) {
		this.tokens = lexFile(contents).filter(t =>
			t.type != TokenType.whitespace && 
			t.type != TokenType.comment
		)
		if (this.tokens.length <= 0) {
			console.error("no tokens")
			console.error({tokens: this.tokens})
			process.exit(1)
		}
		this.filename = filename
	}

	/**
	 * Gets the AST representing an sbl file
	 * @returns AST for file
	 */
	public getAST(): SBLFile | never {
		return this.parseSBLFile()
	}

	// get an identifier, or throw an error
	private getIdent(): string {
		// accept Ident and Keyword so you can make aliases with 
		// names that happen to be keywords
		if (this.tok().type === TokenType.Ident ||
			this.tok().type === TokenType.Keyword) {
			let ident = this.tok().content
			this.scanTok()
			return ident
		} else {
			this.unexpectedToken("expected Ident")
		}
	}

	// get a string's content, or throw an error
	private getString(): string {
		if (this.tok().type === TokenType.String) {
			let str = this.tok().content

			// unescape string
			// maybe change this, maybe not
			str = JSON.parse(str)

			this.scanTok()
			return str
		} else {
			this.unexpectedToken("expected String")
		}
	}

	private getJSExecString(): string {
		if (this.tok().type === TokenType.JSExecString) {
			let str = this.tok().content.slice(3, -3)

			// only escapeable char is backtic
			str = str.replace(/\\`/g, "`")

			this.scanTok()
			return str
		} else {
			this.unexpectedToken("expected JSExecString")
		}
	}

	private getArgLiteral(): string {
		if (this.tok().type === TokenType.ArgLiteral) {
			let literal = this.tok().content
			this.scanTok()
			return literal
		} else {
			this.unexpectedToken("expected ArgLiteral")
		}
	}

	private parseSBLFile(): SBLFile | never {
		let sblfile: SBLFile = { Declarations: [] }

		while (this.tok().type !== TokenType.EOF) {
			let decl = this.parseDeclaration()
			sblfile.Declarations.push(decl)
		}

		return sblfile
	}

	private parseDeclaration(): Declaration | never {
		let decl: Partial<Declaration> = { Pos: this.tok().pos }

		if (this.scanKeyword("entry")) {
			decl.Entrypoint = this.getIdent()
		} else {
			decl.Alias = this.parseAlias()
		}

		return decl as Declaration
	}

	private scanKeyword(keyword: string): boolean {
		if (this.tok().type === TokenType.Keyword && 
			this.tok().content === keyword) {
			this.scanTok()
			return true
		}
		return false
	}

	private parseAlias(): Alias | never {
		let alias: Partial<Alias> = {}
		if (!this.scanKeyword("alias")) {
			this.unexpectedToken("expected Keyword \"alias\"")
		}
		alias.Name = this.getIdent()
		if (this.scanKeyword("prefixed")) {
			alias.Keyprefix = this.getString()
		}
		alias.Body = this.parseAliasBody()
		return alias as Alias
	}

	private parseAliasBody(): AliasBody | never {
		let aliasbody: AliasBody = { Actions: [] }
		this.expectations.push({
			keyword: "end",
			message: "expected Keyword \"end\" or an Action"
		})
		while (!this.scanKeyword("end")) {
			let action = this.parseAliasAction()
			aliasbody.Actions.push(action)
		}
		if (this.expectations.pop()?.keyword != "end") {
			this.unexpectedToken("expectation mismatch")
		}

		return aliasbody
	}
	private parseAliasAction(): AliasAction {
		let aliasaction: Partial<AliasAction> = {}

		if (this.try(()=>{
			aliasaction.ExecuteAction = this.parseExecuteAction()
		})) {
			return aliasaction as AliasAction
		}

		if (this.try(()=>{
			aliasaction.GetCompiledAction = this.parseGetCompiledAction();
		})) {
			return aliasaction as AliasAction
		}

		this.unexpectedToken("expected Action or \"end\"")
	}
	private parseExecuteAction(): ExecuteAction | never {
		let execAction: Partial<ExecuteAction> = {}

		let startTokIndex = this.curtokIndex;
		try {
			this.expectations.push({
				keyword: "->",
				message: "expected Keyword \"->\" (previous action must be chained)"
			})
			execAction.RetrieveAction = this.parseRetrieveAction()
		} catch (e) {
			if (!(e instanceof ParsingError)) {
				throw e
			}
			this.curtokIndex = startTokIndex
		}
		if ((execAction.RetrieveAction && !this.scanKeyword("->")) || this.expectations.pop()?.keyword != "->") {
			this.unexpectedToken("expected \"->\"")
		}

		execAction.SimpleAction = this.parseExecuteActionSimple()

		startTokIndex = this.curtokIndex;
		try {
			if (this.scanKeyword("->"))
				execAction.ContinueAction = this.parseContinuedAction()
		} catch (e) {
			if (!(e instanceof ParsingError)) {
				throw e
			}
			this.curtokIndex = startTokIndex
		}

		return execAction as ExecuteAction
	}
	private parseContinuedAction(): ContinuedAction | never {
		let ret: Partial<ContinuedAction> = {}

		if (this.scanKeyword("set")) {
			if (this.scanKeyword("local")) {
				ret.StoreKeyLocal = true
			}
			ret.StoreKey = this.getString()
			return ret as ContinuedAction
		} else {
			ret.NextAction = this.parseExecuteActionSimple()
			if (this.scanKeyword("->")) {
				ret.SecondContinue = this.parseContinuedAction()
			}
			return ret as ContinuedAction
		}
	}
	private parseExecuteActionSimple(): ExecuteActionSimple | never {
		let ret: Partial<ExecuteActionSimple> = {}
		if (this.scanKeyword("js")) {
			ret.JSExec = this.parseJSExecAction()
			return ret as ExecuteActionSimple
		} else if (this.scanKeyword("exec") || this.scanKeyword("pipe")) {
			ret.PipeCommandLiterals = []
			ret.PipeCommandLiterals.push(this.getString())
			while (this.scanKeyword("|")) {
				ret.PipeCommandLiterals.push(this.getString())
			}
			return ret as ExecuteActionSimple
		} else if (this.scanKeyword("say")) {
			ret.UseSayLiteral = true
			ret.SayLiteral = this.getString()
			return ret as ExecuteActionSimple
		} else {
			ret.CallAlias = this.parseCallAliasAction()
			return ret as ExecuteActionSimple
		}
	}
	private parseCallAliasAction(): CallAliasAction | undefined {
		let ret: Partial<CallAliasAction> = {}

		if (!this.scanKeyword("call")) {
			this.unexpectedToken("expected Keyword \"call\"")
		}
		if (this.tok().type === TokenType.User) {
			ret.User = this.tok().content
			this.scanTok()
		}
		ret.AliasName = this.getIdent()
		return ret as CallAliasAction
	}
	private parseJSExecAction(): JSExecAction | undefined {
		let ret: Partial<JSExecAction> = { Pos: this.tok().pos }
		ret.ExecString = this.getJSExecString()
		return ret as JSExecAction
	}
	private parseRetrieveAction(): RetrieveAction | never {
		let retAction: Partial<RetrieveAction> = { Pos: this.tok().pos }
		if (this.scanKeyword("get")) {
			if (this.scanKeyword("local")) {
				retAction.LocalRetrieveKey = true
			}
			retAction.RetrieveKey = this.getString()
			return retAction as RetrieveAction
		} else {
			retAction.RetrieveArgs = this.getArgLiteral()
			return retAction as RetrieveAction
		}
	}
	private parseGetCompiledAction(): GetCompiledAction | never {
		let ret: Partial<GetCompiledAction> = {}
		if (!(this.scanKeyword("get") && this.scanKeyword("compiled"))) {
			this.unexpectedToken("expected Keywords \"get\" \"compiled\"")
		}
		ret.CompilationRoot = this.parseAliasBody()
		if (this.scanKeyword("->")) {
			ret.ContinueAction = this.parseContinuedAction()
		}
		return ret as GetCompiledAction
	}
}