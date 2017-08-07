#!/usr/bin/env python3
from __future__ import print_function
import sys
import json


def log(s):
    print(s, file=sys.stderr)


class PyBot:
    def __init__(self, name):
        self.name = name
        self.buffer = ''

    def run(self):
        # Handshake
        handshake = {'me': self.name}
        self._write(handshake)
        self._read_structured()

        # Execute for state update
        msg_in = self._read_structured()
        if 'punter' in msg_in:
            msg = self.setup(msg_in)
        else:
            msg = self.gameplay(msg_in)

        self._write(msg)

    def run_online(self):
        # Handshake
        handshake = {'me': self.name}
        self._write(handshake)
        hand_in = self._read_structured()

        state = {}
        while True:
            # Execute for state update
            msg_in = self._read_structured()
            msg_in['state'] = state
            if 'punter' in msg_in:
                msg = self.setup(msg_in)
            elif 'stop' in msg_in:
                self.gameplay(msg_in)
                break
            else:
                msg = self.gameplay(msg_in)

            state = msg['state']
            del msg['state']
            self._write(msg)


    @staticmethod
    def _write(msg):
        serialized_msg = json.dumps(msg)
        msg_ser = "{}:{}".format(len(serialized_msg), serialized_msg)
        sys.stdout.write(msg_ser)
        sys.stdout.flush()

    def setup(self, setup):
        raise NotImplementedError

    def gameplay(self, msg):
        raise NotImplementedError

    def _read_structured(self):
        while ':' not in self.buffer:
            self.buffer += sys.stdin.read(1)

        buffer_size_txt = self.buffer.split(':', 1)[0]
        msg_size = int(buffer_size_txt)
        min_buffer_size = len(buffer_size_txt) + msg_size + 1

        while len(self.buffer) < min_buffer_size:
            self.buffer += sys.stdin.read(1)

        msg_txt = self.buffer[:min_buffer_size]
        self.buffer = self.buffer[min_buffer_size:]
        return json.loads(msg_txt.split(':', 1)[1])
