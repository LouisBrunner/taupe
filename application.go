package taupe

// Application runs the different parts of the program together (UI, Network...)
type Application struct {
	network *Network
	ui      *UI
}

// NewApplication creates an Application with initialized internals
func NewApplication() *Application {
	network := NewNetwork()
	return &Application{
		network: network,
		ui:      NewUI(network),
	}
}

// Run starts the internals and ensure they stop correctly, `address` is the initial Gopher server requested
func (app *Application) Run(address string) {
	app.network.Start()
	defer app.network.Stop()
	app.ui.Run(address)
}
