package main

import (
	"context"
	"fmt"
	"log"
	"main/tcp"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// We add one because we wouldn't send a packet without data. This way all packets has at least 1
// byte of data.
func Len(s string) uint32 {
	return uint32(len(s) + 1)
}

// Some packet types only have len 1, even though we send data in them
func PacketLen(packet *tcp.Packet) uint32 {
	// syn and synack always have exactly 1 byte of data
	if packet.Flag == tcp.Flag_SYN || packet.Flag == tcp.Flag_SYNACK {
		return 1
	} else {
		return Len(packet.Data)
	}
}

// Get a context with a deadline
func Timeout() (context.Context, context.CancelFunc) {
	deadline := time.Now().Add(10 * time.Second)
	return context.WithDeadline(context.Background(), deadline)
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
