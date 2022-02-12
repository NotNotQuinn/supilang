"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TokenType = exports.lexFile = void 0;
function lexFile(contents) {
    //	var aliasLexer = lexer.MustSimple([]lexer.Rule{
    //		// identifiers can "overwrite" keywords, otherwise keywords are priorotized
    //		{`Ident`, `[-a-zA-Z_0-9]{2,30}`, nil},
    //		{`Keyword`, `alias|end|\||exec|pipe|->|prefixed|say|get|compiled|call|say|entry`, nil},
    //		{`User`, `@[-a-zA-Z_0-9]*`, nil},
    //		{`ArgLiteral`, `\${(\d+\+?|-?\d+|-?\d+\.\.(-?\d+)?|\d+-\d+|executor|channel)}`, nil},
    //		{`JSExecString`, `(\x60{3})(?:\\.|[^\x60])*(\x60{3})`, nil},
    //		// {`Word`, `[a-zA-Z_][a-zA-Z0-9_]`, nil},
    //		{`String`, `"(?:\\.|[^"])*"`, nil},
    //		{"comment", `#[^\n]*`, nil},
    //		{"whitespace", `\s+`, nil},
    //	})
    let tokens = [];
    let pos = 0;
    let line = 1;
    let char = 0;
    newtoken: while (pos < contents.length) {
        let startPos = pos;
        for (let key of Object.keys(TokenRegexes)) {
            let tokentype = key;
            const regex = TokenRegexes[tokentype];
            if (regex.test(contents.slice(pos))) {
                let match = contents.slice(pos).match(regex);
                tokens.push({
                    content: match[0],
                    pos: { char, line },
                    type: tokentype
                });
                let lines = match[0].split("\n");
                line += lines.length - 1;
                char = lines.length > 1 ? lines[lines.length - 1].length : char + lines[0].length;
                pos += match[0].length;
                // continue to the start of outer loop to keep token check priority
                continue newtoken;
            }
        }
        if (pos === startPos) {
            // nothing matched
            console.log(tokens);
            console.error("Invalid input text", { char, line });
            process.exit(1);
        }
    }
    tokens.push({
        content: '',
        pos: { line, char },
        type: TokenType.EOF
    });
    return tokens;
}
exports.lexFile = lexFile;
var TokenType;
(function (TokenType) {
    TokenType["Ident"] = "Ident";
    TokenType["Keyword"] = "Keyword";
    TokenType["User"] = "User";
    TokenType["ArgLiteral"] = "ArgLiteral";
    TokenType["JSExecString"] = "JSExecString";
    TokenType["String"] = "String";
    TokenType["comment"] = "comment";
    TokenType["whitespace"] = "whitespace";
    TokenType["EOF"] = "EOF";
})(TokenType = exports.TokenType || (exports.TokenType = {}));
const TokenRegexes = {
    // Tokens are prioritized in this order
    Keyword: new RegExp("^(\\b(alias|end|\\||exec|pipe|prefixed|js|say|get|set|compiled|call|say|entry)\\b|->)"),
    Ident: new RegExp("^([-a-zA-Z_0-9]{2,30})"),
    User: new RegExp("^(@[-a-zA-Z_0-9]*)"),
    ArgLiteral: new RegExp("^(\\${(\\d+\\+?|-?\\d+|-?\\d+\\.\\.(-?\\d+)?|\\d+-\\d+|executor|channel)})"),
    JSExecString: new RegExp("^((\\x60{3})(?:\\\\.|[^\\x60])*(\\x60{3}))"),
    String: new RegExp("^(\"(?:\\\\.|[^\"])*\")"),
    comment: new RegExp("^(#[^\\n\\r]*)"),
    whitespace: new RegExp("^(\\s+)"),
    EOF: new RegExp("$^") // will never match
};
