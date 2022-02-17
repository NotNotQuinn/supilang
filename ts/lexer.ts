export function lexFile (filename: string, contents: string): Token[] {
	let tokens: Token[] = []
	let pos = 0
	let line = 1
	let col = 1
	newtoken: while (pos < contents.length) {
		let startPos = pos
		for (let key of Object.keys(TokenRegexes)) {
			let tokentype = key as TokenType
			const regex = TokenRegexes[tokentype];
			if (regex.test(contents.slice(pos))) {
				let match = contents.slice(pos).match(regex)!
				tokens.push({
					content: match[0],
					pos: { col, line, filename },
					type: tokentype
				})
				let lines = match[0].split("\n")
				line += lines.length-1
				col = lines.length > 1 ? lines[lines.length-1].length+1 : col + lines[0].length
				pos += match[0].length
				// continue to the start of outer loop to keep token check priority
				continue newtoken
			}
		}
		if (pos === startPos) {
			// nothing matched
			console.log(tokens)
			if (SUPIBOT) {
				throw new Error("Invalid input text: "+filename+":"+line+":"+col)
			}
			console.error("Invalid input text", { char: col, line })
			process.exit(1)
		}
	}
	tokens.push({
		content: '',
		pos: { line, col: col, filename },
		type: TokenType.EOF
	})
	return tokens
}

export enum TokenType {
	"Ident" = "Ident",
	"Keyword" = "Keyword",
	"User" = "User",
	"ArgLiteral" = "ArgLiteral",
	"JSExecString" = "JSExecString",
	"String" = "String",
	"comment" = "comment",
	"whitespace" = "whitespace",
	"EOF" = "EOF",
}

const TokenRegexes: { [key in keyof typeof TokenType]: RegExp }  = {
	// Tokens are prioritized in this order
	// TODO: add "temp" keyword (when implementing temp keys)
	Keyword: new RegExp("^(\\b(alias|import|local|end|exec|pipe|prefixed|js|say|get|set|compiled|call|say|entry)\\b|\\||->)"),
	Ident: new RegExp("^([-a-zA-Z_0-9]{2,30})"),
	User: new RegExp("^(@[-a-zA-Z_0-9]+)"),
	ArgLiteral: new RegExp("^(\\${(\\d+\\+?|-?\\d+|-?\\d+\\.\\.(-?\\d+)?|\\d+-\\d+|executor|channel)})"),
	JSExecString: new RegExp("^((\\x60{3})(?:\\\\.|[^\\x60])*(\\x60{3}))"),
	String: new RegExp("^(\"(?:\\\\.|[^\"])*\")"),
	comment: new RegExp("^(#[^\\n\\r]*)"),
	whitespace: new RegExp("^(\\s+)"),
	EOF: new RegExp("$^") // will never match
}

export type Token = {
	pos: Position
	type: TokenType
	content: string
}

export type Position = { line: number, col: number, filename: string }

