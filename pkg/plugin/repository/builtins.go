package repository

// DefaultPluginRepository is the default Draft core plugins.
const DefaultPluginRepository = "github.com/bacongobbler/draft-plugins"

// Builtin is a built-in plugin repository for Draft, installed by default.
type Builtin string

// Builtins fetches all built-in plugins.
func Builtins() []Builtin {
	return []Builtin{"https://" + DefaultPluginRepository}
}
