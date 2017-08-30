package lib

// Application runs the different parts of the program together (UI, Network...)
type Application struct {
	Network *Network
	UI      *UI
}

// MakeApplication creates an Application with initialized internals
func MakeApplication() *Application {
	network := MakeNetwork()
	ui := MakeUI(network)
	return &Application{network, ui}
}

// Run starts the internals and ensure they stop correctly, `address` is the initial Gopher server requested
func (app *Application) Run(address string) {
	app.Network.Start()
	defer app.Network.Stop()
	app.UI.Start(address)
}
