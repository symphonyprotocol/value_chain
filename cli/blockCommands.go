package cli

import (
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/value_chain/node"
)

type BlockChainCommand struct{}

func (a *BlockChainCommand) Text() string        { return "blockchain" }
func (a *BlockChainCommand) Description() string { return "Block related commands" }
func (a *BlockChainCommand) Subcommands() []string {
	return []string{
		"new", "list",
	}
}
func (a *BlockChainCommand) SupportedArguments() []string { return []string{} }
func (a *BlockChainCommand) FollowedBy() []string         { return []string{} }
func (a *BlockChainCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Warn("block need to be followed by commands: new, list.")
}

type BlockChainNewCommand struct{}

func (a *BlockChainNewCommand) Text() string                 { return "new" }
func (a *BlockChainNewCommand) Description() string          { return "Init a new blockchain" }
func (a *BlockChainNewCommand) Subcommands() []string        { return []string{} }
func (a *BlockChainNewCommand) SupportedArguments() []string { return []string{} }
func (a *BlockChainNewCommand) FollowedBy() []string         { return []string{"blockchain"} }
func (a *BlockChainNewCommand) Execute(previousCmds []string, args []IArgument) {
	cAccount := node.GetValueChainNode().Accounts.CurrentAccount
	if cAccount == nil {
		cliLogger.Warn("Need at least an account to run this command")
	} else {
		if node.GetValueChainNode().Chain.GetMyHeight() < 0 {
			if err := block.DeleteBlockchain(); err != nil {
				cliLogger.Error("Failed to delete temp empty blockchain")
			} else {
				block.CreateBlockchain(cAccount.ToWIFCompressed(), func(b *block.Blockchain) {
					cliLogger.Info("Block chain created successfully.")
				})
			}
		} else {
			cliLogger.Warn("Blockchain aleady exists, please remove it first.")
		}
	}
}

type BlockChainListCommand struct{}

func (a *BlockChainListCommand) Text() string                 { return "list" }
func (a *BlockChainListCommand) Description() string          { return "list the content of blockchain" }
func (a *BlockChainListCommand) Subcommands() []string        { return []string{} }
func (a *BlockChainListCommand) SupportedArguments() []string { return []string{} }
func (a *BlockChainListCommand) FollowedBy() []string         { return []string{"blockchain"} }
func (a *BlockChainListCommand) Execute(previousCmds []string, args []IArgument) {
	defer func() {
		if err := recover(); err != nil {
			cliLogger.Error("Got error when printing the chain")
		}
	}()
	block.PrintChain()
}

var __cmd_inst_blockchain = &BlockChainCommand{}
var __cmd_inst_blockchain_new = &BlockChainNewCommand{}
var __cmd_inst_blockchain_list = &BlockChainListCommand{}
