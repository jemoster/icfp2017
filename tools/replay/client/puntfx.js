/* GLOBALS */
let punterID = -1;
let numPunters = -1;
let initialised = false;

let queuedClaims = [];
let queuedPass = false;

// const hostname = "129.215.197.1";
const hostname = "127.0.0.1";
const relayPort = 5000;

/* Graph rendering */

const colours =
  [
    "#1f77b4",
    "#ff7f0e",
    "#e377c2",
    "#bcbd22",
    "#d62728",
    "#17becf",
    "#8c564b",
   "#2ca02c",
    "#9467bd",
   "#98df8a",
   "#ff9896",
   "#c5b0d5",
    "#aec7e8",
   "#c49c94",
    "#ffbb78",
   "#f7b6d2",
   "#7f7f7f",
   "#c7c7c7",
   "#dbdb8d",
   "#9edae5"
  ];

function getPunterColour(punter) {
  return colours[punter % colours.length];
}

function renderGraph(punterID, graph) {
    initCy(graph,
           function() {
               initialised = true;
               cy.autolock(true);
               bindCoreHandlers();
               if (queuedClaims.length > 0 || queuedPass) {
                   playQueuedClaims();
                   ourTurn();
               } else {
                   theirTurn();
               }
           }
          );
    return;
}

function toggleButton(buttonId, st) {
  $("#" + buttonId).attr("disabled", st);
}

function disableButton(buttonId) {
  toggleButton(buttonId, true);
}

function enableButton(buttonId) {
  toggleButton(buttonId, false);
}


$(function() {
  $(document).ready(function() {
    enableButton('connect');
  });
});


/**
 * Communication
 */
function _connect() {
  disableButton('connect');
  enableButton('disconnect');
  const gamePort = $('#gamePort').val();
  const punterName = $('#punterName').val();
  connect(gamePort, punterName);
  return;
}

function _disconnect() {
  disableButton('disconnect');
  enableButton('connect');
  disconnect();
  return;
}

let socket = undefined;

function setStatus(status) {
  $("#game-status").text(status);
}

function writeLog(msg) {
    let id = "log";
    let now = new Date(new Date().getTime()).toLocaleTimeString();
    document.getElementById(id).innerHTML += "(" + now + ") " + msg + "\n";
    document.getElementById(id).scrollTop = document.getElementById(id).scrollHeight;
    return;
}

function logInfo(msg) {
    writeLog("info: " + msg);
    return;
}

function logClaim(claim) {
    writeLog("move: punter #" + claim.punter + " claimed edge " +
             claim.source + " -- " + claim.target + ".");
    return;
}

function logPass(pass) {
    writeLog("pass: punter #" + pass.punter + ".");
    return;
}

function logScore(punter_id, score) {
  writeLog("punter #" + punter_id + " scored " + score);
}

function logMove(move) {
  if (move.claim != undefined) {
    logClaim(move.claim);
  } else if (move.pass != undefined) {
    logPass(move.pass);
  }
}

function logError(msg) {
    writeLog("error: " + msg);
    return;
}

function logRelay(msg) {
    writeLog("relay: " + msg);
    return;
}

function connect(gamePort, punterName) {
    let graph = undefined;
    let ws_uri = "ws://" + hostname + ":" + relayPort;
    logInfo("connecting to relay [" + ws_uri + "]...");
    
    let punterID = null;

    socket = new WebSocket(ws_uri);

    socket.onopen = function(data) {
        logInfo("connection established.");
        setStatus("Connected; waiting for other punters...");
        socket.send(hostname + ":" + gamePort + ":" + punterName);
        return;
    };

    socket.onclose = function(data) {
        logInfo("connection closed by relay.");
        enableButton('connect');
        disableButton('disconnect');
        return;
    }

    socket.onerror = function(err) {
        if (socket.readyState === 3) {
            logError("connection failed.");
        } else {
            logError("connection failure.");
        }
        return;
    };

    socket.onmessage = function(message) {
        try {
            let msg = JSON.parse(message.data.split(/:(.+)/)[1]);

            logInfo("parse it!");
            logInfo(msg);
            // Initial message
            if (msg.map !== undefined) {
                // Record our ID, and the number of punters
                punterID = msg.punter;
                numPunters = msg.punters;

                logInfo("our punter ID: " + punterID + " color: " + getPunterColour(punterID));
                $("#punter-color").css('width', '60px');
                $("#punter-color").css('height', '60px');
                $("#punter-color").css('background-color', getPunterColour(punterID));

                logInfo("number of punters: " + numPunters);
                logInfo("received initial game graph: " + JSON.stringify(msg.map));
                graph = { "nodes": msg.map.sites,
                          "edges": msg.map.rivers,
                          "mines": msg.map.mines };
                logInfo("rendering game graph...");
                renderGraph(punterID, msg.map);
            } else if (msg.move !== undefined) {
                handleIncomingMoves(msg.move.moves);
            } else if (msg.stop !== undefined) {
                handleIncomingMoves(msg.stop.moves);
                printFinalScores(msg.stop.scores);
            } else {
                logError("unknown JSON message: " + message.data);
            }
        } catch (e) { // other message from the server
            logInfo('not the first message');
            logInfo(e);
            console.log(e);
            if (message.data.constructor == String) {
                logRelay(message.data);
            } else {
                logError("received unknown message from relay.");
            }
        }
        return;
    };
    return;
}

