package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"net"
	"sync"
	"time"
)

func main() {
	addr := "bc1qmzf99vrwrpvzpuuc6k5nx40tr39j95vfqsu004"
	v, err := AddressToElectrumScriptHash(addr, &chaincfg.MainNetParams)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

	client := NewElectrumClient("localhost:50001")

	if err := client.Connect(); err != nil {
		panic(err)
	}

	// подписка на новые блоки
	_ = client.SubscribeHeaders()

	// подписка на адрес
	scripthash := addr
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

type ElectrumClient struct {
	addr    string
	conn    net.Conn
	mu      sync.Mutex
	nextID  int
	pending map[int]chan json.RawMessage
	subs    map[string]struct{}
}

func NewElectrumClient(addr string) *ElectrumClient {
	return &ElectrumClient{
		addr:    addr,
		pending: make(map[int]chan json.RawMessage),
		subs:    make(map[string]struct{}),
	}
}

func (c *ElectrumClient) Connect() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.conn = conn
	go c.readLoop()

	// handshake
	_, err = c.Call("server.version", []any{"go-client", "1.4"})
	return err
}

func (c *ElectrumClient) readLoop() {
	reader := bufio.NewReader(c.conn)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("Disconnected:", err)
			go c.reconnect()
			return
		}

		var msg map[string]json.RawMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		// response
		if idRaw, ok := msg["id"]; ok {
			var id int
			_ = json.Unmarshal(idRaw, &id)

			c.mu.Lock()
			ch := c.pending[id]
			delete(c.pending, id)
			c.mu.Unlock()

			if ch != nil {
				ch <- msg["result"]
			}
			continue
		}

		// notification
		if methodRaw, ok := msg["method"]; ok {
			var method string
			_ = json.Unmarshal(methodRaw, &method)

			fmt.Println("Notification:", method, string(msg["params"]))
		}
	}
}

func (c *ElectrumClient) Call(method string, params []any) (json.RawMessage, error) {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan json.RawMessage, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	req := map[string]any{
		"id":     id,
		"method": method,
		"params": params,
	}

	data, _ := json.Marshal(req)
	data = append(data, '\n')

	if _, err := c.conn.Write(data); err != nil {
		return nil, err
	}

	select {
	case res := <-ch:
		return res, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout")
	}
}

func (c *ElectrumClient) SubscribeScriptHash(hash string) error {
	c.subs[hash] = struct{}{}
	_, err := c.Call("blockchain.scripthash.subscribe", []any{hash})
	return err
}

func (c *ElectrumClient) SubscribeHeaders() error {
	_, err := c.Call("blockchain.headers.subscribe", nil)
	return err
}

func (c *ElectrumClient) reconnect() {
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("Reconnecting...")
		if err := c.Connect(); err == nil {
			fmt.Println("Reconnected")
			c.resubscribeAll()
			return
		}
	}
}

func (c *ElectrumClient) resubscribeAll() {
	for h := range c.subs {
		_ = c.SubscribeScriptHash(h)
	}
}
