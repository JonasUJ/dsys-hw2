package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"main/tcp"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
)

var (
	name    = flag.String("name", "<no name>", "Name of this instance")
	port    = flag.Int("port", 50050, "Port to listen on")
	logfile = flag.String("logfile", "stdout", "Log file")
)

func main() {
	flag.Parse()
	log.SetFlags(log.Ltime)

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
				if peer.state == Established {
					peer.msgs <- Message{Send, strings.Join(tokens[2:], " ")}
				}
			} else {
				fmt.Printf("no connection exist to %s\n", tokens[1])
			}
		case "close":
			if peer, ok := server.peers[tokens[1]]; ok {
				peer.msgs <- Message{Close, tokens[1]}
			} else {
				fmt.Printf("no connection exist to %s\n", tokens[1])
			}
		case "peers":
			for name, peer := range server.peers {
				fmt.Printf("%s %s\n", name, peer)
			}
		case "exit":
			fmt.Println("bye")
			os.Exit(0)
		case "help":
			fmt.Println(`available commands:
connect <port> -> establish connection to another instance running on <port>
send <peer> <msg> -> send <msg> to connection with name <peer>
close <peer> -> close the connection to a peer
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
