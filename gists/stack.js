// https://gist.github.com/NotNotQuinn/3016289a81943e3c07160080a94569f1
/** Push to stack identified by `id` */
function stackPush(id, item) {
	let count = customData.get("stack-"+id+"-count") ?? 0
	customData.set("stack-"+id+"-"+utils.zf(count, 4), item);
	customData.set("stack-"+id+"-count", count+1);
}
/** Pop from stack identified by `id`, returning undefined if not found */
function stackPop(id) {
	let count = customData.get("stack-"+id+"-count") ?? 0
	if (count < 1) return undefined;
	let value = customData.get("stack-"+id+"-"+utils.zf(count-1, 4))
	// remove popped item, and lower count. If count is 0 remove it too
	customData.set("stack-"+id+"-"+utils.zf(count-1, 4), undefined)
	customData.set("stack-"+id+"-count", count-1 || undefined)
	return value;
}
/** Delete stack identified by `id` */
function stackDelete(id) {
	// clear stack data
	for (let key of customData.getKeys()) {
		// clears "stack-id-xxxx"
		// and "stack-id-count"
		if ((key.startsWith("stack-"+id+"-") && key.length === 6+id.length+1+4) || key === "stack-"+id+"-count")
			customData.set(key, undefined)
	}
}
