package briefkitctl

type AgentCmd struct {
	List AgentListCmd `cmd:"" help:"List configured agents"`
	Add  AgentAddCmd  `cmd:"" help:"Add new agent"`
}
