#!/usr/bin/env python3
from __future__ import print_function
import argparse
import random
from base_bot import PyBot, log

class trollbot(PyBot):
    def setup(self, setup):
        self.punter = setup['punter']

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
        return ({'ready': self.punter,
            'state': {
                'possible_claims': possible_claims,
                'punter_id': punter_id,
                 'available_mines': available_mines,
                'prev_sites':prev_sites}}
        )

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
        self.punter = msg["state"]["punter_id"]

        possible_claims = msg['state']['possible_claims']
        punter_id = msg['state']['punter_id']
        available_mines = msg['state']['available_mines']
        prev_sites = msg['state']['prev_sites']


        moves = msg['move']['moves']
        for move in moves:
            if 'pass' in move:
                continue
            move = move['claim']
            possible_claims.remove([move['source'], move['target']])

        nextclaim=[-1, -1]

        claim = None

        for move in moves:
            if 'pass' in move:
                continue
            cl = move['claim']
            if cl["punter"] == self.punter:
                continue

            for pc in possible_claims:
                if pc[0] == cl["target"]:
                    claim = pc
                    break


        next_site = None
        last_idx = -1
        if claim is None:
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

        if next_site is not None:
            prev_sites.append(next_site)

        log("claimed {}".format(claim))
        log("gamin w/ {}".format(prev_sites))
        return (
            {"claim": {"punter": punter_id,
             'source': claim[0],
             'target': claim[1]},
             'state': {
                 'possible_claims': possible_claims,
                 'punter_id': punter_id,
                 'available_mines': available_mines,
                 'prev_sites': prev_sites}}
        )


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='ICFP 2017 Online Adapter')
    parser.add_argument('-o', dest='online', action="store_true", help='')
    results = parser.parse_args()

    buffer = ''
    name = random.randint(0, 999999)
    bot = trollbot(str(random.randint(0, 999999)))

    if results.online:
        bot.run_online()
    else:
        bot.run()
