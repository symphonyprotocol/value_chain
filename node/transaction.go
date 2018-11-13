package node

import (
	"github.com/symphonyprotocol/p2p/tcp"
)

var (
	// msg broadcasted from whom raised the transaction.
	TRANSACTION_SEND = "/trans/send"

	// respond to /trans/send, when node has no such trans (not in blocks nor in pending transactions), ask for details
	TRANSACTION_REQ = "/trans/req"

	// send the details to the requester
	TRANSACTION_REQ_RES = "/trans/req/res"
)

type TransactionMiddleware struct {
	*tcp.BaseMiddleware
}

func (t *TransactionMiddleware) Start(ctx *tcp.P2PContext) {
	t.regHandlers()
}

func (t *TransactionMiddleware) regHandlers() {
	t.HandleRequest(TRANSACTION_SEND, func (ctx *tcp.P2PContext) {
		// 1. check if I already recieved this transaction

		// 2. ask for it if I don't have
	})

	t.HandleRequest(TRANSACTION_REQ, func (ctx *tcp.P2PContext) {
		// 1. check if I have this transaction's content

		// 2. if so, send back the content
	})

	t.HandleRequest(TRANSACTION_REQ_RES, func (ctx *tcp.P2PContext) {
		// recieve and save to pending trans list
	})
}
