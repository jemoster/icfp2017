import argparse
import socket
import json
from time import sleep
import random


class OfflineAdapter:
    def __init__(self, server, port, exe):
        self.server = server
        self.port = port
        self.buffer_size = 1024

        self._socket = None
        self.buffer = ''

    def connect(self):
        print('Connecting to {}:{}'.format(self.server, self.port))
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect((self.server, self.port))
        print('Connected')

    def disconnect(self):
        print('Disconnecting')
        self._socket.close()

    def _send(self, msg):
        self._socket.send(msg)
        sleep(1.0)

    def _receive(self):
        while ':' not in self.buffer:
            self.buffer += self._socket.recv(self.buffer_size).decode()

        buffer_size_txt = self.buffer.split(':', 1)[0]
        msg_size = int(buffer_size_txt)
        min_buffer_size = len(buffer_size_txt) + 1 + msg_size

        if len(self.buffer) < min_buffer_size:
            return

        msg_txt = self.buffer[:min_buffer_size]
        self.buffer = self.buffer[min_buffer_size:]
        return json.loads(msg_txt.split(':', 1)[1])


def test_send(adapter, msg):
    serialized_msg = json.dumps(msg)
    msg = "{}:{}".format(len(serialized_msg), serialized_msg).encode()
    print('>>  ', msg)
    adapter._send(msg)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')

    parser.add_argument('exe', action="store", help='The executable to evaluate')
    parser.add_argument('port', action="store", type=int, help='Port of the competition server '
                        'see http://punter.inf.ed.ac.uk/status.html for details')
    parser.add_argument('--server', action="store", default="punter.inf.ed.ac.uk",
                        help='The game server to connect to defaults to "punter.inf.ed.ac.uk"')
    results = parser.parse_args()

    adapter = OfflineAdapter(results.server, results.port, results.exe)
    adapter.connect()
    try:
        test_send(adapter, {'me': 'unhinged_muffin'})

        # Wait for server response
        handshake = adapter._receive()
        print("<<  ", handshake)

        setup = adapter._receive()
        print("<<  ", setup)
        possible_claims = []
        for river in setup['map']['rivers']:
            possible_claims.append((river['source'], river['target']))

        punter_id = setup['punter']
        print('punter_id:', punter_id)

        test_send(adapter, {'ready': punter_id})

        while True:
            play = adapter._receive()
            print("<<  ", play)

            if 'stop' in play:
                break

            for move in play['move']['moves']:
                if 'pass' in move:
                    continue
                move = move['claim']
                possible_claims.remove((move['source'], move['target']))

            claim = random.choice(possible_claims)
            test_send(adapter, {"claim": {"punter": punter_id, 'source': claim[0], 'target': claim[1]}})
            # test_send(adapter, {'move': {"pass": {"punter": punter_id}}})

    finally:
        adapter.disconnect()