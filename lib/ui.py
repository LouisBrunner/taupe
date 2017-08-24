import curses
import time

from threading import Lock

from .network_listener import NetworkListener

class UI(NetworkListener):
    ERROR_TIMEOUT = 5

    def __init__(self, network):
        self.network = network

        self.address = None

        self.line = 0
        self.lines = []

        self.loading = False

        self.error_enabled = False
        self.error_message = ''
        self.error_time = 0

        self.pending_lock = Lock()
        self.pending = None

        self.network.register(self)

    def start(self, address):
        self.address = address
        self._refresh()
        curses.wrapper(self._loop)

    def _loop(self, stdscr):
        curses.curs_set(0)
        curses.use_default_colors()

        try:
            stdscr.timeout(1000)
            self._render(stdscr)

            while self._handle_io(stdscr):
                now = time.clock_gettime(time.CLOCK_MONOTONIC)
                if self.error_enabled and (now - self.error_time > self.ERROR_TIMEOUT): self.error_enabled = False

                self._handle_pending()

                self._render(stdscr)

        except KeyboardInterrupt:
            pass

    def _render(self, stdscr):
        height, width = stdscr.getmaxyx()

        stdscr.clear()

        header = 'Taupe: {}'.format(self.address)
        stdscr.addstr(0, 0, header.ljust(width), curses.A_REVERSE)

        for i, line in enumerate(self.lines):
            stdscr.addstr(i + 1, 0, line, curses.A_UNDERLINE if i == self.line else curses.A_NORMAL)

        status = ''
        if self.error_enabled:
            status = 'Error: {}'.format(self.error_message)
        elif self.loading:
            status = 'Loading...'
        stdscr.insstr(height - 1, 0, status.ljust(width), curses.A_REVERSE)

    def _handle_io(self, stdscr):
        on = True
        c = stdscr.getch()
        if c == ord('q') or c == ord('Q') or c == 27:
            on = False
        if not self.loading:
            if c == ord('r') or c == ord('R'):
                self._refresh()
            elif c == ord('\n'):
                self._request_line()
            elif c == curses.KEY_UP:
                self.line = max(self.line - 1, 0)
            elif c == curses.KEY_DOWN:
                self.line = min(self.line + 1, len(self.lines) - 1)
        return on

    def _request_line(self):
        self._error('cannot follow a non-link item')
        # TODO: check if link, then follow, else error
        # self.network.request(self.address)

    def _refresh(self):
        self.loading = True
        self.network.request(self.address)

    def on_result(self, result):
        with self.pending_lock:
            self.pending = ('result', result)

    def on_error(self, error):
        with self.pending_lock:
            self.pending = ('error', result)

    def _handle_pending(self):
        with self.pending_lock:
            if self.pending != None:
                self.loading = False
                if self.pending[0] == 'error':
                    self._error(error)
                elif self.pending[1] == 'result':
                    self.loading = False
                    self.address = result.address
                    self.lines = self._parse_lines(result.lines)
                    self.line = 0
                self.pending = None

    def _parse_lines(self, lines):
        # TODO: better parsing
        return lines

    def _error(self, message):
        self.error_enabled = True
        self.error_message = message
        self.error_time = time.clock_gettime(time.CLOCK_MONOTONIC)
