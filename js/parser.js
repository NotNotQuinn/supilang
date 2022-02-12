"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Parser = exports.ParsingError = void 0;
const lexer_1 = require("./lexer");
// Error thrown when parsing couldnt be done correctly,
// NOT all errors from the parser are this type of error
class ParsingError extends Error {
}
exports.ParsingError = ParsingError;
class Parser {
    constructor(filename, contents) {
        this.tokens = [];
        this.expectations = [];
        this.curtokIndex = 0;
        this.tokens = lexer_1.lexFile(contents).filter(t => t.type != lexer_1.TokenType.whitespace &&
            t.type != lexer_1.TokenType.comment);
        if (this.tokens.length <= 0) {
            console.error("no tokens");
            console.error({ tokens: this.tokens });
            process.exit(1);
        }
        this.filename = filename;
    }
    // current token
    tok() {
        return this.tokens[this.curtokIndex];
    }
    // move past the current token
    scanTok() {
        this.curtokIndex++;
        if (this.curtokIndex > this.tokens.length) {
            throw new Error("token overflow");
        }
    }
    // raise an error because the current token is unexpected
    unexpectedToken(msg = "") {
        console.error("Unexpected token: " + msg, this.tok());
        if (this.tok().type === lexer_1.TokenType.EOF) {
            if (this.expectations.length > 0) {
                msg = this.expectations[this.expectations.length - 1].message;
            }
            throw new ParsingError("unexpected EOF: " + msg);
        }
        throw new ParsingError("unexpected token: " + msg);
    }
    // Try the callback, if it fails due to a parsing error (other errors are re-thrown)
    // roll back the token index and return false,
    // if the callback is performed without any error return true
    try(cb) {
        let startTokIndex = this.curtokIndex;
        try {
            cb();
            return true;
        }
        catch (e) {
            if (!(e instanceof ParsingError)) {
                throw e;
            }
            this.curtokIndex = startTokIndex;
            return false;
        }
    }
    /**
     * Gets the AST representing an sbl file
     * @returns AST for file
     */
    getAST() {
        return this.parseSBLFile();
    }
    // get an identifier, or throw an error
    getIdent() {
        // accept Ident and Keyword so you can make aliases with 
        // names that happen to be keywords
        if (this.tok().type === lexer_1.TokenType.Ident ||
            this.tok().type === lexer_1.TokenType.Keyword) {
            let ident = this.tok().content;
            this.scanTok();
            return ident;
        }
        else {
            this.unexpectedToken("expected Ident");
        }
    }
    // get a string's content, or throw an error
    getString() {
        if (this.tok().type === lexer_1.TokenType.String) {
            let str = this.tok().content;
            // unescape string
            // maybe change this, maybe not
            str = JSON.parse(str);
            this.scanTok();
            return str;
        }
        else {
            this.unexpectedToken("expected String");
        }
    }
    getJSExecString() {
        if (this.tok().type === lexer_1.TokenType.JSExecString) {
            let str = this.tok().content.slice(3, -3);
            // only escapeable char is backtic
            str = str.replace(/\\`/g, "`");
            this.scanTok();
            return str;
        }
        else {
            this.unexpectedToken("expected JSExecString");
        }
    }
    getArgLiteral() {
        if (this.tok().type === lexer_1.TokenType.ArgLiteral) {
            let literal = this.tok().content;
            this.scanTok();
            return literal;
        }
        else {
            this.unexpectedToken("expected ArgLiteral");
        }
    }
    parseSBLFile() {
        let sblfile = { Declarations: [] };
        while (this.tok().type !== lexer_1.TokenType.EOF) {
            let decl = this.parseDeclaration();
            sblfile.Declarations.push(decl);
        }
        return sblfile;
    }
    parseDeclaration() {
        let decl = { Pos: this.tok().pos };
        if (this.scanKeyword("entry")) {
            decl.Entrypoint = this.getIdent();
        }
        else {
            decl.Alias = this.parseAlias();
        }
        return decl;
    }
    scanKeyword(keyword) {
        if (this.tok().type === lexer_1.TokenType.Keyword &&
            this.tok().content === keyword) {
            this.scanTok();
            return true;
        }
        return false;
    }
    parseAlias() {
        let alias = {};
        if (!this.scanKeyword("alias")) {
            this.unexpectedToken("expected Keyword \"alias\"");
        }
        alias.Name = this.getIdent();
        if (this.scanKeyword("prefixed")) {
            alias.Keyprefix = this.getString();
        }
        alias.Body = this.parseAliasBody();
        return alias;
    }
    parseAliasBody() {
        var _a;
        let aliasbody = { Actions: [] };
        this.expectations.push({
            keyword: "end",
            message: "expected Keyword \"end\" or an Action"
        });
        while (!this.scanKeyword("end")) {
            let action = this.parseAliasAction();
            aliasbody.Actions.push(action);
        }
        if (((_a = this.expectations.pop()) === null || _a === void 0 ? void 0 : _a.keyword) != "end") {
            this.unexpectedToken("expectation mismatch");
        }
        return aliasbody;
    }
    parseAliasAction() {
        let aliasaction = {};
        if (this.try(() => {
            aliasaction.ExecuteAction = this.parseExecuteAction();
        })) {
            return aliasaction;
        }
        if (this.try(() => {
            aliasaction.GetCompiledAction = this.parseGetCompiledAction();
        })) {
            return aliasaction;
        }
        this.unexpectedToken("expected Action or \"end\"");
    }
    parseExecuteAction() {
        var _a;
        let execAction = {};
        let startTokIndex = this.curtokIndex;
        try {
            this.expectations.push({
                keyword: "->",
                message: "expected Keyword \"->\" (previous action must be chained)"
            });
            execAction.RetrieveAction = this.parseRetrieveAction();
        }
        catch (e) {
            if (!(e instanceof ParsingError)) {
                throw e;
            }
            this.curtokIndex = startTokIndex;
        }
        if ((execAction.RetrieveAction && !this.scanKeyword("->")) || ((_a = this.expectations.pop()) === null || _a === void 0 ? void 0 : _a.keyword) != "->") {
            this.unexpectedToken("expected \"->\"");
        }
        execAction.SimpleAction = this.parseExecuteActionSimple();
        startTokIndex = this.curtokIndex;
        try {
            if (this.scanKeyword("->"))
                execAction.ContinueAction = this.parseContinuedAction();
        }
        catch (e) {
            if (!(e instanceof ParsingError)) {
                throw e;
            }
            this.curtokIndex = startTokIndex;
        }
        return execAction;
    }
    parseContinuedAction() {
        let ret = {};
        if (this.scanKeyword("set")) {
            if (this.scanKeyword("local")) {
                ret.StoreKeyLocal = true;
            }
            ret.StoreKey = this.getString();
            return ret;
        }
        else {
            ret.NextAction = this.parseExecuteActionSimple();
            if (this.scanKeyword("->")) {
                ret.SecondContinue = this.parseContinuedAction();
            }
            return ret;
        }
    }
    parseExecuteActionSimple() {
        let ret = {};
        if (this.scanKeyword("js")) {
            ret.JSExec = this.parseJSExecAction();
            return ret;
        }
        else if (this.scanKeyword("exec") || this.scanKeyword("pipe")) {
            ret.PipeCommandLiterals = [];
            ret.PipeCommandLiterals.push(this.getString());
            while (this.scanKeyword("|")) {
                ret.PipeCommandLiterals.push(this.getString());
            }
            return ret;
        }
        else if (this.scanKeyword("say")) {
            ret.UseSayLiteral = true;
            ret.SayLiteral = this.getString();
            return ret;
        }
        else {
            ret.CallAlias = this.parseCallAliasAction();
            return ret;
        }
    }
    parseCallAliasAction() {
        let ret = {};
        if (!this.scanKeyword("call")) {
            this.unexpectedToken("expected Keyword \"call\"");
        }
        if (this.tok().type === lexer_1.TokenType.User) {
            ret.User = this.tok().content;
            this.scanTok();
        }
        ret.AliasName = this.getIdent();
        return ret;
    }
    parseJSExecAction() {
        let ret = { Pos: this.tok().pos };
        ret.ExecString = this.getJSExecString();
        return ret;
    }
    parseRetrieveAction() {
        let retAction = { Pos: this.tok().pos };
        if (this.scanKeyword("get")) {
            if (this.scanKeyword("local")) {
                retAction.LocalRetrieveKey = true;
            }
            retAction.RetrieveKey = this.getString();
            return retAction;
        }
        else {
            retAction.RetrieveArgs = this.getArgLiteral();
            return retAction;
        }
    }
    parseGetCompiledAction() {
        let ret = {};
        if (!(this.scanKeyword("get") && this.scanKeyword("compiled"))) {
            this.unexpectedToken("expected Keywords \"get\" \"compiled\"");
        }
        ret.CompilationRoot = this.parseAliasBody();
        if (this.scanKeyword("->")) {
            ret.ContinueAction = this.parseContinuedAction();
        }
        return ret;
    }
}
exports.Parser = Parser;
