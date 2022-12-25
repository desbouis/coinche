package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "math/rand"
    "net/http"
    "regexp"
    "runtime"
    "strings"
    "time"

    "github.com/gomodule/redigo/redis"
    "github.com/google/uuid"
    "github.com/gorilla/websocket"
)

type Player struct {
    Id               string
    Name             string
    Alias            string
    GameId           string
    DistributedCards map[string]string
}

type Game struct {
    Id            string
    Name          string
    NordId        string
    NordName      string
    SudId         string
    SudName       string
    EstId         string
    EstName       string
    OuestId       string
    OuestName     string
    ShuffledCards []string
}

// struct to give to the player template
type ViewPlayerData struct {
    Player *Player
    Game   *Game
}

const gamePrefix   string = "game/"
const playerPrefix string = "player/"

var imgColors = map[string]string{
    "heart"   : "h",
    "spade"   : "s",
    "diamond" : "d",
    "club"    : "c",
}
var baseCards = []string{"K", "Q", "J", "A", "10", "9", "8", "7"}
var refCards = make(map[string]string, 32)
var simpleCards []string

var (
    redCon redis.Conn
    err error
    reply interface{}
)

/*
 * websocket
 */
// connected clients
// the string will contain the GameId
var wsClientsRegistry = make(map[*websocket.Conn]string)
// broadcast channel
var wsBroadcast = make(chan WsMessage)
// Configure the upgrader
var wsUpgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}
// Define our message object sent and received in websocket
type WsMessage struct {
    GameId        string `json:"game_id"`
    GameName      string `json:"game_name"`
    PlayerId      string `json:"player_id"`
    PlayerName    string `json:"player_name"`
    PlayerAlias   string `json:"player_alias"`
    PlayerCard    string `json:"player_card"`
    PlayerCardSrc string `json:"player_card_src"`
    Action        string `json:"action_type"`
    Message       string `json:"message"`
}



/* functions */

func trace() {
    pc := make([]uintptr, 15)
    n := runtime.Callers(2, pc)
    frames := runtime.CallersFrames(pc[:n])
    frame, _ := frames.Next()
    fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}

func initCards() {
    for key_color, val_color := range imgColors {
        for _, val_card := range baseCards {
            refCards[key_color + val_card] = val_card + val_color
            simpleCards = append(simpleCards, key_color + val_card)
        }
    }
    if _, err := redCon.Do("HMSET", redis.Args{}.Add("refCards").AddFlat(refCards)...); err != nil {
        fmt.Println(err)
        return
    }
}

func shuffleCards(c []string) []string {
    rand.Seed(time.Now().UnixNano())
    rand.Shuffle(len(c), func(i, j int) {
        c[i], c[j] = c[j], c[i]
    })
    return c
}

func initRedis() {
    redCon, err = redis.Dial("tcp", ":6379")
    if err != nil {
        log.Fatal(err)
    }
}

func generateId() string {
    return strings.Split(uuid.New().String(), "-")[0]
}

func (p *Player) savePlayer() error {
    jsonPayload, err := json.Marshal(p)
    if err != nil {
        fmt.Println(err)
        return err
    }
    _, err = redCon.Do("SET", playerPrefix+p.Id, jsonPayload)
    if err != nil {
        fmt.Println(err)
        return err
    }
    return nil
}

func loadPlayer(id string) (*Player, error) {
    var p Player
    jsonPayload, err := redis.String(redCon.Do("GET", playerPrefix+id))
    if err == redis.ErrNil {
        fmt.Printf("Player %s does not exist", id)
    } else if err != nil {
        fmt.Println(err)
        return nil, err
    }
    err = json.Unmarshal([]byte(jsonPayload), &p)
    if err != nil {
        return nil, err
    }
    fmt.Println(p)
    return &p, nil
}

func (g *Game) saveGame() error {
    // cf. https://github.com/gilcrest/redigo-example/blob/master/main.go
    jsonPayload, err := json.Marshal(g)
    if err != nil {
        fmt.Println(err)
        return err
    }
    _, err = redCon.Do("SET", gamePrefix+g.Id, jsonPayload)
    if err != nil {
        fmt.Println(err)
        return err
    }
    return nil
}

