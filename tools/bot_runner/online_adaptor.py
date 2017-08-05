import argparse
import socket
import json
from time import sleep
import subprocess


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
        print(">>  ", msg)
        self._socket.send(msg)
        sleep(1.0)

    def _receive(self):
        while ':' not in self.buffer:
            self.buffer += self._socket.recv(self.buffer_size).decode()

        buffer_size_txt = self.buffer.split(':', 1)[0]
        msg_size = int(buffer_size_txt)
        min_buffer_size = len(buffer_size_txt) + 1 + msg_size

        if len(self.buffer) < min_buffer_size:
            self.buffer += self._socket.recv(self.buffer_size).decode()
            return

        msg_txt = self.buffer[:min_buffer_size]
        self.buffer = self.buffer[min_buffer_size:]
        msg = json.loads(msg_txt.split(':', 1)[1])
        print("<<  ", json.dumps(msg))
        return msg, msg_txt


def test_send(adapter, msg):
    serialized_msg = json.dumps(msg)
    msg = "{}:{}".format(len(serialized_msg), serialized_msg).encode()
    print('>>  ', json.dumps(msg))
    adapter._send(msg)


def format_as_message(msg_dict):
    serialized_msg = json.dumps(msg_dict)
    return "{}:{}".format(len(serialized_msg), serialized_msg).encode()


def ranked(scores):
    """
    [{'score': 100, 'punter': 0}, ...]
    :param scores:
    :return:
    """
    return sorted(scores, key=lambda x: x['score'], reverse=True)

def get_dict_from_message(msg):
    buffer_size_txt = msg.split(':', 1)[0]
    msg_size = int(buffer_size_txt)
    min_buffer_size = len(buffer_size_txt) + 1 + msg_size

    msg_txt = msg[:min_buffer_size]
    return json.loads(msg_txt.split(':', 1)[1])

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
        proc = subprocess.Popen(results.exe.split(' '), stdout=subprocess.PIPE, stdin=subprocess.PIPE)

        # Handshake with server and bot
        handshake = get_dict_from_message(proc.stdout.readline().decode())
        adapter._send(format_as_message(handshake))
        handshake, _ = adapter._receive()
        proc.stdin.write(format_as_message(handshake))
        proc.stdin.flush()

        # Get Setup from server
        setup = None
        while not setup:
            setup = adapter._receive()
        punter_id = setup[0]['punter']

        # Setup Bot
        proc.stdin.write(setup[1].encode())
        proc.stdin.flush()

        # Get Bot's setup
        bot_setup = get_dict_from_message(proc.stdout.readline().decode())
        game_state = bot_setup['state']
        bot_setup.pop('state')
        adapter._send(format_as_message(bot_setup))

        while True:
            play, raw = adapter._receive()

            proc = subprocess.Popen(results.exe.split(' '), stdout=subprocess.PIPE, stdin=subprocess.PIPE)

            # Handshake
            get_dict_from_message(proc.stdout.readline().decode())
            proc.stdin.write(format_as_message(handshake))
            proc.stdin.flush()

            play['state'] = game_state
            proc.stdin.write(format_as_message(play))
            proc.stdin.flush()

            if 'stop' in play:
                for player in ranked(play['stop']['scores']):
                    player_name = 'punter:' if player['punter'] != punter_id else "me:    "
                    print('{} {punter}, score: {score}'.format(player_name, **player))
                break

            if 'timeout' in play:
                continue

            move = get_dict_from_message(proc.stdout.readline().decode())
            game_state = move['state']
            move.pop('state')
            adapter._send(format_as_message(move))

    finally:
        adapter.disconnect()