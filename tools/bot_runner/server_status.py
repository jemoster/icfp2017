#!/usr/bin/env python3

import urllib.request
import io
import re
import random


def parse(line):
    blocks = re.split('<[^>]*>', line)
    status = blocks[2]
    players = blocks[4]
    map = blocks[-3]
    port = blocks[-6]
    return status, players, port, map


def read_status():
    PUNTER_STATUS = 'http://punter.inf.ed.ac.uk/status.html'

    with urllib.request.urlopen(PUNTER_STATUS) as response:
       html = response.read()

    buff = io.StringIO(html.decode())

    #discard the header
    for _ in range(18):
        buff.readline()

    statuses = []
    for line in buff:
        try:
            statuses.append(parse(line))
        except:
            pass

    return statuses


def waiting_for(entry):
    if 'Waiting for punters' in entry[0]:
        a, b = entry[0].split('Waiting for punters. (')[1].replace(')', '').split('/')
        return int(b)-int(a)
    return 0


def find_new_game():
    stats = read_status()

    for x in range(len(stats)):
        entry = random.choice(stats)
        open_spots = waiting_for(entry)
        if open_spots > 0:
            return entry

if __name__ == '__main__':
    print(find_new_game())