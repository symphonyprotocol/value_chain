package node

import (
	"github.com/symphonyprotocol/simple-node/node/diagram"
	"github.com/symphonyprotocol/p2p/tcp"
)

type TransactionMiddleware struct {
	*tcp.BaseMiddleware
}

func (t *TransactionMiddleware) Start(ctx *tcp.P2PContext) {
	t.regHandlers()
}

func (t *TransactionMiddleware) regHandlers() {
	t.HandleRequest(diagram.TX_SEND, func (ctx *tcp.P2PContext) {
		// 1. check if I already recieved this transaction

		// 2. ask for it if I don't have
	})

	t.HandleRequest(diagram.TX_REQ, func (ctx *tcp.P2PContext) {
		// 1. check if I have this transaction's content

		// 2. if so, send back the content
	})

	t.HandleRequest(diagram.TX_REQ_RES, func (ctx *tcp.P2PContext) {
		// recieve and save to pending trans list
	})
}
