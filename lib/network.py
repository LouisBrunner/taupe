import socket
from urllib.parse import urlparse
from threading import Thread, Event, Condition

class Network:
    def __init__(self):
        self.listeners = []

        self.stop_event = Event()
        self.cancel_event = Event()
        self.request_event = Condition()
        self.pending_request = None

        self.thread = None

    def register(self, listener):
        self.listeners.append(listener)

    def start(self):
        self.thread = Thread(target=self._loop)
        self.thread.start()

    def stop(self):
        self.stop_event.set()
        self.thread.join()

    def request(self, address):
        self.cancel_event.set()
        with self.request_event:
            self.pending_request = address
            self.request_event.notify()

    def _loop(self):
        with self.request_event:
            while not self.stop_event.is_set():
                self.request_event.wait_for(self._has_no_request, 0.5)
                if not self._has_no_request():
                    self.cancel_event.clear()
                    self._handle_request()

    def _has_no_request(self):
        return self.pending_request == None

    def _handle_request(self):
        url = urlparse(address)
        if url.scheme != 'gopher' or url.scheme != '':
            return self._error('invalid scheme {}'.format(url.scheme))

        if self._should_cancel(): return

        self._result(['abc', 'cde', 'fde'])
        self.pending_request = None

    def _should_cancel(self):
        if self.cancel_event.is_set():
            self.pending_request = None
            return True
        return False

    def _result(self, result):
        self.pending_request = None
        for listener in self.listeners:
            listener.on_result(result)

    def _error(self, error):
        self.pending_request = None
        for listener in self.listeners:
            listener.on_error(error)
