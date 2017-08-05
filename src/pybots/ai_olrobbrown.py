#!/usr/bin/env python3 -u
import random
from base_bot import PyBot
from astar import graph_neighbors
from base_bot import PyBot, log


class OlRobBrownBot(PyBot):
    def setup(self, setup):
        p = setup['punter']

        possible_claims = []
        rivers = setup['map']['rivers']
        log(setup['map']['rivers'])
        for river in setup['map']['rivers']:
            possible_claims.append((river['source'], river['target']))

        punter_id = setup['punter']

        # Start with first mine
        available_mines = setup['map']['mines']
        start_mine = random.choice(available_mines)
        target_mine = None

        if len(available_mines) > 1:
            target_mine = random.choice(available_mines)
            while target_mine == start_mine:
                target_mine = random.choice(available_mines)

        available_mines.remove(start_mine)
        prev_sites = [start_mine]
        log("starting w/ {}".format(prev_sites))

        return {
            'ready': p,
            'state': {
                'possible_claims': possible_claims,
                'punter_id': punter_id,
                'available_mines': available_mines,
                'target_mine': target_mine,
                'rivers': rivers,
                'prev_sites':prev_sites
            }
        }

    def get_neighbors(self, claims, prev_sites):

        def func(currentNode):
            def get_possible_neighbors(n):

                for claim in claims:
                    if claim[0] == n:
                        if not claim[1] in prev_sites:
                            return claim[1], claim
                    if claim[1] == currentNode:
                        if not claim[0] in prev_sites:
                            return claim[0], claim
                return None, None

            possible_paths = list(filter(get_possible_neighbors, claims))
            return possible_paths
        return func

    def determine_best_move(self, state, last_end):
        target_mine = state['target_mine']
        possible_claims = state['possible_claims']
        prev_sites = state['prev_sites']

        if target_mine is not None:
            pass

        choices = graph_neighbors(target_mine, possible_claims, last_end)

        for claim in possible_claims:
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
        target_mine = msg['state']['target_mine']
        state = msg['state']

        for move in msg['move']['moves']:
            if 'pass' in move:
                continue
            move = move['claim']
            possible_claims.remove([move['source'], move['target']])

        last_idx = -1
        next_site, claim = self.determine_best_move(state, prev_sites[last_idx])
        while claim is None:
            log("no options moving from {} in moves {}".format(prev_sites[last_idx], possible_claims))
            last_idx -= 1
            if -last_idx > len(prev_sites):
                break
            next_site, claim = self.determine_best_move(state, prev_sites[last_idx])

        if claim is None and available_mines:
            next_mine = random.choice(available_mines)
            available_mines.remove(next_mine)
            next_site, claim = self.determine_best_move(state, next_mine)

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
                'prev_sites': prev_sites,
                'target_mine': target_mine
            }
        }


if __name__ == '__main__':
    bot = OlRobBrownBot('ol\'robbrown')
    bot.run()
