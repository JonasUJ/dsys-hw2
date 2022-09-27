[Link to the github](https://github.com/JonasUJ/dsys-hw2)

# dsys-hw2

(short for: Distributed systems homework 2)

Apologies to however ends up reviewing this, things got a bit out of hand. We are totally aware that our implementation is excessive, but it was fun to do so anyway :)

## Run the thing
- Start a process: `go run . -name alice -port 50050`
- Then, start another process: `go run . -name bob -port 50051`
- (Optionally, start a third process: `go run . -name charlie -port 50052`)
- (And so on...)

It displays a message "type 'help' for a list of commands" when it is ready.

Every process has its own server that can be connected to, so from any of the processes, type the command `connect <port>` where `<port>` is the port of any other process. This will make the two processes perform a TCP 3-way handshake and display a message once complete. Then, if you started more than two processes, you can repeat this with the port of any of the other processes to connect to them as well (on process can have multiple connections).

Once connected, use the `send <peer> <msg>` command to send a message (packet with non-empty data field) to the specified peer. The `<peer>` parameter is the name of a peer you've connected to - use the `peers` command to list all connections with names. When a process attempts to send a message, there is a chance it will be "lost" (i.e. we discard it instead of sending it), if it is sent however, we apply a random delay so that things might arrive out-of-order.

When you're done and want to terminate the connection you can do one of the following:
- Use the command `exit`. This will kill the process and hence the gRPC stream, which all connected peers will see (not very pretty :c).
- Type `^C` (signal interrupt) and terminate the process like `exit`.
- Use the command `close <peer>`. This will perform a 3/4-way handshake (sending FIN/ACK packets) to terminate the connection with the peer according to spec (very pretty :D).

### Logs
The program accepts a `-logfile` parameter, which is a file where all logs are written. You might want to do this `go run . -name alice -port 50050 -logfile alice.log` and then do `tail -f alice.log` in another terminal. Going a step further would be something like `tail -f alice.log | rg '\[.*?\]|\(.*? -> .*?\)'` if you have `rg`, or just use `grep` if not. This will make the logs slighty more readable:

![log output for one side of 3-way handshake](https://github.com/JonasUJ/dsys-hw2/blob/main/media/log.png?raw=true)

## Q&A

1. What are packages in your implementation? What data structure do you use to transmit data and meta-data?

We assume that by packages the question means packets.
Packets in our implementation is a struct, and is defined as a message in [tcp.proto](https://github.com/JonasUJ/dsys-hw2/blob/main/tcp/tcp.proto#L19-L24). It has the following declaration (which `protoc` transforms into a go struct with the same fields).
```cs
message Packet {
    Flag flag = 1;
    uint32 seq = 2;
    uint32 ack = 3;
    string data = 4;
}
```

2. Does your implementation use threads or processes? Why is it not realistic to use threads?

Our implementation uses threads and processes, our threads, however, are not used for simulating TCP. One thing that makes threads unrealistic is that they share memory, while processes do not (unless we do some funky stuff).

3. How do you handle message re-ordering?

We use sequence numbers to keep track of our packets and compare them to ACKs. If we receive an out-of-order packet we discard it and rely on the other party to retransmit it, upon not receiving an ACK for it within a reasonable timeframe.

4. How do you handle message loss?

We handle message loss by checking every second if there is a packet that we should retransmit, in case we haven't received an ackownledgement for that packet yet.

5. Why is the 3-way handshake important?

We agree on a sequence number with the other party, which is needed for further communication. This number is especially important as is evident from our answers above.
