package cli

import (
	"github.com/symphonyprotocol/value_chain/node/diagram"
	"strconv"
	"github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/value_chain/node"
)


type TransactionCommand struct {}
func (a *TransactionCommand) Text() string { return "transaction" }
func (a *TransactionCommand) Description() string { return "Transaction related commands" }
func (a *TransactionCommand) Subcommands() []string {
	return []string{
		"send", "get", "list-pending",
	}
}

func (a *TransactionCommand) SupportedArguments() []string { return []string{} }
func (a *TransactionCommand) FollowedBy() []string { return []string{} }
func (a *TransactionCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Warn("transaction need to be followed by commands: send, get, list-pending.")
}

type TransactionSendCommand struct {}
func (a *TransactionSendCommand) Text() string { return "send" }
func (a *TransactionSendCommand) Description() string { return "Raise a transaction from current or specified account to someone." }
func (a *TransactionSendCommand) Subcommands() []string { return []string{} }
func (a *TransactionSendCommand) SupportedArguments() []string { return []string{ "-from", "-to", "-amount" } }
func (a *TransactionSendCommand) FollowedBy() []string { return []string{ "transaction" } }
func (a *TransactionSendCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Info("Going to send")
	fromArg, fromArgOK := getArgument(args, "-from")
	toArg, toArgOK := getArgument(args, "-to")
	amountArg, amountArgOK := getArgument(args, "-amount")
	

	if !toArgOK || !amountArgOK {
		cliLogger.Warn("Need at least -to and -amount parameters to transaction/send command")
		return
	}

	var from, to, amount string
	if fromArgOK {
		from = fromArg.GetValue()
	} else {
		from = node.GetValueChainNode().Accounts.CurrentAccount.ECPubKey().ToAddressCompressed()
	}

	to = toArg.GetValue()
	amount = amountArg.GetValue()

	iAmount, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		cliLogger.Warn("-amount need to be a valid number")
		return
	}

	cliLogger.Info("Really going to send %v to %v from %v", iAmount, to, from)
	tx := block.SendTo(from, to, iAmount, node.GetValueChainNode().Accounts.ExportAccount(from), to == from)
	cliLogger.Info("Initialized transaction: %v", tx.IDString())
	tmpCtx := node.GetValueChainNode().P2PServer.GetP2PContext()
	tmpCtx.BroadcastToNearbyNodes(diagram.NewTransactionSendDiagram(tmpCtx, tx), 20, nil)
	cliLogger.Info("Broadcasted", tx.IDString())
}

type TransactionFromArgument struct { *BaseArgument }
func (a *TransactionFromArgument) Text() string { return "-from" }
func (a *TransactionFromArgument) Description() string { return "The account address where the transaction raised by." }

type TransactionToArgument struct { *BaseArgument }
func (a *TransactionToArgument) Text() string { return "-to" }
func (a *TransactionToArgument) Description() string { return "The account address where the token will be sent to." }

type TransactionAmountArgument struct { *BaseArgument }
func (a *TransactionAmountArgument) Text() string { return "-amount" }
func (a *TransactionAmountArgument) Description() string { return "The amount of the transaction." }

var __cmd_inst_tx = &TransactionCommand{}
var __cmd_inst_tx_send = &TransactionSendCommand{}

var __arg_inst_tx_from = &TransactionFromArgument{&BaseArgument{}}
var __arg_inst_tx_to = &TransactionToArgument{&BaseArgument{}}
var __arg_inst_tx_amount = &TransactionAmountArgument{&BaseArgument{}}
