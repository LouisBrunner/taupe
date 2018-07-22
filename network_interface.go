package taupe

// NetworkManager is a class that can do Gopher requests
type NetworkManager interface {
	Request(string) <-chan *NetworkEvent
}

// NetworkEventType is a type of event that be returned by the NetworkManager
type NetworkEventType int

// Type of possible network results
const (
	NetworkEventOK NetworkEventType = iota
	NetworkEventHTML
	NetworkEventError
)

// NetworkEvent represents any answer from the Network
type NetworkEvent struct {
	Event       NetworkEventType
	Result      *NetworkResult
	ResultHTML  *NetworkResultHTML
	ResultError error
}

// NetworkResult is a Gopher answer from a request to the NetworkManager class
type NetworkResult struct {
	Address string
	List    []string
}

// NetworkResultHTML is a HTML answer from a request to the NetworkManager class
type NetworkResultHTML struct {
	Address string
	HTML    string
}
