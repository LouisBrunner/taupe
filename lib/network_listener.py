class NetworkListener:
    def on_error(self, error):
        raise NotImplementedError('NetworkListener::on_error not implemented')

    def on_result(self, result):
        raise NotImplementedError('NetworkListener::on_result not implemented')
