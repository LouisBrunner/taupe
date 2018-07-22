package taupe

type netOp int

const (
	opStop netOp = iota
	opRequest
)

type netCmd struct {
	op      netOp
	address string
	replyTo chan<- *NetworkEvent
}

// Network starts its own thread to handle network requests asynchronously
type Network struct {
	subscribers []chan *NetworkEvent
	events      chan netCmd
}

// NewNetwork builds a valid Network structure with channels, etc
func NewNetwork() *Network {
	return &Network{
		events: make(chan netCmd, 10),
	}
}

// Start spawns the Network internal loop in a thread
func (network *Network) Start() {
	go network.loop()
}

// Stop sends the stop event to the internal network thread
func (network *Network) Stop() {
	network.events <- netCmd{op: opStop}
}

// Request sends a cancel event for the current request and starts a new request to the provided `address`
func (network *Network) Request(address string) <-chan *NetworkEvent {
	replyTo := make(chan *NetworkEvent)
	network.events <- netCmd{op: opRequest, address: address, replyTo: replyTo}
	return replyTo
}

func (network *Network) loop() {
	for {
		select {
		case event := <-network.events:
			switch event.op {
			case opStop:
				break
			case opRequest:
				event.replyTo <- network.doRequest(event.address)
			}
		}
	}
}
