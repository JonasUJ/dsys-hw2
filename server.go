// A server is just a bit of a wrapper around the grpc server. It holds all the peers and creates
// new ones when we get told to by grpc.

package main

import (
	"main/tcp"
	"sync"
)

type Server struct {
	tcp.UnimplementedTcpServer

	mu    sync.Mutex // protects peers
	peers map[string]*Peer
}

// Called by grpc when someone dials us, and that someone is a new Peer
func (s *Server) Connect(stream tcp.Tcp_ConnectServer) error {
	NewPeer(stream, Listen, s).Run()

	return nil
}

// Wrap writes in mutex
func (s *Server) AddPeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.peers[peer.name] = peer
}

// Wrap writes in mutex
func (s *Server) RemovePeer(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.peers, peer.name)
}
