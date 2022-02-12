export function lexFile (contents: string): Token[] {
	let tokens: Token[] = []
	let pos = 0
	let line = 1
	let char = 0
	newtoken: while (pos < contents.length) {
		let startPos = pos
		for (let key of Object.keys(TokenRegexes)) {
			let tokentype = key as TokenType
			const regex = TokenRegexes[tokentype];
			if (regex.test(contents.slice(pos))) {
				let match = contents.slice(pos).match(regex)!
				tokens.push({
					content: match[0],
					pos: { char, line },
					type: tokentype
				})
				let lines = match[0].split("\n")
				// line and char logic might be wrong
				line += lines.length-1
				char = lines.length > 1 ? lines[lines.length-1].length : char + lines[0].length
				pos += match[0].length
				// continue to the start of outer loop to keep token check priority
				continue newtoken
			}
		}
		if (pos === startPos) {
			// nothing matched
			console.log(tokens)
			console.error("Invalid input text", { char, line })
			process.exit(1)
		}
	}
	tokens.push({
		content:'',
		pos: { line, char },
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
	Keyword: new RegExp("^(\\b(alias|end|exec|pipe|prefixed|js|say|get|set|compiled|call|say|entry)\\b|\\||->)"),
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

export type Position = { line: number, char: number }

