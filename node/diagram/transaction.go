package diagram

import (
	"github.com/symphonyprotocol/p2p/tcp"
	"github.com/symphonyprotocol/p2p/models"
	"github.com/symphonyprotocol/scb/block"
)

var (
	// msg broadcasted from whom raised the transaction.
	TX_SEND = "/tx/send"

	// respond to /tx/send, when node has no such trans (not in blocks nor in pending transactions), ask for details
	TX_REQ = "/tx/req"

	// send the details to the requester
	TX_REQ_RES = "/tx/req/res"
)

type TransactionSendDiagram struct {
	models.TCPDiagram
	block.Transaction
}

func NewTransactionSendDiagram(ctx *tcp.P2PContext, tx *block.Transaction) *TransactionSendDiagram {
	tDiag := ctx.NewTCPDiagram()
	tDiag.DType = TX_SEND
	return &TransactionSendDiagram{
		TCPDiagram: *tDiag,
		Transaction: *tx,
	}
}