function disconnect() {
    socket.close();
    logInfo("disconnected.");
    socket = undefined;
    graph = undefined;
    return;
}

function send(json) {
  const str = JSON.stringify(json);
  socket.send(str.length + ":" + str);
}

function sendClaim(source, target) {
  const req = {
    claim: {
      punter: punterID,
      source: source,
      target: target
    }
  };

  send(req);
}

function sendPass() {
  const req = {
    pass: { punter: punterID }
  };

  send(req);
}

/* EVENT HANDLING LOGIC */

function handleEdgeClick(edge) {
  const source = edge.data("source");
  const target = edge.data("target");

  console.log("edge data; " + edge.data("owner"));
  if (edge.data("owner") == undefined) {
    sendClaim(parseInt(source), parseInt(target));
    cy.edges().unselect();
    updateEdgeOwner(punterID, source, target);
    theirTurn();
  } else {
    logError("That edge is already claimed! (" + source + " -- " + target + ")");
  }
}

function handlePass() {
  sendPass();
  writeLog("Passed!");
  theirTurn();
}

function bindCoreHandlers() {
  cy.edges().on("mouseover", function(evt) {
    this.style("content", this.data("owner"));
  });
  cy.edges().on("mouseout", function(evt) {
    this.style("content", "");
  });
}

function bindOurTurnHandlers() {
  cy.edges().off("select");
  cy.edges().on("select", function(evt) { handleEdgeClick(this) } );
  $("#pass-button").removeAttr("disabled");
}

function bindTheirTurnHandlers() {
  cy.edges().off("select");
  cy.edges().on("select", function(evt) {
    logError("Can't select an edge when it's not your turn to move!");
    cy.edges().unselect();
  } );
  $("#pass-button").attr("disabled", true);
}

function ourTurn() {
  bindOurTurnHandlers();
  setStatus("Your move!");
}

function theirTurn() {
  bindTheirTurnHandlers();
  setStatus("Waiting for others to make a move...");
}


/* GAME UPDATE LOGIC */

function normaliseEdgeData(edgeData) {
  const src = edgeData.source;
  const trg = edgeData.target;
  if (trg < src) {
    let tmp = edgeData["source"];
    edgeData["source"] = edgeData["target"];
    edgeData["target"] = tmp;
  }
}

function updateEdgeOwner(punter, source, target) {
  const es = cy.edges("[source=\"" + source + "\"][target=\"" + target + "\"]");
  if (es.length > 0) {
    const e = es[0];
    e.data()["owner"] = punter;
    e.animate({
      style: {
        "line-color": getPunterColour(punter),
        width: 30
      }
    }, {
      duration: 220,
      easing: 'ease-in-quad'
    });
    //e.style("line-color", getPunterColour(punter));

  } else {
    logError("Trying to update nonexistent edge! (" + source + " -- " + target + ")");
  }
}

function printFinalScores(scores) {
  logInfo("Game finished!");
  for (let i = 0; i < scores.length; i++) {
    logScore(scores[i].punter, scores[i].score);
  }
}

function handleIncomingMoves(moves) {
  for (let i = 0; i < moves.length; i++) {
    handleIncomingMove(moves[i]);
  }

  if (initialised) {
    ourTurn();
  }
}

function handleIncomingMove(move) {
  logMove(move);
  if (move.claim !== undefined) {
    const claim = move.claim;
    normaliseEdgeData(claim);
    if (initialised) {
      updateEdgeOwner(claim.punter, claim.source, claim.target);
    } else {
      queueClaim(claim);
    }
  } else if (move.pass !== undefined) {
    if (!initialised) {
      queuedPass = true;
    }
  }
}

function queueClaim(claim) {
  queuedClaims.push(claim);
}

function playQueuedClaims() {
  for (let i = 0; i < queuedClaims.length; i++) {
    const claim = queuedClaims[i];
    updateEdgeOwner(claim.punter, claim.source, claim.target);
  }
  queuedClaims = [];
  queuedPass = false;
  ourTurn();
}
