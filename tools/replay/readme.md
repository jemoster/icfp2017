# How to run

## Setup the playlog server
I typed these from memory so there are probably errors

    $ virtualenv server_env
    $ source server_env/bin/activate
    $ pip install -r requirements.txt
    
start the server

    $ python3 simple_server.py
    
## Setup the client server

    $ python3 -m http.server 8000
    
# Usage

Go to `http://localhost:8000/client/replay.html`

Enter the path to your file and click "connect"

# Recording a game

You can add the `--record` flag to the run tool:

```
python3 tools/bot_runner/online_adapter.py --record="./brownian-test" "./brownian" 9020
```
