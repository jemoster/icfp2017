#!/usr/bin/env python3

import argparse
import json
import os
import random

from ai_ray import RayBot

maps = {
    'sample.json': 2,
    'Sierpinski-triangle.json': 3,
    'circle.json': 4,
    'lambda.json': 4,
    'randomMedium.json': 4,
    'randomSparse.json': 4,
    # 'tube.json': 8,
    # 'boston-sparse.json': 8,
    # 'nara-sparse.json': 16,
    # 'oxford.json': 16,
    # 'edinburgh-sparse.json': 16,
    # 'gothenburg-sparse.json': 16,
}


class Game():
    def __init__(self, mapfile):
        if mapfile is None:
            mapfile = random.choice(list(maps.keys()))
        assert mapfile in maps, 'Invalid mapfile: {}'.format(mapfile)
        self.mapfile = mapfile
        self.map = json.load(open(os.path.join('maps', mapfile)))
        self.n_punters = maps[mapfile]
        self.punters = [RayBot(str(i)) for i in range(self.n_punters)]

    def run(self):
        setup = {'punters': self.n_punters, 'map': self.map}
        for i in range(self.n_punters):
            setup['punter'] = i
            self.punters[i].setup(setup)
        print(self.punters)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Play a game')
    parser.add_argument('--mapfile', type=str, help='map file to play')
    args = parser.parse_args()
    Game(args.mapfile).run()
