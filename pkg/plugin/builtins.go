package plugin

// Builtin is a built-in plugin for Draft, installed by default.
type Builtin string

// Builtins fetches all built-in plugins.
func Builtins() []Builtin {
	return []Builtin{"pack-repo"}
}
