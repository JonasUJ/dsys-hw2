package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"main/tcp"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type State uint8

const (
	New State = iota // Start at New instead of Closed just to be able to differentiate them
	Listen
	SynSent
	SynRecv
	Established
	Closed
	FinWait1
	FinWait2
	Closing
	TimeWait
	CloseWait
	LastAck
)

type MessageType uint8

const (
	Send MessageType = iota
	Close
)

var (
	name    = flag.String("name", "[no name]", "Name of this instance")
	port    = flag.Int("port", 50050, "Port to listen on")
	logfile = flag.String("logfile", "stdout", "Log file")
)

func Len(s string) uint32 {
	return uint32(len(s) + 1)
}

func PacketLen(packet *tcp.Packet) uint32 {
	// syn and synack always have + 1 to ack
	if packet.Flag == tcp.Flag_SYN || packet.Flag == tcp.Flag_SYNACK {
		return 1
	} else {
		return Len(packet.Data)
	}
}

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
		name:    "[unknown]",
		server:  server,
		unacked: make(map[uint32]Retransmit),
		packets: make(chan *tcp.Packet),
		msgs:    make(chan Message),
	}
}

func (peer *Peer) Format() string {
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
		if rand.Intn(3) != 0 {
			log.Printf("sending packet %+v to %s\n", packet, peer.name)

			// Pretend delay on the wire
			// This also simulates reordering if we send multiple packets quickly
			time.Sleep(time.Second * time.Duration(rand.Int31n(3)))

			peer.stream.Send(packet)
		} else {
			log.Printf("simulated packet loss for %+v to %s\n", packet, peer.name)
		}
	}()
}

func (peer *Peer) Recv(chExit chan struct{}) {
	for {
		packet, err := peer.stream.Recv()
		if err != nil {
			log.Printf("Recv lost connection to %s\n", peer.name)
			chExit <- struct{}{}
			return
		}

		log.Printf("received packet %+v from %s\n", packet, peer.name)

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
	chRecv := make(chan struct{})
	chRetransmit := make(chan struct{})
	go peer.Recv(chRecv)
	go peer.RetransmitLoop(chRetransmit)

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
						// TODO
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
				}
			case <-chRecv: // If grpc dies
				peer.state = Closed
			}
		case Closed:
			chRetransmit <- struct{}{}
			fmt.Printf("connection to %s closed\n", peer.name)
			peer.server.RemovePeer(peer)
			return
		default:
			panic("unimplemented state")
		}
	}
}

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

// Dial grpc server listening on the given port and create a new Peer
func Connect(port string, server *Server) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := tcp.NewTcpClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		log.Fatalf("fail to connect: %v", err)
	}

	go NewPeer(stream, New, server).Run()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.Parse()

	if *logfile != "stdout" {
		file, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("could not open file %s: %v", *logfile, err)
		}
		log.SetOutput(file)
	}

	// We need a listener for grpc
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("fail to listen on port %d: %v", *port, err)
	}
	defer listener.Close()

	server := &Server{peers: make(map[string]*Peer)}

	// Serve grpc server in another thread as not to block user input
	go func() {
		grpcServer := grpc.NewServer()
		tcp.RegisterTcpServer(grpcServer, server)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("stopped serving: %v", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("type 'help' for a list of commands")
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")

		// Dispatch commands
		switch tokens[0] {
		case "connect":
			Connect(tokens[1], server)
		case "send":
			if peer, ok := server.peers[tokens[1]]; ok {
				peer.msgs <- Message{Send, strings.Join(tokens[2:], " ")}
			} else {
				fmt.Printf("no connection exist to %s\n", tokens[1])
			}
		case "peers":
			for name, peer := range server.peers {
				fmt.Printf("%s %s\n", name, peer.Format())
			}
		case "exit":
			fmt.Println("bye")
			os.Exit(0)
		case "help":
			fmt.Println(`available commands:
connect <port> -> establish connection to another instance running on <port>
send <peer> <msg> -> send <msg> to connection with name <peer>
peers -> display all connected peers
exit -> terminate the program
help -> display this help message`)
		default:
			fmt.Printf("unknown command '%s', type 'help' for a list of commands\n", tokens[0])
		}
	}

	if scanner.Err() != nil {
		log.Fatalf("fail to read stdin: %v", scanner.Err())
	}
}
