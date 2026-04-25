package bus

// Event Kind constants. Stable strings so subscribers can filter without
// importing producer types.
const (
	KindBindingSet    = "binding.set"
	KindBindingUpdate = "binding.update"
	KindThemeSwitch   = "theme.switch"
	KindNavPush       = "nav.push"
	KindNavPop        = "nav.pop"
	KindNavReplace    = "nav.replace"
	KindLoaderStart   = "loader.start"
	KindLoaderSuccess = "loader.success"
	KindLoaderError   = "loader.error"
	KindLoaderCancel  = "loader.cancel"
	KindInput         = "input.key"
	KindEffectMsg     = "effect.msg"
)

// Source identifiers.
const (
	SourceBinding = "binding"
	SourceTheme   = "theme"
	SourceNav     = "nav"
	SourceAsync   = "async"
	SourceInput   = "input"
	SourceEffect  = "effect"
)

// BindingChange is the payload for KindBindingSet / KindBindingUpdate.
// Old and New are stored as any; subscribers must type-assert if they want details.
type BindingChange struct {
	Name string // optional binding label, may be empty
	Old  any
	New  any
}

// ThemeSwitch is the payload for KindThemeSwitch.
type ThemeSwitch struct {
	From string
	To   string
}

// PageNav is the payload for KindNavPush / KindNavPop / KindNavReplace.
type PageNav struct {
	Op    string // "push" | "pop" | "replace"
	Name  string
	Depth int
}

// LoaderState is the payload for KindLoader* events.
type LoaderState struct {
	Name  string
	Phase string // "start" | "success" | "error" | "cancel"
	Err   error  // populated on phase == "error"
}

// InputEvent is the payload for KindInput.
type InputEvent struct {
	Key      string
	Consumed bool
}

// EffectMsg is the payload for KindEffectMsg.
type EffectMsg struct {
	Msg any
}
