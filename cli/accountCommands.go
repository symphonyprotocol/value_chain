package cli

import (
	"github.com/symphonyprotocol/swa"
)

type AccountCommand struct {}
func (a *AccountCommand) Text() string { return "account" }
func (a *AccountCommand) Description() string { return "Account related commands" }
func (a *AccountCommand) Subcommands() []string {
	return []string{
		"newmnemonic", "getkey", "derivekey",
	}
}
func (a *AccountCommand) SupportedArguments() []string { return []string{} }
func (a *AccountCommand) FollowedBy() []string { return []string{} }
func (a *AccountCommand) Execute(previousCmds []string, args []IArgument) {
	cliLogger.Warn("account need to be followed by commands: newmnemonic, getkey or derivekey.")
}

type AccountNewCommand struct {}
func (a *AccountNewCommand) Text() string { return "newmnemonic" }
func (a *AccountNewCommand) Description() string { return "Create a new mnemonic." }
func (a *AccountNewCommand) Subcommands() []string { return []string{} }
func (a *AccountNewCommand) SupportedArguments() []string { return []string{} }
func (a *AccountNewCommand) FollowedBy() []string { return []string{ "account" } }
func (a *AccountNewCommand) Execute(previousCmds []string, args []IArgument) {
	m, err := swa.GenMnemonic()
	if err == nil {
		cliLogger.Debug("Mnemonic created: %v", m)
	} else {
		cliLogger.Error("%v", err)
	}
}

type AccountGetKeyCommand struct {}
func (a *AccountGetKeyCommand) Text() string { return "getkey" }
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
			cliLogger.Trace("We got the wallet: %v", pubKey.ToAddress())
		} else {
			cliLogger.Error("Error: %v", err)
		}
	}
}

type AccountDeriveCommand struct {}
func (a *AccountDeriveCommand) Text() string { return "derivekey" }
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
				cliLogger.Trace("We got the wallet: %v", pubKey.ToAddress())
				cliLogger.Trace("Trying to derive key with path: %v", argPath.GetValue())
				_, pub, err2 := w.DeriveKey(argPath.GetValue())
				if err2 == nil {
					cliLogger.Trace("We got the derived address with path: %v", pub.ToAddress())
				}
			} else {
				cliLogger.Error("Error: %v", err)
			}
		}
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



var __cmd_inst_account = &AccountCommand{}
var __cmd_inst_account_new = &AccountNewCommand{}
var __cmd_inst_account_getkey = &AccountGetKeyCommand{}
var __cmd_inst_account_derive = &AccountDeriveCommand{}

var __arg_inst_account_getkey_m = &AccountGetKeyArgumentMnemonic{&BaseArgument{}}
var __arg_inst_account_getkey_p = &AccountGetKeyArgumentPassword{&BaseArgument{}}
var __arg_inst_account_derive_pwd = &AccountDeriveArgumentPassword{&BaseArgument{}}
var __arg_inst_account_derive_path = &AccountDeriveArgumentPath{&BaseArgument{}}


