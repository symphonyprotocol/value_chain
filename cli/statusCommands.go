package cli

import (
)


type StatusCommand struct {}
func (a *StatusCommand) Text() string { return "status" }
func (a *StatusCommand) Description() string { return "Show current node's status." }
func (a *StatusCommand) Subcommands() []string { return []string{} }
func (a *StatusCommand) SupportedArguments() []string { return []string{} }
func (a *StatusCommand) FollowedBy() []string { return []string{} }
func (a *StatusCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Warn("transaction need to be followed by commands: send, get, list-pending.")
}

var __cmd_inst_status = &StatusCommand{}
