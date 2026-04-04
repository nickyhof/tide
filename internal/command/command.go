package command

// Command represents an editor command that can be executed.
type Command struct {
	Name        string
	Description string
	Shortcut    string // display string for help, e.g. "F2"
	Execute     func()
}

// Registry stores all available commands.
type Registry struct {
	commands map[string]Command
	order    []string // insertion order for stable listing
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd Command) {
	if _, exists := r.commands[cmd.Name]; !exists {
		r.order = append(r.order, cmd.Name)
	}
	r.commands[cmd.Name] = cmd
}

// Execute runs a command by name, returning false if not found.
func (r *Registry) Execute(name string) bool {
	cmd, ok := r.commands[name]
	if !ok {
		return false
	}
	cmd.Execute()
	return true
}

// List returns all registered commands in insertion order.
func (r *Registry) List() []Command {
	cmds := make([]Command, 0, len(r.order))
	for _, name := range r.order {
		cmds = append(cmds, r.commands[name])
	}
	return cmds
}

// Search returns commands whose name or description contains the query (case-insensitive).
func (r *Registry) Search(query string) []Command {
	if query == "" {
		return r.List()
	}
	query = toLower(query)
	var results []Command
	for _, name := range r.order {
		cmd := r.commands[name]
		if contains(toLower(cmd.Name), query) || contains(toLower(cmd.Description), query) {
			results = append(results, cmd)
		}
	}
	return results
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
