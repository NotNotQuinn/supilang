import type { Position } from './lexer';
// Compileable 

// options passed to compile
export type AliasOptions = {
	Aliasname          :string
	Keyprefix          :string
	ForcePipeCommand   :boolean
	JSForceErrorInfo   :boolean
	DisallowArgLiteral :boolean
	MinifyJS           :boolean
}

// implemented
export type SBLFile = {
	Declarations: Declaration[]
}

// implemented
export type Declaration = {
	Pos: Position
	Entrypoint: string 
} & {
	Pos: Position
	Alias: Alias
}

// implemented
export type Alias = {
	Pos: Position
	Name: string
	Keyprefix?: string
	Body: AliasBody
}

// implemented
export type AliasBody = {
	Pos: Position
	Actions: AliasAction[]
}

// implemented
export type AliasAction = {
	Pos: Position
	ExecuteAction?: ExecuteAction
} & {
	Pos: Position
	GetCompiledAction?: GetCompiledAction
}

// Execute a command, storing the output for later use
// implemented
export type ExecuteAction = {
	Pos: Position
	RetrieveAction?: RetrieveAction
	SimpleAction: ExecuteActionSimple
	ContinueAction?: ContinuedAction
}

// implemented
export type RetrieveAction = {
	Pos: Position
	LocalRetrieveKey?: boolean
	RetrieveKey?: string
} & {
	Pos: Position
	RetrieveArgs?: string
}

// implemented
export type ContinuedAction = {
	Pos: Position
	StoreKeyLocal?: boolean
	StoreKey?: string
} & {
	Pos: Position
	NextAction?: ExecuteActionSimple
	SecondContinue?: ContinuedAction
}

// Execute a command, voiding the output
// implemented
export type ExecuteActionSimple = {
	// ?Pos                 lexer.Position
	Pos: Position
	JSExec?: JSExecAction
} & {
	// ?Pos                 lexer.Position
	Pos: Position
	PipeCommandLiterals?: string[]
} & {
	// ?Pos                 lexer.Position
	Pos: Position
	UseSayLiteral?: true
	SayLiteral?: string
} & {
	// ?Pos                 lexer.Position
	Pos: Position
	CallAlias?: CallAliasAction
}

// implemented
export type JSExecAction = {
	Pos: Position
	ImportedGist?: string
	ExecString: JSExecString
}

export type JSExecString = {
	Pos: Position
	RawString: string
}

export type GetCompiledAction = {
	Pos: Position
	CompilationRoot: AliasBody
	ContinueAction?: ContinuedAction
}

// implemented
export type CallAliasAction = {
	Pos: Position
	User?: string
	AliasName: string
}
