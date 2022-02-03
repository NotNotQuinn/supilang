// https://gist.github.com/NotNotQuinn/d2a7d308c6686e1e881f7f6fdbff9113
// const tee = [`pipe abb ac 1.. \${0} | null | alias code \${0} | abb tee | null | alias code \${0} | js importGist:d2a7d308c6686e1e881f7f6fdbff9113 function:"\`Filtered through $abb: \\n\${tee[0]}\\n\\nFiltered through $js: \\n\${args.join(' ')}\\n\\nCalculated original definition?:\\n\${calculateAliasDefinition()}\\n\`" | hbp | abb say \${0}:`];
// const args = `pipe abb ac 1.. em:"No input provided! Expected alias name, found nothing." \${0} | null | alias code \${0} | abb tee | null | alias code \${0} | js | hbp | abb say \${0}:`.split(" ");
// const tee = ["abb say function:123"]
// const args = "abb say em:\" test\"".split(" ")

// returns original alias definition
function calculateAliasDefinition() {
	// @ts-ignore
	let abbTxt = tee[tee.length-1]; // filtered through $abb, so all of the params for $abb are missing
	// @ts-ignore
	let jsTxt = args.join(" "); // filtered through $js, so all of the params for $js are missing
	const originalJsTxt = jsTxt
	const originalAbbTxt = abbTxt
	const jsParamNames = ["arguments", "errorInfo", "force", "function", "importGist"]; // https://github.com/Supinic/supibot-package-manager/blob/7cc614369b0a1dc038717858d03e860cde1a56c6/commands/dankdebug/index.js#L8-L14
	const abbParamNames = ["em", "errorMessage", "excludeSelf", "regex", "replacement"]; // https://github.com/Supinic/supibot-package-manager/blob/7cc614369b0a1dc038717858d03e860cde1a56c6/commands/aliasbuildingblock/index.js#L8-L14
	// https://github.com/Supinic/supi-core/blob/13c31df066ca8ca54d9eb3784e853dfbcb628ee5/classes/command.js#L673
	const jsQuotesRegex = new RegExp(`(?<name>${jsParamNames.join("|")}):(?<!\\\\)"(?<value>.*?)(?<!\\\\)"`, "g");
	const abbQuotesRegex = new RegExp(`(?<name>${abbParamNames.join("|")}):(?<!\\\\)"(?<value>.*?)(?<!\\\\)"`, "g");
	// https://github.com/Supinic/supi-core/blob/13c31df066ca8ca54d9eb3784e853dfbcb628ee5/classes/command.js#L714
	const jsQuotelessRegex = new RegExp(`^(?<name>${jsParamNames.join("|")}):(?<value>.*)$`)
	const abbQuotelessRegex = new RegExp(`^(?<name>${abbParamNames.join("|")}):(?<value>.*)$`)
	// console.log({jsTxt, abbTxt})

	// Nothing was consumed
	if (abbTxt === jsTxt) {
		return abbTxt;
	}

	// scan over strings and look for non-matching characters.
	// once we find one that isnt the same on both, then there will be
	// a paramater definition in one of them only at that point
	// we will parse the parameter from the one that contains it, and 
	// copy it to the other at the same point in the string
	// we will keep going until both strings are the same, and we've 
	// reached the end of the strings.
	for (let i = 0; abbTxt !== jsTxt; i++) {
		if (i > abbTxt.length && i > jsTxt.length) break;
		if (abbTxt[i] === jsTxt[i]) {
			// console.log("skip");
			continue;
		};
		// console.log({jsTxt, abbTxt})
		// console.log("index",i)
		let jsHasParam = jsTxt.slice(i).split(" ")[0].includes(":");
		if (jsHasParam) {
			// console.log("ABB -", abbTxt)
			// console.log("     "," ".repeat(i) + "^")
			// copy the JS param to ABB
			let quotedMatches = [...(jsTxt.slice(i)).matchAll(abbQuotesRegex)];
			// console.log("debug12314545", jsTxt.slice(i))
			let quotedMatch = quotedMatches.find(i => i.index === 0)
			// console.log("     "," ".repeat(i) + "^".repeat(( quotedMatch ? quotedMatch[0] : jsTxt.slice(i).split(" ")[0] ).length))
			// console.log("JS  -",jsTxt)
			// console.log({quotedMatch, quotedMatches})
			if (quotedMatch) {
				// console.log({quotedMatch})
				abbTxt = abbTxt.slice(0, i) + quotedMatch[0] + " " + abbTxt.slice(i);
				i+= quotedMatch.length
				continue;
			} else {
				let append = jsTxt.slice(i).split(" ")[0]
				if (abbQuotelessRegex.test(append)) {
					abbTxt = abbTxt.slice(0, i) + append + " " + abbTxt.slice(i);
					i+= append.length
					continue;
				}
			}
		}
		let abbHasParam = abbTxt.slice(i).split(" ")[0].includes(":");
		if (abbHasParam) {
			// console.log("JS  -",jsTxt)
			// console.log("     "," ".repeat(i) + "^")
			// console.log("     "," ".repeat(i) + "^".repeat(abbTxt.slice(i).split(" ")[0].length))
			// console.log("ABB -", abbTxt)
			// copy the ABB param to JS
			let quotedMatches = [...(abbTxt.slice(i)).matchAll(jsQuotesRegex)];
			let quotedMatch = quotedMatches.find(i => i.index === 0)
			// console.log({quotedMatch, quotedMatches})
			if (quotedMatch) {
				jsTxt = abbTxt.slice(0, i) + quotedMatch[0] + " " + jsTxt.slice(i);
				i += quotedMatch[0].length
				continue;
			} else {
				let append = abbTxt.slice(i).split(" ")[0]
				if (jsQuotelessRegex.test(append)) {
					jsTxt = jsTxt.slice(0, i) + append + " " + jsTxt.slice(i);
					i += append.length
					continue;
				}
			}
		}
	}
	// console.log({jsTxt, originalJsTxt, abbTxt, originalAbbTxt})
	return abbTxt;
}

// console.log("Result:", calculateAliasDefinition())

