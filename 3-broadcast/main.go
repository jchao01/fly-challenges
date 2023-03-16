package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type server struct {
	n *maelstrom.Node

	idsMu sync.RWMutex
	ids   []int

	topologyMu      sync.RWMutex
	currentTopology map[string][]string
}

func main() {
	n := maelstrom.NewNode()
	s := &server{n: n}

	n.Handle("broadcast", s.handleBroadcast)
	n.Handle("read", s.handleRead)
	n.Handle("topology", s.handleTopology)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var body struct {
		Value int `json:"message"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.idsMu.Lock()
	s.ids = append(s.ids, body.Value)
	s.idsMu.Unlock()

	return s.n.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
}

func (s *server) handleRead(msg maelstrom.Message) error {
	s.idsMu.RLock()
	// copy s.ids to ids
	ids := make([]int, len(s.ids))
	copy(ids, s.ids)
	s.idsMu.RUnlock()

	return s.n.Reply(msg, map[string]any{
		"type":     "read_ok",
		"messages": ids,
	})
}

func (s *server) handleTopology(msg maelstrom.Message) error {
	// parse the topology
	var body struct {
		Topology map[string][]string `json:"topology"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.topologyMu.Lock()
	s.currentTopology = body.Topology
	s.topologyMu.Unlock()

	return s.n.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}
