package main

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

var (
	State_name = map[State]string{
		New: "New",
		Listen: "Listen",
		SynSent: "SynSent",
		SynRecv: "SynRecv",
		Established: "Established",
		CloseWait: "CloseWait",
		LastAck: "LastAck",
		TimeWait: "TimeWait",
		FinWait1: "FinWait1",
		FinWait2: "FinWait2",
		Closing: "Closing",
		Closed: "Closed",
	}
)

func (state State) String() string {
	return State_name[state]
}
