import { Position } from './lexer';
// Compileable 

// options passed to compile
type AliasOptions = {
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
	Name: string
	Keyprefix?: string
	Body: AliasBody
}

// implemented
export type AliasBody = {
	Actions: AliasAction[]
}

// implemented
export type AliasAction = {
	ExecuteAction: ExecuteAction
} & {
	GetCompiledAction: GetCompiledAction
}

// Execute a command, storing the output for later use
// implemented
export type ExecuteAction = {
	RetrieveAction?: RetrieveAction
	SimpleAction: ExecuteActionSimple
	ContinueAction?: ContinuedAction
}

// implemented
export type RetrieveAction = {
	Pos: Position
	LocalRetrieveKey: boolean
	RetrieveKey: string
} & {
	RetrieveArgs: string
}

// implemented
export type ContinuedAction = {
	StoreKeyLocal: boolean
	StoreKey: string
} & {
	NextAction: ExecuteActionSimple
	SecondContinue?: ContinuedAction
}

// Execute a command, voiding the output
// implemented
export type ExecuteActionSimple = {
	// ?Pos                 lexer.Position
	JSExec: JSExecAction
} & {
	// ?Pos                 lexer.Position
	PipeCommandLiterals: string[]
} & {
	// ?Pos                 lexer.Position
	UseSayLiteral: true
	SayLiteral?: string
} & {
	// ?Pos                 lexer.Position
	CallAlias: CallAliasAction
}

// implemented
export type JSExecAction = {
	Pos: Position
	ExecString: string
}

export type GetCompiledAction = {
	CompilationRoot: AliasBody
	ContinueAction?: ContinuedAction
}

// implemented
export type CallAliasAction = {
	User?: string
	AliasName: string
}
