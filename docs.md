# SBL Docs

Since people are asking, I guess I'll write a "tutorial" for SBL.

Language not finalized. (I'll add stuff as I see fit until either I remake it in itself to run on supibot, or stop working on it)

## What even is SBL?

SBL stands for Supibot Language, its a language that compiles to supibot aliases. I created it because I was having a hard time writing aliases in a linear string, and they were getting too long to manage on my own (>1k characters, which was <50 lines in SBL).

Supibot aliases (as defined in the help for `$alias`) let you create your own aliases (shorthands) for any other combination of commands and arguments.

So you write SBL, and the compiler generates an alias that is considered equivelent.

## How to define an alias

To define an alias use the `"alias"` keyword. Blocks are ended with the keyword `"end"`. Inline comments start with `"#"`.

An alias consists of a list of "actions", each action being entirely seperate from the last. There is no direct interaction between actions, but they are executed in order. An example of an action is executing the `$ping` command.

So this is how you would define alias `"xd"`, so that when you do `$$xd` it runs `$ping`.

```ini
alias xd
	# comment
	exec "ping"
end
```

The output of an alias will be the output of its last action, so here `$$xd` will output the same as the ping command.

The example below will never output anything except `"(empty string)"`, but `$ping` will still be executed.

```ini
alias xd2
	# Will this output the response of $ping? No.
	exec "ping"
	exec "abb say"
end
```

### Key prefix

Optionally, a "key prefix" can be defined for the alias. The key prefix will be used by get/set actions using the `"local"` keyword. The syntax used in this example is explained further under "action chains". Key prefixes are prepended at compile time. If its not set, the key prefix will be an empty string.

```ini
alias xd prefixed "my-unique-key-prefix-"
	# Sets the "my-unique-key-prefix-xd" key
	exec "ping" -> set local "xd"
end
```
### Entrypoint

Also, it is possible to define an "entrypoint" for the entire file, this allows you to place multiple aliases inside one file. Although there is no practical reason for doing this at the moment, because you can only compile one alias per file. (that alias being the "entrypoint")

```ini
entry alias2

alias alias1
	exec "ping xd"
end

alias alias2
	exec "abb say xd lmao"
end
```

## Actions

An "action" is similar to using a supibot command in pipe, although actions arent chained (piped) by default.

There are 3 types of actions, normal actions, Post-actions and Pre-actions.

Pre-actions can only be used at the start of an action chain, and Post-Actions can only be used at the end.

### List of actions

* ### `exec` Action (aka `pipe`)

	There are quite a bit of actions availible, the most simple being `exec`, it just takes a string and uses it as a command.
	```ini
	# Outputs "xd"
	alias execExample
		exec "abb say xd"
	end
	```

	Also, exec can be used to pipe commands together, with this feature its possible to mimic most, if not all, supibot aliases 1:1.

	```ini
	# Outputs "xd" but in smol text
	alias pipeCharExample
		pipe "abb say xd" | "tt smol"
	end
	```

	The `exec` and `pipe` keywords are equivelent, meaning anywhere you see `exec` you can put `pipe` instead, and vice versa.

* ### `say` Action

	This action is a shorthand for `exec "abb say"`. Optionally `say` can be used without a string, then it becomes a no-op. (gets optimized out, e.g. `$pipe abb say xd | abb say | abb say | ping` is equivelent to `$pipe abb say xd | ping`)

	```ini
	# Outputs "xd"
	alias sayKeywordExample
		say "xd"
	end
	```

* ### Arg literals Pre-Action

	The arg literals syntax is from supbot, and supports any argument literal that matches this regex: `\${(\d+\+?|-?\d+|-?\d+\.\.(-?\d+)?|\d+-\d+|executor|channel)}`.
	
	Meaning things like this:

		${10}
		${2+}
		${2-4}
		${2..4}
		${-1}
		${-4..-2}
		${executor}
		${channel}

	This alias just outputs its input.

	```ini
	alias say
		${0+} -> say
	end
	```
	More information on arg literals: https://supinic.com/bot/command/detail/alias

* ### `call` Action
	This action is a shorthand for `exec "alias run"` or `exec "alias try"`

	#### Calling someone else's alias

	To call someone else's alias, you use `"call @user alias"`.
	Usernames must be prefixed with the `@` character.

	```ini
	# Runs @xduser's alias "xd"
	alias xd
		call @xduser xd
	end
	```

	#### Calling one of your own aliases
	To call your own alias, use `"call alias"`
	```ini
	# Runs your alias "xd"
	alias xdd
		call xd
	end
	```

* ### `js` Action

	This action is a shorthand for `exec "js function:\" (()=>{ <escaped & minified javascript> })(); \""`

	There is a special syntax for javascript, it uses tripple backtics (<code>"\`\`\`"</code>) as quotes, and the only character that needs escaping is backtic (<code>"\`"</code>), not even backslash needs escaped.

	The text inside the tripple backtics is processed by esbuild (a javascript bundler built in go) with some smol injected runtime functions.

	```ini
	# Outputs the user's username
	alias jsAlias
		js ```
			return executor
		```
	end
	```

	#### Injected runtime

	The sbl js runtime is automatically removed by esbuild if you dont use it or parts of it.

	The runtime consists of 3 functions:

		* getLocal(key)
		* setLocal(key, value)
		* getLocalPrefix()

	`setLocal` and `getLocal` act like `customData.set` and `customData.get` but work with the key prefix defined for the alias.
	`getLocalPrefix` returns the key prefix defined for the alias.

* ### Action chains (`"->"`)

	In SBL, there are "contined actions", "continuations", or "action chains" that allow you to pipe actions into other actions.

	These two aliases compile to the same code: `$pipe ping | abb say`.

	```ini
	alias alias1
		exec "ping"
		-> exec "abb say"
	end
	```
	```ini
	alias alias2
		exec "ping" | "abb say"
	end
	```
	###### Note: Whitespace doesnt have any effect, feel free to format it however you want

	Both of the aliases have one action, but the first one has an "action chain" meaning after the first action finishes, it will then "continue" to another action before it ends. There is no limit to the length of an action chain.

	With this, you can combine the other actions in more useful ways, for example:

	```ini
	alias xdfancy
		call xd -> exec "tt fancy"
	end
	```
* ### `get` Pre-Action

	The `"get"` action is considered a "pre-action" because it cannot be used stand-alone, and must be chained to another command. (this makes sense, because if you are getting a value and do nothing with it, why would you get the value?)

	It could be said that pre-actions must be at the beggining of an action chain, and post-actions (such as set) must be at the end of a chain (meaning it can't continue).

	`"get"` gets a key from `customData` through `$js`.

	```ini
	alias getxdlmao
		get "xd" -> say
	end
	```

	The `"local"` keyword can be added after a get/set of a key to use the key prefix defined for the alias.

	```ini
	alias xdddd prefixed "local-key-prefix-"
		# Key value used is "local-key-prefix-mykeyxd" (prefix and key concatenated)
		get local "mykeyxd" -> say
	end
	```

* ### `set` Post-action
	The `"set"` action can only be used at the end of an action chain, making it a "post-action".

	The set action will set a key in `customData` through `$js`. Currently its only possible to set string values, not other javascript primitives. If you want to use those or unset a value just use javascript and interact with `customData` on your own.

	```ini
	alias xdddd
		# Key value used is "mykeyxd"
		say "xdddd" -> set "mykeyxd"
	end
	```

	The `"local"` keyword can be added after a get/set of a key to use the key prefix defined for the alias.

	```ini
	alias xdddd prefixed "local-key-prefix-"
		# Key value used is "local-key-prefix-mykeyxd" (prefix and key concatenated)
		say "xd" -> set local "mykeyxd"
	end
	```

* ### `get compiled` Pre-Action
	`"get compiled"` will output the compiled string of all actions contained inside it (except argument literals (i.e. `${0+}`) arent allowed, because those can mess up escaping)

	You can think of `"get compiled"` as just being a string that you can pass to `$pipe` to execute the commands inside (because thats what it is).

	This example generates `js function:" 'null | ping' "`, which will always output the string `null | ping`. The `$null` command is added because there is only one command, and `$pipe` requires at least 2 commands to run.

	```ini
	alias getExecPing
		get compiled
			exec "ping"
		end -> say
	end
	```

	This is useful if you want to implement conditional execution of a command as to not unnecessarily envoke a cooldown, or for example editing an alias only if the alias meets some conditions.

	```ini
	alias inspireme prefixed "inspireme-"
		get compiled
			say "Not feeling inspired right now, try again later. :("
		end -> set local "losing-path"
		# set local key "losing-path" to the $pipe input needed to run the code in the block

		get compiled
			exec "inspireme" ->
			say "FeelsGoodMan Feeling very inspired already: "
		end -> set local "winning-path"
		# set local key "winning-path" to the $pipe input needed to run the code in the block

		js ```
			if (Math.random()>0.9) return getLocal("losing-path")
			else return getLocal("winning-path")
		``` -> exec "pipe"
	end
	```
