# Supibot Language

Supibot Language (SBL), is a language that compiles to supibot aliases. I created it because I was having a hard time writing aliases in a linear string, and they were getting too long to manage on my own (>1k characters, which was <50 lines in SBL).

Supibot aliases (as defined in the help for `$alias`) let you create your own aliases (shorthands) for any other combination of commands and arguments.

### Docs

Documentation is in [./docs.md](./docs.md).
### Simple example:

```ini
# xd.sbl

# Executes ping, replacing "Ping" at the start with "pajaDink PING!!!!!"
alias xd
	exec "ping" -> exec "abb replace regex:^Pong replacement:\"pajaDink PING!!!!!\""
end
```

### Smol feature list:

- JS Syntax that doesnt require escaping
- Minified JS
- Localized key prefix for each alias (`set local "key"` vs. `set "key"`)
- Work with blocks of compiled alias as values (for branching execution)
- Automatically insert `errorInfo:true` to all `$js` calls for easier debugging

### VS Code extention

There is a [VSCode extention](https://marketplace.visualstudio.com/items?itemName=QuinnDT.supibot-language-support) that adds syntax highlighting for sbl. [Source code.](https://github.com/notnotquinn/supilang-ext)

