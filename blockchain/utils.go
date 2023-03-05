package blockchain

import (
	"github.com/btcsuite/btcd/bittrace"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func GetGenesisTX(b *wire.MsgBlock) *wire.MsgTx {
	var genesisTX *wire.MsgTx
	for _, tx := range b.Transactions {
		if IsCoinBaseTx(tx) {
			genesisTX = tx
			break
		}
	}
	return genesisTX
}

func GetMinerAddress(b *wire.MsgBlock, params *chaincfg.Params) string {
	genesisTX := GetGenesisTX(b)
	var minerAddress string
	for _, out := range genesisTX.TxOut {
		// 将第一个 tx out 的地址作为 miner address
		_, minerAddresses, n, err := txscript.ExtractPkScriptAddrs(out.PkScript, params)
		if err == nil && n > 0 && len(minerAddresses) > 0 {
			minerAddress = minerAddresses[0].EncodeAddress()
		} else {
			if err != nil {
				bittrace.Warn("parse miner address get err:%v", err)
			}
		}
	}
	return minerAddress
}
