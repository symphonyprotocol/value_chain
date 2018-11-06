package node

import (
	"github.com/symphonyprotocol/simple-node/storage"
	"github.com/symphonyprotocol/sutil/elliptic"
	"fmt"
)

type NodeAccounts struct {
	Accounts []*elliptic.PrivateKey
	CurrentAccount	*elliptic.PrivateKey
}

func LoadNodeAccounts() *NodeAccounts {
	res := &NodeAccounts{ }
	accounts := make([]*elliptic.PrivateKey, 0, 0)
	len, err := storage.GetInt("accounts_length")
	if err == nil {
		for i := int64(0); i < len; i++ {
			bytes, _err := storage.GetData(fmt.Sprintf("account_id_%v", i))
			if _err == nil {
				pri, _ := elliptic.PrivKeyFromBytes(elliptic.S256(), bytes)
				accounts = append(accounts, pri)
			}
		}
	}
	res.Accounts = accounts
	return res
}

func (a *NodeAccounts) NewDerivedAccount() {

}

func (a *NodeAccounts) NewSingleAccount() {

}

func (a *NodeAccounts) ImportAccount() {

}

func (a *NodeAccounts) ExportAccount() {

}

func (a *NodeAccounts) Use(pubKey string) {

}
