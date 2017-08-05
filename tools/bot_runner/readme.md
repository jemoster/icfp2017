#Online Adaptor

## Quick Start

From the root folder run:

    python3 tools/bot_runner/online_adapter.py "python -u src/randobot/main.py" 9007
    
`9007` is the port number to direct at. 
See http://punter.inf.ed.ac.uk/status.html for port states.

For further usage instructions:

    python3 main.py --help
    
## It failed! Why!?

 * When a game ends the server resets the port causing an error
 * If a game is in progress the script will appear to get stuck.
    
