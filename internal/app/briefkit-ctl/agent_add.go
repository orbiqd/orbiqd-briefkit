package briefkitctl

import (
	"errors"
	"fmt"
)

type AgentAddCmd struct {
	ID   string `arg:"" required:"" help:"Agent ID"`
	Kind string `arg:"" required:"" help:"Agent kind (codex|claude|gemini)"`
	Path string `arg:"" required:"" help:"Executable path"`
}

func (a *AgentAddCmd) Run() error {
	fmt.Println("Agent add - not implemented")
	return errors.New("not implemented")
}
