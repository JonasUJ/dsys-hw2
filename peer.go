// A peer represents a connection with another process.
// IMPORTANT: One instance can have multiple peers - this is not a server/client setup, instead
// every process has a server that any other process can connect to.

package main

import (
	"fmt"
	"log"
	"main/tcp"
	"math/rand"
	"sync"
	"time"
)

type State uint8

const (
	New State = iota // Start at New instead of Closed just to be able to differentiate them
	Listen
	SynSent
	SynRecv
	Established
	CloseWait
	LastAck
	TimeWait
	FinWait1
	FinWait2
	Closing
	Closed
)

type MessageType uint8

const (
	Send MessageType = iota
	Close
)

type Stream interface {
	Send(*tcp.Packet) error
	Recv() (*tcp.Packet, error)
}

type Message struct {
	msgType MessageType
	data    string
}

type Retransmit struct {
	when   int64
	packet *tcp.Packet
}

type Peer struct {
	stream  Stream
	state   State
	seq     uint32
	ack     uint32
	name    string
	server  *Server
	mu      sync.Mutex // protects unacked
	unacked map[uint32]Retransmit
	packets chan *tcp.Packet
	msgs    chan Message
}

func NewPeer(stream Stream, state State, server *Server) *Peer {
	return &Peer{
		stream:  stream,
		state:   state,
		name:    "<unknown>",
		server:  server,
		unacked: make(map[uint32]Retransmit),
		packets: make(chan *tcp.Packet),
		msgs:    make(chan Message),
	}
}

func (peer *Peer) String() string {
	return fmt.Sprintf("name:%s state:%d seq:%d ack:%d unacked:%d",
		peer.name,
		peer.state,
		peer.seq,
		peer.ack,
		len(peer.unacked))
}

func (peer *Peer) LowestUnacked() uint32 {
	// Choose high initial min (unlikely that a packet will have this number)
	var min uint32 = 1 << 31
	for k := range peer.unacked {
		if k < min {
			min = k
		}
	}

	return min
}

func (peer *Peer) RetransmitLoop(chExit chan struct{}) {
	for {
		time.Sleep(time.Second)

		// Exit loop when something on channel
		select {
		case <-chExit:
			return
		default:
		}

		if len(peer.unacked) == 0 {
			continue
		}

		// Retransmit the packet if enough time has passed
		retransmit := peer.unacked[peer.LowestUnacked()]
		now := time.Now().UnixMilli()
		if retransmit.when < now {
			log.Printf("retransmission available for packet %+v to %s", retransmit.packet, peer.name)
			peer.Send(retransmit.packet)
		}
	}
}

func (peer *Peer) Send(packet *tcp.Packet) {
	go func() {
		// Keep track of when packet is acked
		if packet.Flag != tcp.Flag_ACK {
			expected := packet.Seq + PacketLen(packet)

			log.Printf("expecting ack %d from %s", expected, peer.name)

			peer.mu.Lock()
			peer.unacked[expected] = Retransmit{
				time.Now().UnixMilli() + time.Second.Milliseconds()*2,
				packet,
			}
			peer.mu.Unlock()

			if expected != peer.LowestUnacked() {
				log.Printf("there exists unacked packets to %s, withholding %+v for now", peer.name, packet)
				return
			}
		}

		// Remember how much we've acked
		if peer.ack < packet.Ack &&
			(packet.Flag == tcp.Flag_ACK ||
				packet.Flag == tcp.Flag_SYNACK) {
			peer.ack = packet.Ack
		}

		// Chance of packet loss
		if rand.Intn(4) != 0 {
			log.Printf("[SEND] %s -> %s (%+v)\n", *name, peer.name, packet)

			// Pretend delay on the wire
			// This also simulates reordering if we send multiple packets quickly
			time.Sleep(time.Second * time.Duration(rand.Int31n(3)))

			peer.stream.Send(packet)
		} else {
			log.Printf("[SEND (lost)] %s -> %s (%+v)\n", *name, peer.name, packet)
		}
	}()
}

func (peer *Peer) Recv(chExit, chExited chan struct{}) {
	for {
		packet, err := peer.stream.Recv()
		if err != nil {
			if peer.state != Closed && peer.state != TimeWait {
				log.Printf("Recv() lost connection to %s\n", peer.name)
			}
			chExited <- struct{}{}
			return
		}

		// Exit loop when something on channel
		select {
		case <-chExit:
			return
		default:
		}

		log.Printf("[RECV] %s -> %s (%+v)\n", peer.name, *name, packet)

		peer.mu.Lock()
		if _, ok := peer.unacked[packet.Ack]; ok {
			// Packet was acked, no longer any need to save for retransmission
			log.Printf("%d was acked by %s\n", packet.Ack, peer.name)

			delete(peer.unacked, packet.Ack)
		}
		peer.mu.Unlock()

		peer.packets <- packet
	}
}

