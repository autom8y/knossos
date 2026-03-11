package compiler

type ChannelCompiler interface {
	// CompileCommand transforms mena dromenon content for the target channel.
	// Returns (filename, content, error).
	CompileCommand(name, description, argHint, body string) (string, []byte, error)

	// CompileSkill transforms mena legomenon content for the target channel.
	// Returns (dirName, filename, content, error).
	CompileSkill(name, description, body string) (string, string, []byte, error)

	// CompileAgent transforms agent content for the target channel.
	CompileAgent(name string, frontmatter map[string]any, body string) ([]byte, error)

	// ContextFilename returns the context file name (CLAUDE.md or GEMINI.md).
	ContextFilename() string
}
