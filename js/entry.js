#!/usr/bin/env node
"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
// entry for debugging
const fs = __importStar(require("fs"));
const parser_1 = require("./parser");
const util_1 = require("util");
function main() {
    let filename;
    if (process.argv.length > 2) {
        filename = process.argv[2];
    }
    else {
        console.log("Usage: " + process.argv.slice(0, 2).join(" ") + " file");
        process.exit(1);
    }
    let contents = fs.readFileSync(filename).toString('utf-8');
    let parser = new parser_1.Parser(filename, contents);
    let fileAST = parser.getAST();
    console.log(util_1.inspect(fileAST, false, null, true));
    // let compiled = fileAST.Compile()
    // console.log(compiled);
    // fs.writeFileSync("out.alias", compiled[0] || "")
}
main();
