export {}

declare global {
	const VMError: (...args: Array<unknown>) => unknown 
	const Buffer: {
		(...args: Array<unknown>): unknown,
		poolSize: number,
		from: (...args: Array<unknown>) => unknown
		of: (...args: Array<unknown>) => unknown
		alloc: (...args: Array<unknown>) => unknown
		allocUnsafe: (...args: Array<unknown>) => unknown
		allocUnsafeSlow: (...args: Array<unknown>) => unknown
		isBuffer: (...args: Array<unknown>) => unknown
		compare: (...args: Array<unknown>) => unknown
		isEncoding: (...args: Array<unknown>) => unknown
		concat: (...args: Array<unknown>) => unknown
		byteLength: (...args: Array<unknown>) => unknown
	}
	var Function: typeof Function;
	// @ts-expect-error
	const eval: () => never;
	const aliasStack: Array<string>;
	const args: Array<string> | null;
	const channel: string;
	interface newConsole extends Omit<Console, string> {}
	// @ts-expect-error
	var console: newConsole
	const executor: string;
	const platform: string;
	const tee: Array<string>
	/**
	 * Object used to access customDeveloperData keys
	 * 
	 * Keys set are linked to each user.
	 */
	const customData: {
		/**
		 * Returns an array of all keys on customData
		 */
		getKeys (): Array<string>
		/**
		 * Set key to the value provided
		 * @param key Key to set
		 * @param value Value to assign
		 */
		set (key: string, value: string | number | boolean | undefined | null): undefined
		/**
		 * Get the value stored at key
		 * @param key Key to access
		 */
		get (key: string): string | number | boolean | undefined | null
	}
	/**
	 * Util functions provided by supibot.
	 */
	const utils: Utils 
	interface Utils {
		/**
		 * Get's the best availible emote for the current context
		 * @param array Array of emotes to check, in order
		 * @param fallback Fallback to use if none of the emotes are availible
		 */
		getEmote (array: Array<string>, fallback: string): string
		capitalize (string: string): string
		/**
		 * Picks a random item from an array
		 * @param arr Array to pick from
		 */
		randArray<T> (arr: Array<T>): T
		/**
		 * Random int from min to max
		 * @param min at least -9007199254740992
		 * @param max at most 9007199254740992
		 */
		random (min: number, max: number): number
		randomString (length: number, characters?: string): string
		removeAccents (string: string): string
		timeDelta (target: Date, skipAffixes?: boolean, respectLeapYears?: boolean, deltaTo?: Date): string
		wrapString (string: string, length: number, options?: {keepWhitespace: boolean}): string
		/**
		 * Stringifys a number with an amount of padding
		 * @param number number to stringify
		 * @param padding amount of padding
		 */
		zf (number: number, padding: number): string
	}
}
