// https://gist.github.com/NotNotQuinn/a75a4b8f709aa56d89702e6601122f38
// PROBLEM: WHAT IF YOU USE SPPC INSIDE SPPC???
// I am going to back to the drawing board on this one

// spp = Supibot Procedure Protocol
const spp = {args: [], sessionId: "uninitialized-session-id"}

// bad uuid doesnt follow spec
function baduuid() {
	function uuidv4() { return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) { var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8); return v.toString(16); }); }
	return uuidv4()
}
// initialize runtime
function sppInit() {
	let isInitialized = "__isInitialized"
	if (spp[isInitialized]) return;
	spp[isInitialized] = true;
	spp["sessionId"] = baduuid()
	customData.set("spp-last-id", customData.get("spp-current-id"))
	customData.set("spp-current-id", spp["sessionId"])
	// load arguments
	let debug = customData.get("spp-debug")
	const keys = customData.getKeys().sort((a,b) => a.localeCompare(b));
	for (let key of keys) {
		
		if (key.startsWith("tmp-spp-arg-")) {
			customData.set("spp-debug-id-"+spp.sessionId+"-"+key.slice(8), customData.get(key))
			if (parseInt(key.slice(12)) < customData.get("tmp-spp-arg-count")) {
				spp.args.push(customData.get(key))
				customData.set(key, undefined)
			}
		}
	}
}

function loadSpp() {
	let id = customData.get("spp-current-id")
	spp.sessionId = id;
	spp.args = [];
	spp.__isInitialized = true;
	// load arguments
	const keys = customData.getKeys().sort((a,b) => a.localeCompare(b));
	let prefix = "spp-debug-id-"+spp.sessionId+"-arg-"
	for (let key of keys) {
		if (key.startsWith(prefix)) {
			
			if (parseInt(key.slice(12)) < customData.get(prefix+"count") && (key != prefix+"count")) {
				spp.args.push(customData.get(key))
			}
		}
	}
}

function sppa(string) {
	let count = customData.get("tmp-spp-arg-count") ?? 0
	customData.set("tmp-spp-arg-"+utils.zf(count, 3), string);
	customData.set("tmp-spp-arg-count", count+1);
	return '';
}
// sppc = spp Call
function sppcSetup() {
	sppInit();
	return '';
}
function sppcTeardown() {
	// clear tmp arguments
	for (let key of customData.getKeys()) {
		// clears "tmp-spp-arg-xxx"
		// and "tmp-spp-arg-count"
		if (key.startsWith("tmp-spp-arg-") || (!customData.get("spp-debug") && key.startsWith("spp-debug-id-"+spp.sessionId+"-"))) customData.set(key, undefined)
	}
	return '';
}
function resetSpp() {
	spp.args = [];
	spp.sessionId = "uninitialized-session-id"
	spp.__isInitialized = false;
	// {args: [], sessionId: "uninitialized-session-id"}
}
function loadtestdata() {
	const customDeveloperData = {
		"0111100001100100-aliasname": "test",
		"0111100001100100-log": " testxd testxdd testxddd DuckerZ 1-100 markov eval backupAlias-debug funfact_eval",
		"0111100001100100-overwrite-target": "test",
		"0111100001100100-passthrough": false,
		"alias-backup-backupAlias": "pipe abb ac 1.. em:\"No input provided! Expected alias name, found nothing.\" ${0} | null | alias code ${0} | abb tee | null | alias code ${0} | js importGist:d2a7d308c6686e1e881f7f6fdbff9113 function:\"customData.set('alias-backup-${0}', calculateAliasDefinition())\"",
		"alias-backup-test": "pipe abb say this is some text, em:\"asdsd\" asd:xd function:\"urmom lol\" | abb say",
		"alias-backup-testxdd": "pipe rw | abb tee | urban | js",
		"alias-log": " logAliases force pipe 0111100001100100 testxd testxddd DuckerZ backupAlias-debug",
		"alias-whitelist": "[\"0111100001100100\"]",

		"tmp-spp-arg-007": "7",
		"tmp-spp-arg-001": "1",
		"tmp-spp-arg-008": "8",
		"tmp-spp-arg-003": "3",
		"tmp-spp-arg-006": "6",
		"tmp-spp-arg-005": "5",
		"tmp-spp-arg-009": "9",
		"tmp-spp-arg-002": "2",
		"tmp-spp-arg-000": "0",
		"tmp-spp-arg-004": "4",
		"tmp-spp-arg-010": "10",
		"tmp-spp-arg-count": 100,
		"spp-debug": true,
  }
	globalThis["customData"] = {
		getKeys: () => Object.keys(customDeveloperData),
		set: (key, value) => {
			if (typeof key !== "string") {
				throw new Error("Only strings are available as keys");
			}
			else if (value && (typeof value === "object" || typeof value === "function")) {
				throw new Error("Only primitives are accepted as object values");
			}

			customDeveloperData[key] = value;
		},
		get: (key) => (Object.hasOwn(customDeveloperData, key))
			? customDeveloperData[key]
			: undefined
	}
	globalThis["args"] = []
	globalThis["tee"] = []
	globalThis["utils"] = {
		zf: (number, padding) => {
			("0".repeat(padding) + number).slice(-padding)
		}
	}
	console.log(customDeveloperData)
	sppInit();
	console.log(spp);
	console.log(JSON.parse(JSON.stringify(customDeveloperData)));
	resetSpp();
	loadSpp();
	console.log(spp);
	resetSpp();
	sppInit();
	console.log(JSON.parse(JSON.stringify(customDeveloperData)));
	console.log(spp)
	sppcTeardown();
}
loadtestdata();