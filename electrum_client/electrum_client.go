package electrum_client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

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
	_, err = c.call("server.version", []any{"go-client", "1.4"})
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

			// TODO: pass to invoice manager
			if method == "blockchain.headers.subscribe" {
				var params []map[string]any
				json.Unmarshal(msg["params"], &params)
				// height := int(params[0]["height"].(float64))
				// manager.UpdateHeight(height)
			}

			if method == "blockchain.scripthash.subscribe" {
				var params []any
				json.Unmarshal(msg["params"], &params)
				// hash := params[0].(string)
				// manager.HandleScriptHash(hash)
			}
		}
	}
}
func (c *ElectrumClient) CallHashHistory(hash string) (json.RawMessage, error) {
	return c.call("blockchain.scripthash.get_history", []any{hash})
}

func (c *ElectrumClient) call(method string, params []any) (json.RawMessage, error) {
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
	_, err := c.call("blockchain.scripthash.subscribe", []any{hash})
	return err
}

func (c *ElectrumClient) SubscribeHeaders() error {
	_, err := c.call("blockchain.headers.subscribe", nil)
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
