package lib

type Application struct {
  Network *Network
  UI *UI
}

func MakeApplication() *Application {
  network := MakeNetwork()
  ui := MakeUI(network)
  return &Application{network, ui}
}

func (self *Application) Run(address string) {
  self.Network.Start()
  defer self.Network.Stop()
  self.UI.Start(address)
}
