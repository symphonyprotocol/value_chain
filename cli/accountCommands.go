package cli

import (
	"github.com/symphonyprotocol/simple-node/node"
	"github.com/symphonyprotocol/swa"
	"github.com/symphonyprotocol/scb/block"
)

type AccountCommand struct {}
func (a *AccountCommand) Text() string { return "account" }
func (a *AccountCommand) Description() string { return "Account related commands" }
func (a *AccountCommand) Subcommands() []string {
	return []string{
		"new", "list", "use", "export", "getbalance",
		"_newmnemonic", "_getkey", "_derivekey",
	}
}
func (a *AccountCommand) SupportedArguments() []string { return []string{} }
func (a *AccountCommand) FollowedBy() []string { return []string{} }
func (a *AccountCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Warn("account need to be followed by commands: newmnemonic, getkey or derivekey.")
}

type AccountNewMCommand struct {}
func (a *AccountNewMCommand) Text() string { return "_newmnemonic" }
func (a *AccountNewMCommand) Description() string { return "Create a new mnemonic." }
func (a *AccountNewMCommand) Subcommands() []string { return []string{} }
func (a *AccountNewMCommand) SupportedArguments() []string { return []string{} }
func (a *AccountNewMCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountNewMCommand) Execute(previousCmds []string, args []IArgument) {
	m, err := swa.GenMnemonic()
	if err == nil {
		cliLogger.Debug("Mnemonic created: %v", m)
	} else {
		cliLogger.Error("%v", err)
	}
}

type AccountGetKeyCommand struct {}
func (a *AccountGetKeyCommand) Text() string { return "_getkey" }
func (a *AccountGetKeyCommand) Description() string { return "Create a private key with your mnemonic." }
func (a *AccountGetKeyCommand) Subcommands() []string { return []string{} }
func (a *AccountGetKeyCommand) SupportedArguments() []string { return []string{ "-m", "-p" } }
func (a *AccountGetKeyCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountGetKeyCommand) Execute(previousCmds []string, args []IArgument) {
	if argM, ok := getArgument(args, "-m"); ok {
		pValue := ""
		if argP, ok := getArgument(args, "-p"); ok {
			pValue = argP.GetValue()
		}
		w, err := swa.NewFromMnemonic(argM.GetValue(), pValue)
		if err == nil {
			pubKey, _ := w.GetMasterKey().ECPubKey()
			cliLogger.Trace("We got the wallet: %v", pubKey.ToAddressCompressed())
		} else {
			cliLogger.Error("Error: %v", err)
		}
	}
}

type AccountDeriveCommand struct {}
func (a *AccountDeriveCommand) Text() string { return "_derivekey" }
func (a *AccountDeriveCommand) Description() string { return "Derive a sub key with your mnemonic and path." }
func (a *AccountDeriveCommand) Subcommands() []string { return []string{} }
func (a *AccountDeriveCommand) SupportedArguments() []string { return []string{ "-m", "-pwd", "-path" } }
func (a *AccountDeriveCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountDeriveCommand) Execute(previousCmds []string, args []IArgument) {
	if argM, ok := getArgument(args, "-m"); ok {
		if argPath, ok2 := getArgument(args, "-path"); ok2 {
			pValue := ""
			if argP, ok := getArgument(args, "-pwd"); ok {
				pValue = argP.GetValue()
			}
			w, err := swa.NewFromMnemonic(argM.GetValue(), pValue)
			if err == nil {
				pubKey, _ := w.GetMasterKey().ECPubKey()
				cliLogger.Trace("We got the wallet: %v", pubKey.ToAddressCompressed())
				cliLogger.Trace("Trying to derive key with path: %v", argPath.GetValue())
				_, pub, err2 := w.DeriveKey(argPath.GetValue())
				if err2 == nil {
					cliLogger.Trace("We got the derived address with path: %v", pub.ToAddressCompressed())
				}
			} else {
				cliLogger.Error("Error: %v", err)
			}
		}
	}
}

type AccountNewCommand struct {}
func (a *AccountNewCommand) Text() string { return "new" }
func (a *AccountNewCommand) Description() string { return "create a new account." }
func (a *AccountNewCommand) Subcommands() []string { return []string{} }
func (a *AccountNewCommand) SupportedArguments() []string { return []string{ } }
func (a *AccountNewCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountNewCommand) Execute(previousCmds []string, args []IArgument) {
	sn := node.GetSimpleNode()
	addr := sn.Accounts.NewSingleAccount("")
	cliLogger.Info("New account created: %v", addr)
}

type AccountListCommand struct {}
func (a *AccountListCommand) Text() string { return "list" }
func (a *AccountListCommand) Description() string { return "list all accounts." }
func (a *AccountListCommand) Subcommands() []string { return []string{} }
func (a *AccountListCommand) SupportedArguments() []string { return []string{ } }
func (a *AccountListCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountListCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Info("Found %v account(s):", len(node.GetSimpleNode().Accounts.Accounts))
	for n, account := range node.GetSimpleNode().Accounts.Accounts {
		if node.GetSimpleNode().Accounts.CurrentAccount == account {
			cliLogger.Info("-> %v) %v", n + 1, account.ECPubKey().ToAddressCompressed())
		} else {
			cliLogger.Info("   %v) %v", n + 1, account.ECPubKey().ToAddressCompressed())
		}
	}
}

