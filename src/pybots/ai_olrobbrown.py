#!/usr/bin/env python -u
from __future__ import print_function
import sys
import json
import random
from base_bot import PyBot

class OlRobBrownBot(PyBot):
    def setup(self, setup):
        p = setup['punter']

        possible_claims = []
        for river in setup['map']['rivers']:
            possible_claims.append((river['source'], river['target']))

        punter_id = setup['punter']

        # Start with first mine
        available_mines = setup['map']['mines']
        start_mine = random.choice(available_mines)
        available_mines.remove(start_mine)
        prev_sites = [start_mine]
        log("starting w/ {}".format(prev_sites))

        return {
            'ready': p,
            'state': {
                'possible_claims': possible_claims,
                'punter_id': punter_id,
                'available_mines': available_mines,
                'prev_sites':prev_sites
            }
        }

    def choose_first_connected(self, claims, prev_sites, last_end):
        for claim in claims:
            if claim[0] == last_end:
                if not claim[1] in prev_sites:
                    return claim[1], claim
            if claim[1] == last_end:
                if not claim[0] in prev_sites:
                    return claim[0], claim
        return None, None

    def gameplay(self, msg):
        if 'stop' in msg:
            return {}

        if 'timeout' in msg:
            return {}

        possible_claims = msg['state']['possible_claims']
        punter_id = msg['state']['punter_id']
        available_mines = msg['state']['available_mines']
        prev_sites = msg['state']['prev_sites']

        for move in msg['move']['moves']:
            if 'pass' in move:
                continue
            move = move['claim']
            possible_claims.remove([move['source'], move['target']])

        last_idx = -1
        next_site, claim = self.choose_first_connected(possible_claims, prev_sites, prev_sites[last_idx])
        while claim is None:
            log("no options moving from {} in moves {}".format(prev_sites[last_idx], possible_claims))
            last_idx -= 1
            if -last_idx > len(prev_sites):
                break
            next_site, claim = self.choose_first_connected(possible_claims, prev_sites, prev_sites[last_idx])

        if claim is None and available_mines:
            next_mine = random.choice(available_mines)
            available_mines.remove(next_mine)
            next_site, claim = self.choose_first_connected(possible_claims, prev_sites, next_mine)

        if claim is None:
            log("outta options, moves left: {}".format(possible_claims))
            claim = random.choice(possible_claims)
            next_site = claim[0]

        prev_sites.append(next_site)

        log("claimed {}".format(claim))
        log("gamin w/ {}".format(prev_sites))
        return {
            "claim": {
                "punter": punter_id,
                'source': claim[0],
                'target': claim[1]
            },
            'state': {
                'possible_claims': possible_claims,
                'punter_id': punter_id,
                'available_mines': available_mines,
                'prev_sites': prev_sites
            }
        }

    def format_send(self, msg):
        serialized_msg = json.dumps(msg)
        return "{}:{}".format(len(serialized_msg), serialized_msg)

    def read_structured(self, buffer):
        while ':' not in buffer:
            buffer += sys.stdin.read(1)

        buffer_size_txt = buffer.split(':', 1)[0]
        msg_size = int(buffer_size_txt)
        min_buffer_size = len(buffer_size_txt) + msg_size + 1

        while len(buffer) < min_buffer_size:
            buffer += sys.stdin.read(1)

        msg_txt = buffer[:min_buffer_size]
        buffer = buffer[min_buffer_size:]
        return json.loads(msg_txt.split(':', 1)[1]), buffer


    def print_err(self, msg):
        sys.stderr.write("{}\n".format(msg))
        sys.stderr.flush()


def log(s):
    print(s, file=sys.stderr)


if __name__ == '__main__':
    bot = OlRobBrownBot('ol\'robbrown')
    bot.run()
