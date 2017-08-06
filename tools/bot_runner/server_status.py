#!/usr/bin/env python3
import urllib.request
import json


def waiting_for(game):
    return game['total_punters'] - len(game['punters'])


def read_status(retries=10):
    PUNTER_STATUS = 'http://punter.inf.ed.ac.uk/status.json'

    for retry in range(retries):
        try:
            with urllib.request.urlopen(PUNTER_STATUS) as response:
               html = response.read()
            game_list = json.loads(html.decode())
        except Exception as e:
            continue
        break

    games = {}
    for game in game_list['games']:
        games[game['port']] = game
    return games


if __name__ == '__main__':
    openings = [k for k, v in read_status().items() if waiting_for(v) > 0]
    print(openings)