func loadGame(id string) (*Game, error) {
    var g Game
    jsonPayload, err := redis.String(redCon.Do("GET", gamePrefix+id))
    if err == redis.ErrNil {
        fmt.Printf("Game %s does not exist", id)
    } else if err != nil {
        fmt.Println(err)
        return nil, err
    }
    err = json.Unmarshal([]byte(jsonPayload), &g)
    if err != nil {
        return nil, err
    }
    fmt.Println(g)
    return &g, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    g := &Game{}
    gameRenderTemplate(w, "index", g)
}

func gameViewHandler(w http.ResponseWriter, r *http.Request, id string) {
    g, err := loadGame(id)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    gameRenderTemplate(w, "game/view", g)
}

func gameEditHandler(w http.ResponseWriter, r *http.Request, id string) {
    g, err := loadGame(id)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    gameRenderTemplate(w, "game/edit", g)
}

func gameSaveHandler(w http.ResponseWriter, r *http.Request) {
    var err error = nil
    gameId      := r.FormValue("gameId")
    gameName    := r.FormValue("gameName")
    nordId      := r.FormValue("nordId")
    nordName    := r.FormValue("nordName")
    sudId       := r.FormValue("sudId")
    sudName     := r.FormValue("sudName")
    estId       := r.FormValue("estId")
    estName     := r.FormValue("estName")
    ouestId     := r.FormValue("ouestId")
    ouestName   := r.FormValue("ouestName")
    if gameId == "" {
        gameId = generateId()
    }
    if nordId == "" {
        nordId = generateId()
    }
    if sudId == "" {
        sudId = generateId()
    }
    if estId == "" {
        estId = generateId()
    }
    if ouestId == "" {
        ouestId = generateId()
    }
    g := &Game{Id: gameId,
               Name: gameName,
               NordId: nordId,
               NordName: nordName,
               SudId: sudId,
               SudName: sudName,
               EstId: estId,
               EstName: estName,
               OuestId: ouestId,
               OuestName: ouestName}
    err = g.saveGame()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    pNord  := &Player{Id: nordId,
                      Name: nordName,
                      Alias: "Nord",
                      GameId: gameId}
    err = pNord.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    pSud   := &Player{Id: sudId,
                      Name: sudName,
                      Alias: "Sud",
                      GameId: gameId}
    err = pSud.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    pEst   := &Player{Id: estId,
                      Name: estName,
                      Alias: "Est",
                      GameId: gameId}
    err = pEst.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    pOuest := &Player{Id: ouestId,
                      Name: ouestName,
                      Alias: "Ouest",
                      GameId: gameId}
    err = pOuest.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/coinche/game/"+gameId, http.StatusFound)
}

func gameDistributeHandler(w http.ResponseWriter, r *http.Request) {
    gameId      := r.FormValue("gameId")
    nordId      := r.FormValue("nordId")
    sudId       := r.FormValue("sudId")
    estId       := r.FormValue("estId")
    ouestId     := r.FormValue("ouestId")

    shuffledCards := shuffleCards(simpleCards)

    g, err := loadGame(gameId)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    g.ShuffledCards = shuffledCards
    err = g.saveGame()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    pNord, err := loadPlayer(nordId)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    pNord.DistributedCards = make(map[string]string, 8)
    for _, v := range shuffledCards[:8] {
        pNord.DistributedCards[v] = refCards[v]
    }
    err = pNord.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    pEst, err := loadPlayer(estId)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    pEst.DistributedCards = make(map[string]string, 8)
    for _, v := range shuffledCards[8:16] {
        pEst.DistributedCards[v] = refCards[v]
    }
    err = pEst.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    pSud, err := loadPlayer(sudId)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    pSud.DistributedCards = make(map[string]string, 8)
    for _, v := range shuffledCards[16:24] {
        pSud.DistributedCards[v] = refCards[v]
    }
    err = pSud.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    pOuest, err := loadPlayer(ouestId)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    pOuest.DistributedCards = make(map[string]string, 8)
    for _, v := range shuffledCards[24:] {
        pOuest.DistributedCards[v] = refCards[v]
    }
    err = pOuest.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/coinche/game/"+gameId, http.StatusFound)
}