func (peer *Peer) Run() {
	chRecvExit := make(chan struct{}, 1)
	chRecvExited := make(chan struct{}, 1)
	chRetransmitExit := make(chan struct{}, 1)
	go peer.Recv(chRecvExit, chRecvExited)
	go peer.RetransmitLoop(chRetransmitExit)

	for {
		// Encodes the TCP State Machine
		switch peer.state {

		// We want to handshake with someone, send them a syn and go to SynSent
		case New:
			// Send SYN packet. We also send our name for UX reasons
			peer.seq = 1000 // 1000 is random
			peer.Send(&tcp.Packet{
				Flag: tcp.Flag_SYN,
				Seq:  peer.seq,
				Data: *name,
			})
			peer.state = SynSent

		// We're waiting for someone to send a syn, and then go to SynRecv
		case Listen:
			packet := <-peer.packets

			// Ignore other packets
			if packet.Flag == tcp.Flag_SYN {
				peer.name = packet.Data
				peer.seq = 2000 // 2000 is random

				peer.Send(&tcp.Packet{
					Flag: tcp.Flag_SYNACK,
					Seq:  peer.seq,
					Ack:  packet.Seq + PacketLen(packet),
					Data: *name,
				})

				peer.state = SynRecv
			}

		// We've sent a syn, listen for a synack and go to Established
		case SynSent:
			packet := <-peer.packets

			if packet.Flag == tcp.Flag_SYNACK {
				peer.name = packet.Data
				peer.seq += PacketLen(packet)

				peer.Send(&tcp.Packet{
					Flag: tcp.Flag_ACK,
					Seq:  peer.seq,
					Ack:  packet.Seq + PacketLen(packet),
				})

				log.Printf("successfully handshaked with %s\n", peer.name)
				fmt.Printf("connected to %s\n", peer.name)
				peer.server.AddPeer(peer)
				peer.state = Established
			}

		// We received a syn, ack it and go to Established
		case SynRecv:
			packet := <-peer.packets

			if packet.Flag == tcp.Flag_ACK {
				log.Printf("successfully handshaked with %s\n", peer.name)
				fmt.Printf("connected to %s\n", peer.name)
				peer.server.AddPeer(peer)
				peer.state = Established
				peer.seq += PacketLen(packet)
			}

		// Connection establish, listen for packets and user input
		case Established:
			// select on three channels. We want to handle packets and user input, and not block either
			select {
			case packet := <-peer.packets:
				// Check if to make sure we haven't already received and acked this. We don't want
				// to Printf more than once. If we have, then our ack must have been lost.
				if peer.ack <= packet.Seq {
					if packet.Flag == tcp.Flag_NONE {
						fmt.Printf("%s> %s\n", peer.name, packet.Data)
					} else if packet.Flag == tcp.Flag_FIN {
						fmt.Printf("%s is closing the connection\n", peer.name)
						peer.state = CloseWait
					}
				} else {
					log.Printf("got packet that has already been acked (%d > %d)", peer.ack, packet.Seq)
				}

				// Ack the packet
				if packet.Flag != tcp.Flag_ACK {
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_ACK,
						Seq:  peer.seq,
						Ack:  packet.Seq + PacketLen(packet),
					})
				}
			case msg := <-peer.msgs:
				switch msg.msgType {
				case Send:
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_NONE,
						Seq:  peer.seq,
						Data: msg.data,
					})
					peer.seq += Len(msg.data)
				case Close:
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_FIN,
						Seq:  peer.seq,
					})
					peer.seq += 1
					peer.state = FinWait1
				}
			case <-chRecvExited: // If grpc dies
				peer.state = Closed
			}

		case CloseWait:
			time.Sleep(time.Second)
			peer.Send(&tcp.Packet{
				Flag: tcp.Flag_FIN,
				Seq:  peer.seq,
			})
			peer.seq += 1
			peer.state = LastAck

		case LastAck:
			ctx, cancel := Timeout()

			select {
			case <-ctx.Done():
				log.Printf("timed out in LAST_ACK, closing connection to %s\n", peer.name)
				peer.state = Closed

			case packet := <-peer.packets:
				if packet.Flag == tcp.Flag_ACK {
					peer.state = Closed
				} else {
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_ACK,
						Seq:  peer.seq,
						Ack:  packet.Seq + PacketLen(packet),
					})
				}
			}

			cancel()

		case FinWait1:
			ctx, cancel := Timeout()

			select {
			case <-ctx.Done():
				log.Printf("timed out in FIN_WAIT_1, closing connection to %s\n", peer.name)
				peer.state = Closed

			case packet := <-peer.packets:
				// State machine has two paths because the other side might try to FIN at the same
				// time as us or their ack might have been lost.
				if packet.Flag == tcp.Flag_ACK {
					peer.state = FinWait2
				} else if packet.Flag == tcp.Flag_FIN {
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_ACK,
						Seq:  peer.seq,
					})
					peer.seq += 1
					peer.state = Closing
				}
			}

			cancel()

		case FinWait2:
			ctx, cancel := Timeout()

			select {
			case <-ctx.Done():
				log.Printf("timed out in FIN_WAIT_2, closing connection to %s\n", peer.name)
				peer.state = Closed

			case packet := <-peer.packets:
				if packet.Flag == tcp.Flag_FIN {
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_ACK,
						Seq:  peer.seq,
					})
					peer.seq += 1
					peer.state = TimeWait
				}
			}

			cancel()

		case Closing:
			ctx, cancel := Timeout()

			select {
			case <-ctx.Done():
				log.Printf("timed out in CLOSING, closing connection to %s\n", peer.name)
				peer.state = Closed

			case packet := <-peer.packets:
				if packet.Flag == tcp.Flag_ACK {
					peer.state = TimeWait
				}
			}

			cancel()

		case TimeWait:
			fmt.Printf("waiting for timeout to give %s a chance to finish closing connection\n", peer.name)
			ctx, cancel := Timeout()

			select {
			case <-ctx.Done():
				peer.state = Closed

			case packet := <-peer.packets:
				// We're done, but the other side might never have seen our final ack! If they send
				// us something, just ack it. I don't think this is part of tcp, but it makes sense.
				if packet.Flag != tcp.Flag_ACK {
					peer.Send(&tcp.Packet{
						Flag: tcp.Flag_ACK,
						Seq:  peer.seq,
						Ack:  packet.Seq + PacketLen(packet),
					})
				}
			}

			cancel()

		case Closed:
			chRecvExit <- struct{}{}
			chRetransmitExit <- struct{}{}
			fmt.Printf("connection to %s closed\n", peer.name)
			peer.server.RemovePeer(peer)
			return
		}
	}
}