type AccountUseCommand struct {}
func (a *AccountUseCommand) Text() string { return "use" }
func (a *AccountUseCommand) Description() string { return "use an account as current account for sending transaction or mining." }
func (a *AccountUseCommand) Subcommands() []string { return []string{} }
func (a *AccountUseCommand) SupportedArguments() []string { return []string{ "-addr" } }
func (a *AccountUseCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountUseCommand) Execute(previousCmds []string, args []IArgument) {
	if argAddr, ok := getArgument(args, "-addr"); ok {
		if (node.GetSimpleNode().Accounts.Use(argAddr.GetValue())) {
			cliLogger.Info("current account changed to %v", argAddr.GetValue())
		} else {
			cliLogger.Error("No such account: %v", argAddr.GetValue())
		}
	} else {
		cliLogger.Warn("account address must be provided via -addr")
	}
}

type AccountExportCommand struct {}
func (a *AccountExportCommand) Text() string { return "export" }
func (a *AccountExportCommand) Description() string { return "export an account's private key to wif." }
func (a *AccountExportCommand) Subcommands() []string { return []string{} }
func (a *AccountExportCommand) SupportedArguments() []string { return []string{ "-addr" } }
func (a *AccountExportCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountExportCommand) Execute(previousCmds []string, args []IArgument) {
	if argAddr, ok := getArgument(args, "-addr"); ok {
		wif := node.GetSimpleNode().Accounts.ExportAccount(argAddr.GetValue())
		if (wif != "") {
			cliLogger.Info("Account exported (WIF Compress): %v", wif)
		} else {
			cliLogger.Error("No such account: %v", argAddr.GetValue())
		}
	} else if node.GetSimpleNode().Accounts.CurrentAccount != nil {
		wif := node.GetSimpleNode().Accounts.CurrentAccount.ToWIFCompressed()
		cliLogger.Info("Account exported (WIF Compress): %v", wif)
	} else {
		cliLogger.Warn("Need an account been selected or pass the account via -addr")
	}
}

type AccountGetBalanceCommand struct {}
func (a *AccountGetBalanceCommand) Text() string { return "getbalance" }
func (a *AccountGetBalanceCommand) Description() string { return "get the balance of an account." }
func (a *AccountGetBalanceCommand) Subcommands() []string { return []string{} }
func (a *AccountGetBalanceCommand) SupportedArguments() []string { return []string{ "-addr" } }
func (a *AccountGetBalanceCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountGetBalanceCommand) Execute(previousCmds []string, args []IArgument) {
	if argAddr, ok := getArgument(args, "-addr"); ok {
		addr := argAddr.GetValue()
		res := block.GetBalance(addr, false)
		gasRes := block.GetBalance(addr, true)
		cliLogger.Info("Account %v's balance: %v, gas balance: %v", addr, res, gasRes)
	} else if node.GetSimpleNode().Accounts.CurrentAccount != nil {
		addr := node.GetSimpleNode().Accounts.CurrentAccount.ECPubKey().ToAddressCompressed()
		res := block.GetBalance(addr, false)
		gasRes := block.GetBalance(addr, true)
		cliLogger.Info("Account %v's balance: %v, gas balance: %v", addr, res, gasRes)
	} else {
		cliLogger.Warn("Need an account been selected or pass the account via -addr")
	}
}

type AccountGetKeyArgumentMnemonic struct { *BaseArgument }
func (a *AccountGetKeyArgumentMnemonic) Text() string { return "-m" }
func (a *AccountGetKeyArgumentMnemonic) Description() string { return "The mnemonic you have." }

type AccountGetKeyArgumentPassword struct { *BaseArgument }
func (a *AccountGetKeyArgumentPassword) Text() string { return "-p" }
func (a *AccountGetKeyArgumentPassword) Description() string { return "The password you want." }

type AccountDeriveArgumentPassword struct { *BaseArgument }
func (a *AccountDeriveArgumentPassword) Text() string { return "-pwd" }
func (a *AccountDeriveArgumentPassword) Description() string { return "The password you want." }

type AccountDeriveArgumentPath struct { *BaseArgument }
func (a *AccountDeriveArgumentPath) Text() string { return "-path" }
func (a *AccountDeriveArgumentPath) Description() string { return "The generation path for the derived key, like m/0." }

type AccountUseArgumentPubkey struct { *BaseArgument }
func (a *AccountUseArgumentPubkey) Text() string { return "-addr" }
func (a *AccountUseArgumentPubkey) Description() string { return "The account address." }


var __cmd_inst_account = &AccountCommand{}
var __cmd_inst_account_newm = &AccountNewMCommand{}
var __cmd_inst_account_getkey = &AccountGetKeyCommand{}
var __cmd_inst_account_derive = &AccountDeriveCommand{}

var __cmd_inst_account_new = &AccountNewCommand{}
var __cmd_inst_account_list = &AccountListCommand{}
var __cmd_inst_account_use = &AccountUseCommand{}
var __cmd_inst_account_export = &AccountExportCommand{}
var __cmd_inst_account_getbalance = &AccountGetBalanceCommand{}

var __arg_inst_account_getkey_m = &AccountGetKeyArgumentMnemonic{&BaseArgument{}}
var __arg_inst_account_getkey_p = &AccountGetKeyArgumentPassword{&BaseArgument{}}
var __arg_inst_account_derive_pwd = &AccountDeriveArgumentPassword{&BaseArgument{}}
var __arg_inst_account_derive_path = &AccountDeriveArgumentPath{&BaseArgument{}}
var __arg_inst_account_use_addr = &AccountUseArgumentPubkey{&BaseArgument{}}