func playerViewHandler(w http.ResponseWriter, r *http.Request, id string) {
    // get player data
    p, err := loadPlayer(id)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    // get game data
    g, err := loadGame(p.GameId)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    data := &ViewPlayerData{p, g}
    playerRenderTemplate(w, "player/view", data)
}

func noescape(s string) template.HTML {
    return template.HTML(s)
}

var templates = template.Must(template.New("").Funcs(template.FuncMap{
    "noescape": noescape,
}).ParseFiles("templates/index.html",
              "templates/game/edit.html",
              "templates/game/view.html",
              "templates/player/view.html"))

func gameRenderTemplate(w http.ResponseWriter, tmpl string, g *Game) {
    err := templates.ExecuteTemplate(w, tmpl, g)
    if err != nil {
        log.Fatalf("Template execution failed!")
        http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
        return
    }
}

func playerRenderTemplate(w http.ResponseWriter, tmpl string, data *ViewPlayerData) {
    err := templates.ExecuteTemplate(w, tmpl, data)
    if err != nil {
        log.Fatalf("Template execution failed!")
        http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
        return
    }
}

var validUrl = regexp.MustCompile("^/coinche/(game|game/edit|player|ws)/([a-zA-Z0-9]+)$")

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Here we will extract the game id from the Request, and call the provided handler 'fn'
        m := validUrl.FindStringSubmatch(r.URL.Path)
        if m == nil {
            log.Fatalf("Url validation failed!")
            http.NotFound(w, r)
            return
        }
        // The id is the second subexpression
        fn(w, r, m[2])
    }
}

func wsConnectionsHandler(w http.ResponseWriter, r *http.Request, id string) {
    wsUpgrader.CheckOrigin = func(r *http.Request) bool {
        return true
    }
    // Upgrade initial GET request to a websocket
    ws, err := wsUpgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    // Make sure we close the connection when the function returns
    defer ws.Close()
    // Register our new client
    wsClientsRegistry[ws] = id
    log.Printf("websocket[game %s]: new client connected and registered!", id)
    // inifinite loop that continuously waits for a new message to be written to the WebSocket,
    // unserializes it from JSON to a Message object
    // and then throws it into the broadcast channel
    for {
        var msg WsMessage
        // Read in a new message as JSON and map it to a WsMessage object
        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Printf("error: %v", err)
            trace()
            delete(wsClientsRegistry, ws)
            break
        }
        // Send the newly received message to the wsBroadcast channel
        log.Printf("Websocket[game %s]: receive and broadcast the message %s", id, msg)
        wsBroadcast <- msg
    }
}

func wsMessagesHandler() {
    for {
        // Grab the next message from the broadcast channel
        msg := <-wsBroadcast
        // Send it out to every clients that are currently connected at the same game
        for client, gameId := range wsClientsRegistry {
            if (gameId == msg.GameId) {
                log.Printf("Websocket[game %s]: send the message %s to client", gameId, msg)
                err := client.WriteJSON(msg)
                if err != nil {
                    log.Printf("error: %v", err)
                    trace()
                    client.Close()
                    delete(wsClientsRegistry, client)
                }
            }
        }
    }
}

func main() {
    log.Println("Starting Coinche app...")

    log.Println("Initializing Redis connection...")
    initRedis()
    log.Println("...Redis connection initialized!")

    log.Println("Initializing cards...")
    initCards()
    log.Println("...cards initialized!")

    fs := http.FileServer(http.Dir("./assets/"))
    http.Handle("/coinche/assets/", http.StripPrefix("/coinche/assets/", fs))

    http.HandleFunc("/coinche/", indexHandler)
    http.HandleFunc("/coinche/game/", makeHandler(gameViewHandler))
    http.HandleFunc("/coinche/game/edit/", makeHandler(gameEditHandler))
    http.HandleFunc("/coinche/game/save", gameSaveHandler)
    http.HandleFunc("/coinche/game/distribute", gameDistributeHandler)
    http.HandleFunc("/coinche/player/", makeHandler(playerViewHandler))
    http.HandleFunc("/coinche/ws/", makeHandler(wsConnectionsHandler))

    // Start listening for incoming websocket messages
    go wsMessagesHandler()

    log.Println("Starting http server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
