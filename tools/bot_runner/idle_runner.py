#!/usr/bin/env python3

import os
import random
import signal
import time
from server_status import read_status, waiting_for
from adapter import OfflineAdapter
from make_player_data import server, known_bots


def idle_run():
    ''' Play a random game against an opening '''
    statuses = read_status()
    openings = [k for k, v in statuses.items() if waiting_for(v) > 0]
    choice = random.choice(openings)
    status = statuses[choice]
    port = int(choice)
    print('Choosing game:', int(port))
    bot_name = random.choice(list(known_bots.keys()))
    bot_exe = known_bots[bot_name]
    print('Choosing bot:', bot_name, bot_exe)
    record = os.path.join('data', 'idle', str(int(time.time())))
    print('Record path:', record)
    header = {'status': status,
              'choice': choice,
              'record': record}
    print('Header:', header)
    adapter = OfflineAdapter(server, port, bot_exe, record, header)
    scores = adapter.run()
    for player in scores:
        player_name = 'punter:' if player['punter'] != adapter.punter_id else "me:    "
        print('{} {punter}, score: {score}'.format(player_name, **player))


if __name__ == '__main__':
    global running
    running = True

    def handler(signum, frame):
        global running
        print('Signal handler called with signal', signum)
        running = False

    print('registering signal handler')
    signal.signal(signal.SIGUSR1, handler)

    print('Idle runner starting')
    while running:
        try:
            idle_run()
        except Exception as e:
            print('Well that was bad')
            print(e)
            print('Spending some time in timeout')
            time.sleep(10)
        print('Sleeping inbetween games')
        time.sleep(5)
