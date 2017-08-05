#Online Adaptor

## Quick Start

From the root folder run:

    python3 tools/bot_runner/online_adapter.py "python -u src/randobot/main.py" 9007

`9007` is the port number to direct at.
See http://punter.inf.ed.ac.uk/status.html for port states.

For further usage instructions:

    python3 tools/bot_runner/online_adapter.py --help

## It failed! Why!?

 * When a game ends the server resets the port causing an error
 * If a game is in progress the script will appear to get stuck.


# Make Player Data
Requires the number of players and port to be specified.
After that a list of bots may be added.
Any shortage of bots will be filled with random selections unless the
`--default` option is specified.

    make_player_data.py 4 9053 --bot brown --bot random --default random

    make_player_data.py 4 9053 --bot "python -u src/randobot/main.py" --default random

