// entry for debugging
import * as fs from "fs";
import { Parser } from './parser';
import { inspect } from 'util'
import { Compile } from './compile';
// @ts-ignore not running inside supibot
globalThis["SUPIBOT"] = false
function main() {
	let filename: string;
	if (process.argv.length > 2) {
		filename = process.argv[2]
	} else {
		console.log("Usage: "+process.argv.slice(0, 2).join(" ")+" file")
		process.exit(1)
	}
	let contents = fs.readFileSync(filename).toString('utf-8')
	let parser = new Parser(filename, contents)
	let fileAST = parser.getAST()
	console.log(inspect(fileAST, false, null, true))
	let compiled = Compile(fileAST)
	console.log(compiled);
	fs.writeFileSync("out.alias.2", compiled)
}
main()


