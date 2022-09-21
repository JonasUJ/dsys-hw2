# dsys-hw2

Distributed systems homework 2

---

1. What are packages in your implementation? What data structure do you use to transmit data and meta-data?

2. Does your implementation use threads or processes? Why is it not realistic to use threads?

We don't use threads, we use processes. One thing that makes threads unrealistic is that they share memory, while processes do not (unless we do some funky stuff).

3. How do you handle message re-ordering?

4. How do you handle message loss?

5. Why is the 3-way handshake important?

We agree on a sequence number with the other party, which is needed for further communication.
