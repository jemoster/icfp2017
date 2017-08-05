#!/usr/bin/env python3

import urllib.request
import io
import re
import random


def parse(line):
    blocks = re.split('<[^>]*>', line)

    entry = {
        'status': blocks[2],
        'players': blocks[4],
        'map': blocks[-3],
        'port': blocks[-6],
        'waiting': 0,
    }
    if 'Waiting for punters' in entry['status']:
        a, b = entry['status'].split('Waiting for punters. (')[1].replace(')', '').split('/')
        entry['waiting'] = int(b)-int(a)

    return entry


def read_status():
    PUNTER_STATUS = 'http://punter.inf.ed.ac.uk/status.html'

    with urllib.request.urlopen(PUNTER_STATUS) as response:
       html = response.read()

    buff = io.StringIO(html.decode())

    #discard the header
    for _ in range(18):
        buff.readline()

    statuses = {}
    for line in buff:
        try:
            s = parse(line)
            statuses[s['port']] = s
        except:
            pass

    return statuses



