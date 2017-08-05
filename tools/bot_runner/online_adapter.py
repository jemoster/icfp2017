import argparse
import socket
import json
from time import sleep
import subprocess


class OfflineAdapter:
    def __init__(self, server, port, exe, log=None):
        self.server = server
        self.port = port
        self.exe = exe
        self.buffer_size = 1024
        self.punter_id = None

        self._socket = None
        self.buffer = ''

        self.log_file = None
        if log:
            self.log_file = open(log, 'w')

    def connect(self):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect((self.server, self.port))

    def disconnect(self):
        self._socket.close()

    def send(self, msg):
        print(">>  {}\n".format(json.dumps(msg)))
        if self.log_file:
            self.log_file.write(">> {}\n".format(json.dumps(msg)))
        msg = format_as_message(msg)
        self._socket.send(msg)
        sleep(1.0)

    def receive(self, blocking=True):
        while ':' not in self.buffer:
            self.buffer += self._socket.recv(self.buffer_size).decode()

        buffer_size_txt = self.buffer.split(':', 1)[0]
        msg_size = int(buffer_size_txt)
        min_buffer_size = len(buffer_size_txt) + 1 + msg_size

        while len(self.buffer) < min_buffer_size:
            self.buffer += self._socket.recv(self.buffer_size).decode()
            if not blocking:
                return

        msg_txt = self.buffer[:min_buffer_size]
        self.buffer = self.buffer[min_buffer_size:]
        msg = json.loads(msg_txt.split(':', 1)[1])
        print("<<  {}".format(json.dumps(msg)))
        if self.log_file:
            self.log_file.write("<< {}\n".format(json.dumps(msg)))
        return msg

    def run(self):
        self.connect()
        try:
            bot = Bot(self.exe)

            # Handshake with server and bot
            handshake = bot.read()
            self.send(handshake)
            handshake = self.receive()
            bot.write(handshake)

            # Get Setup from server
            setup = self.receive()
            self.punter_id = setup['punter']

            # Setup Bot
            bot.write(setup)

            # Get Bot's setup
            bot_setup = bot.read()
            game_state = bot_setup['state']
            bot_setup.pop('state')
            self.send(bot_setup)

            while True:
                play = self.receive()

                bot = Bot(self.exe)
                bot.read()  # Ignore handshake and use the one we got from the server earlier
                bot.write(handshake)

                play['state'] = game_state
                bot.write(play)

                if 'stop' in play:
                    return ranked(play['stop']['scores'])

                if 'timeout' in play:
                    continue

                move = bot.read()
                game_state = move['state']
                move.pop('state')
                self.send(move)

        finally:
            self.disconnect()
            if self.log_file:
                self.log_file.close()


def format_as_message(msg_dict):
    serialized_msg = json.dumps(msg_dict)
    return "{}:{}".format(len(serialized_msg), serialized_msg).encode()


def ranked(scores):
    return sorted(scores, key=lambda x: x['score'], reverse=True)


def get_dict_from_message(msg):
    buffer_size_txt = msg.split(':', 1)[0]
    msg_size = int(buffer_size_txt)
    min_buffer_size = len(buffer_size_txt) + 1 + msg_size

    msg_txt = msg[:min_buffer_size]
    return json.loads(msg_txt.split(':', 1)[1])


class Bot:
    def __init__(self, exe):
        self.proc = subprocess.Popen(exe.split(' '), stdout=subprocess.PIPE, stdin=subprocess.PIPE)
        self.buffer = ''

    def write(self, msg):
        self.proc.stdin.write(format_as_message(msg))
        self.proc.stdin.flush()

    def read(self, blocking=True):
        while ':' not in self.buffer:
            self.buffer += self.proc.stdout.read(1).decode()

        buffer_size_txt = self.buffer.split(':', 1)[0]
        msg_size = int(buffer_size_txt)
        min_buffer_size = len(buffer_size_txt) + 1 + msg_size

        while len(self.buffer) < min_buffer_size:
            self.buffer += self.proc.stdout.read(1).decode()
            if not blocking:
                return

        msg_txt = self.buffer[:min_buffer_size]
        self.buffer = self.buffer[min_buffer_size:]
        msg = json.loads(msg_txt.split(':', 1)[1])

        return msg


def main():
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')

    parser.add_argument('exe', action="store", help='The executable to evaluate')
    parser.add_argument('port', action="store", type=int, help='Port of the competition server '
                        'see http://punter.inf.ed.ac.uk/status.html for details')
    parser.add_argument('--server', action="store", default="punter.inf.ed.ac.uk",
                        help='The game server to connect to defaults to "punter.inf.ed.ac.uk"')
    parser.add_argument('--record', action="store", default=None, help='filename to save playlog to')
    results = parser.parse_args()

    adapter = OfflineAdapter(results.server, results.port, results.exe, results.record)
    scores = adapter.run()

    for player in scores:
        player_name = 'punter:' if player['punter'] != adapter.punter_id else "me:    "
        print('{} {punter}, score: {score}'.format(player_name, **player))

if __name__ == "__main__":
    main()
