import sys
import json
import random


def print_err(msg):
    """ Print a message to stderr to keep stdout clear"""
    sys.stderr.write("{}\n".format(msg))
    sys.stderr.flush()


class PyBot:
    def __init__(self, name):
        self.name = name
        self.buffer = ''

    def run(self):
        # Handshake
        handshake = {'me': self.name}
        self._write(handshake)
        hand_in = self._read_structured()

        # Execute for state update
        msg_in = self._read_structured()
        if 'punter' in msg_in:
            msg = self.setup(msg_in)
        else:
            msg = self.gameplay(msg_in)

        self._write(msg)

    @staticmethod
    def _write(msg):
        serialized_msg = json.dumps(msg)
        msg_ser = "{}:{}".format(len(serialized_msg), serialized_msg)
        sys.stdout.write(msg_ser)
        sys.stdout.flush()

    @staticmethod
    def setup(setup):
        p = setup['punter']

        possible_claims = []
        for river in setup['map']['rivers']:
            possible_claims.append((river['source'], river['target']))

        punter_id = setup['punter']

        return {
            'ready': p,
            'state': {
                'possible_claims': possible_claims,
                'punter_id': punter_id
            }
        }

    @staticmethod
    def gameplay(msg):
        if 'stop' in msg:
            return {}

        if 'timeout' in msg:
            return {}

        possible_claims = msg['state']['possible_claims']
        punter_id = msg['state']['punter_id']

        for move in msg['move']['moves']:
            if 'pass' in move:
                continue
            move = move['claim']
            possible_claims.remove([move['source'], move['target']])

        claim = random.choice(possible_claims)

        return {
            'claim': {
                'punter': punter_id,
                'source': claim[0],
                'target': claim[1]
            },
            'state': {
                'possible_claims': possible_claims,
                'punter_id': punter_id
            }
        }

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


if __name__ == '__main__':
    bot = PyBot('EAGLESSSSS!')
    bot.run()
