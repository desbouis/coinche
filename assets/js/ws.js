window.addEventListener("load", function(event) {

  var loc = window.location, ws_proto;
  var ws;
  var card_nb;
  var ws_action = {
    play_card: "PLAY_CARD",
    cancel_card: "CANCEL_CARD",
    pickup_cards: "PICKUP_CARDS"
  };
  var players_alias = ["Nord", "Est", "Sud", "Ouest"];

  if (loc.protocol === "https:") {
    ws_proto = "wss:";
  } else {
    ws_proto = "ws:";
  }

  ws = new WebSocket(ws_proto+"//"+loc.host+"/coinche/ws/"+document.getElementById("gameId").value);

  ws.onopen = function(event) {
    console.log("WEBSOCKET OPENED");
  };

  ws.onclose = function(event) {
    console.log("WEBSOCKET CLOSED");
    ws = null;
  };

  ws.onerror = function(event) {
    console.log("WEBSOCKET ERROR: " + event.data);
  };

  ws.onmessage = function(event) {
    var msg = JSON.parse(event.data);
    console.log("WEBSOCKET MESSAGE RECEIVED: " + event.data);
    displayPlayerCard(msg);
  };

  var displayPlayerCard = function(message) {
    // play or cancel a card
    if (document.getElementById("img"+message.player_alias+"Card")) {
      var old_img = document.getElementById("img"+message.player_alias+"Card");
      var elt_target = old_img.parentNode;
      var new_img = document.createElement("img");
      new_img.id = "img"+message.player_alias+"Card";
      new_img.className = "img-card";
      if (message.action == ws_action.play_card) {
        new_img.src = message.player_card_src;
        new_img.name = message.player_alias+"-"+message.player_card;
        elt_target.replaceChild(new_img, old_img);
      }
      if (message.action == ws_action.cancel_card) {
        new_img.src = "/coinche/assets/img/back.png";
        new_img.name = "";
        elt_target.replaceChild(new_img, old_img);
      }
    }
    // clean playmat and players cards when pickup cards
    if (message.action == ws_action.pickup_cards) {
      let elt_playmat = document.getElementById("playMat");
      while (elt_playmat.firstChild) {
        elt_playmat.removeChild(elt_playmat.firstChild);
      }
      for (let i =0; i < players_alias.length; i++) {
        let img = document.getElementById("img"+players_alias[i]+"Card");
        if (img) {
          img.src = "/coinche/assets/img/back.png";
          img.name = "";
        }
      }
    }
  };

  // when player is playing a card
  sortablePlayMat.options.onAdd = function(event) {
    if (!ws) {
      return false;
    }

    card_nb = (8 - document.getElementById("myCards").childElementCount).toString();
    var message = document.getElementById("playerName").value + " a joué " + event.item.name + " comme carte n°" + card_nb;
    console.log("playMat onAdd event: " + message);

    var msg = {
      action:          ws_action.play_card,
      message:         message,
      game_id:         document.getElementById("gameId").value,
      game_name:       document.getElementById("gameName").value,
      game_distrib_nb: document.getElementById("gameDistribNb").value,
      player_id:       document.getElementById("playerId").value,
      player_name:     document.getElementById("playerName").value,
      player_alias:    document.getElementById("playerAlias").value,
      player_team:     document.getElementById("playerTeam").value,
      player_card:     event.item.name,
      player_card_src: event.item.src,
      card_nb:         card_nb,
    };

    ws.send(JSON.stringify(msg));
    console.log("WEBSOCKET MESSAGE SENT: " + JSON.stringify(msg));

    return false;
  };

  // when player is canceling a card
  sortableMyCards.options.onAdd = function(event) {
    if (!ws) {
      return false;
    }

    var message = document.getElementById("playerName").value + " a annulé " + event.item.name + " comme carte n°" + card_nb;
    console.log("myCards onAdd event: " + message);

    var msg = {
      action:          ws_action.cancel_card,
      message:         message,
      game_id:         document.getElementById("gameId").value,
      game_name:       document.getElementById("gameName").value,
      game_distrib_nb: document.getElementById("gameDistribNb").value,
      player_id:       document.getElementById("playerId").value,
      player_name:     document.getElementById("playerName").value,
      player_alias:    document.getElementById("playerAlias").value,
      player_team:     document.getElementById("playerTeam").value,
      player_card:     "",
      player_card_src: "",
      card_nb:         card_nb,
    };

    ws.send(JSON.stringify(msg));
    console.log("WEBSOCKET MESSAGE SENT: " + JSON.stringify(msg));

    return false;
  };

  // when wining player picks up all played cards
  pickupPlayedCards = function() {
    if (!ws) {
      return false;
    }

    var message = document.getElementById("playerName").value + " a ramassé les cartes jouées pour l'équipe " + document.getElementById("playerTeam").value;
    console.log("pickupPlayedCards onClick event: " + message);

    var msg = {
      action:          ws_action.pickup_cards,
      message:         message,
      game_id:         document.getElementById("gameId").value,
      game_name:       document.getElementById("gameName").value,
      game_distrib_nb: document.getElementById("gameDistribNb").value,
      player_id:       document.getElementById("playerId").value,
      player_name:     document.getElementById("playerName").value,
      player_alias:    document.getElementById("playerAlias").value,
      player_team:     document.getElementById("playerTeam").value,
      player_card:     "",
      player_card_src: "",
      card_nb:         card_nb,
    };

    ws.send(JSON.stringify(msg));
    console.log("WEBSOCKET MESSAGE SENT: " + JSON.stringify(msg));

    return false;
  };

});
