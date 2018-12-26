package node

import (
	"github.com/symphonyprotocol/log"
	"github.com/symphonyprotocol/swa"
	"github.com/symphonyprotocol/value_chain/storage"
	"github.com/symphonyprotocol/sutil/elliptic"
	"fmt"
)

var walletLogger = log.GetLogger("wallet")

type NodeAccounts struct {
	Accounts []*elliptic.PrivateKey
	CurrentAccount	*elliptic.PrivateKey
}

func LoadNodeAccounts() *NodeAccounts {
	res := &NodeAccounts{ }
	accounts := make([]*elliptic.PrivateKey, 0, 0)
	length, err := storage.GetInt("accounts_length")
	if err == nil {
		for i := int64(0); i < length; i++ {
			bytes, _err := storage.GetData(fmt.Sprintf("account_id_%v", i))
			if _err == nil {
				pri, _ := elliptic.PrivKeyFromBytes(elliptic.S256(), bytes)
				accounts = append(accounts, pri)
			}
		}
	} else {
		storage.SaveInt("accounts_length", int64(0))
	}
	if len(accounts) > 0 {
		res.CurrentAccount = accounts[0]
	}
	res.Accounts = accounts
	return res
}

func (a *NodeAccounts) NewDerivedAccount() {
}

func (a *NodeAccounts) NewSingleAccount(pwd string) string {
	m, err := swa.GenMnemonic()
	if err == nil {
		w, err := swa.NewFromMnemonic(m, "")
		if err == nil {
			mKey := w.GetMasterKey()
			k, err := mKey.ECPrivKey()
			if err == nil {
				addr := k.ECPubKey().ToAddressCompressed()
				walletLogger.Trace("Successfully created the account: %v", addr)
				a.Accounts = append(a.Accounts, k)
				a.storeAccount(k)
				if len(a.Accounts) == 1 {
					a.CurrentAccount = k
				}
				return addr
			}
		}
	}
	return ""
}

func (a *NodeAccounts) ImportAccount() {

}

func (a *NodeAccounts) ExportAccount(pubKey string) string {
	for _, account := range a.Accounts {
		if account.ECPubKey().ToAddressCompressed() == pubKey {
			return account.ToWIFCompressed()
		}
	}
	return ""
}

func (a *NodeAccounts) Use(pubKey string) (ok bool) {
	if a.CurrentAccount != nil && a.CurrentAccount.ECPubKey().ToAddressCompressed() == pubKey {
		return true
	}

	found := false

	for _, account := range a.Accounts {
		if account.ECPubKey().ToAddressCompressed() == pubKey {
			a.CurrentAccount = account
			found = true
			break
		}
	}
	return found
}

func (a *NodeAccounts) storeAccount(pri *elliptic.PrivateKey) {
	if len, err := storage.GetInt("accounts_length"); err == nil {
		_err := storage.SaveData(fmt.Sprintf("account_id_%v", len), pri.PrivatekeyToBytes())
		if _err == nil {
			walletLogger.Trace("Account stored successfully")
			storage.SaveInt("accounts_length", len + 1)
		}
	}
}
