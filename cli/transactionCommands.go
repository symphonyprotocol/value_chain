package cli


type TransactionCommand struct {}
func (a *TransactionCommand) Text() string { return "transaction" }
func (a *TransactionCommand) Description() string { return "Transaction related commands" }
func (a *TransactionCommand) Subcommands() []string {
	return []string{
		"new", "list", "show",
	}
}

func (a *TransactionCommand) SupportedArguments() []string { return []string{} }
func (a *TransactionCommand) FollowedBy() []string { return []string{} }
func (a *TransactionCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Warn("transaction need to be followed by commands: new, list or show.")
}
 