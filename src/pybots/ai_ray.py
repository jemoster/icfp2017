#!/usr/bin/env python3

from base_bot import PyBot


class RayBot(PyBot):
    def __init__(self, name):
        super().__init__(name)
        self.state = {}

    def setup(self, setup):
        self.state.update(setup)
        return {'ready': self.id, 'state': self.state}

    def gameplay(self, msg):
        self.state.update(msg.get('state', {}))
        move = {'pass': {'punter': self.id}, 'state': self.state}
        return move

    @property
    def id(self):
        return self.state['punter']


if __name__ == '__main__':
    bot = RayBot('ray')
    bot.run()
