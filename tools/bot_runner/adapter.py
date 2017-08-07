import json
import os
import socket
import subprocess
from multiprocessing import Process, Queue

import utils
from server_status import read_status

DEFAULT_SERVER = 'punter.inf.ed.ac.uk'


class OfflineAdapter:
    def __init__(self, server, port, exe, log=None, header=None, realtime=False):
        self.server = server
        self.port = port
        self.exe = exe
        self.buffer_size = 1024
        self.punter_id = None

        self._socket = None
        self.buffer = ''

        self._realtime_queue = None

        self.log_file = None
        if log is not None:
            filename = log + '.2.txt'
            os.makedirs(os.path.dirname(filename), exist_ok=True)
            self.log_file = open(filename, 'w')

            status = {}
            if self.server == DEFAULT_SERVER:
                try:
                    status = read_status()[self.port]
                except:
                    pass  # Noop

            metadata = {
                'metadata': 0,
                'server': self.server,
                'port': self.port,
                'status': status,
            }
            if header:
                metadata['extra'] = header

            self.log_file.write(json.dumps(metadata) + '\n')

        if realtime:
            self._start_realtime()

    def _start_realtime(self):
        from realtime_viewer import realtime_viewer

        self._realtime_queue = Queue()
        self.rt_process = Process(target=realtime_viewer, args=(self._realtime_queue,))
        self.rt_process.start()

    def connect(self):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect((self.server, self.port))

    def disconnect(self):
        self._socket.close()
        if self._realtime_queue:
            self._realtime_queue.put('stop')
            self.rt_process.join()

    def send(self, msg):
        print(">>  {}".format(json.dumps(msg)))
        if self.log_file:
            self.log_file.write(">> {}\n".format(json.dumps(msg)))
        msg = format_as_message(msg)
        self._socket.send(msg)

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
        if self._realtime_queue:
            self._realtime_queue.put(json.dumps(msg))
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
            bot.proc.kill()
            self.disconnect()
            if self.log_file:
                self.log_file.close()


class Bot:
    def __init__(self, exe):
        self.proc = subprocess.Popen(exe.split(' '), stdout=subprocess.PIPE, stdin=subprocess.PIPE)

    def write(self, msg):
        self.proc.stdin.write(format_as_message(msg))
        self.proc.stdin.flush()

    def read(self):
        prefix = ''
        while True:
            c = self.proc.stdout.read(1).decode()
            if c == ':':
                break
            prefix += c

        msg_size = int(prefix)
        msg_txt = self.proc.stdout.read(msg_size).decode()
        msg = json.loads(msg_txt)

        return msg


def format_as_message(msg_dict):
    serialized_msg = json.dumps(msg_dict)
    return "{}:{}".format(len(serialized_msg), serialized_msg).encode()


def get_dict_from_message(msg):
    buffer_size_txt = msg.split(':', 1)[0]
    msg_size = int(buffer_size_txt)
    min_buffer_size = len(buffer_size_txt) + 1 + msg_size

    msg_txt = msg[:min_buffer_size]
    return json.loads(msg_txt.split(':', 1)[1])


def ranked(scores):
    return sorted(scores, key=lambda x: x['score'], reverse=True)


def get_metrics(filename):
    with open(filename) as f:
        metadata = None
        if '2.txt' in filename:
            metadata = json.loads(f.readline())
        name = parse(f.readline())['me']
        f.readline()
        punter_id = parse(f.readline())['punter']

        # Skip to the end of the file
        line = None
        for line in f:
            pass

        if not line:
            return 0, 0, name, metadata

        score = parse(line)
        if 'stop' not in score:
            return 0, 0, name, metadata

        scores = score['stop']['scores']
        rank = get_rank(punter_id, scores)
        return rank, len(scores), name, scores, metadata


def get_rank(punter_id, scores):
    for rank, score in enumerate(utils.ranked(scores)):
        if punter_id == score['punter']:
            return rank + 1
    return None


def parse(line):
    return json.loads(line[2:])
