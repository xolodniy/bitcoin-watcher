package main

import (
	"bitcoin-watcher/api"
	"bitcoin-watcher/electrum_client"
	"bitcoin-watcher/invoice_manager"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func main() {
	addr := "bc1qmzf99vrwrpvzpuuc6k5nx40tr39j95vfqsu004"
	scripthash, err := AddressToElectrumScriptHash(addr, &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}

	client := electrum_client.NewElectrumClient("localhost:50001")

	if err := client.Connect(); err != nil {
		panic(err)
	}
	// подписка на новые блоки
	_ = client.SubscribeHeaders()

	manager := invoice_manager.NewInvoiceManager(client)
	ctrl := api.New(manager)
	_ = ctrl
	manager.HandleScriptHash(scripthash)

	_ = client.SubscribeScriptHash(scripthash)

	select {}
}

func AddressToElectrumScriptHash(addr string, params *chaincfg.Params) (string, error) {
	address, err := btcutil.DecodeAddress(addr, params)
	if err != nil {
		return "", err
	}
	if !address.IsForNet(params) {
		return "", errors.New("address is for wrong network")
	}

	scriptPubKey, err := txscript.PayToAddrScript(address)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(scriptPubKey)

	// electrum требует reverse byte order
	reversed := reverseBytes(hash[:])

	return hex.EncodeToString(reversed), nil
}

func reverseBytes(b []byte) []byte {
	result := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		result[i] = b[len(b)-1-i]
	}
	return result
}
