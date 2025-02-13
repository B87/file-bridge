package filesystem

type Config struct {
	Name    string
	Type    string
	Default bool
	Config  map[string]interface{}
}
