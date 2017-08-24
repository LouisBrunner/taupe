from .network import Network
from .ui import UI

class Application:
    def __init__(self):
        self.network = Network()
        self.ui = UI(self.network)

    def run(self, args):
        try:
            self.network.start()
            self.ui.start(args.address)
        finally:
            self.network.stop()
