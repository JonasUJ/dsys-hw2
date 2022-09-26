package main

type TcpClient struct {
	sendChan chan Packet
	recvChan <-chan Packet
}

type Packet struct {
	seq   int32
	ack   int32
	flags byte
	data  []byte
}
