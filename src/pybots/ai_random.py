import random
from base_bot import PyBot


class RandoBot(PyBot):
    def setup(self, setup):
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

    def gameplay(self, msg):
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


if __name__ == '__main__':
    bot = RandoBot('Rando-EAGLESSSSS!')
    bot.run()
