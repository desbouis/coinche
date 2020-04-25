package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "regexp"
    "strings"
    "time"

    "github.com/google/uuid"
)

type Player struct {
    Id               string
    Name             string
    Alias            string
    GameId           string
    Cards            map[string]string
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
    Cards         map[string]string
    ShuffledCards []string
}
/*
var baseColors = map[string]string{
    "heart"   : "\U00002665",
    "spade"   : "\U00002660",
    "diamond" : "\U00002666",
    "club"    : "\U00002663",
}
*/
/*
var imgColors = map[string]string{
    "heart"   : "<img src=\"/coinche/assets/img/heart.png\" width=\"30\" height=\"30\">",
    "spade"   : "<img src=\"/coinche/assets/img/spade.png\" width=\"30\" height=\"30\">",
    "diamond" : "<img src=\"/coinche/assets/img/diamond.png\" width=\"30\" height=\"30\">",
    "club"    : "<img src=\"/coinche/assets/img/club.png\" width=\"30\" height=\"30\">",
}
var baseCards = []string{"R", "D", "V", "A", "10", "9", "8", "7"}
*/
var imgColors = map[string]string{
    "heart"   : "h",
    "spade"   : "s",
    "diamond" : "d",
    "club"    : "c",
}
var baseCards = []string{"K", "Q", "J", "A", "10", "9", "8", "7"}
var explicitCards = make(map[string]string, 32)
var simpleCards []string

func initCards() {
    for key_color, val_color := range imgColors {
        for _, val_card := range baseCards {
//            explicitCards[key_color + val_card] = val_color + val_card
//            simpleCards = append(simpleCards, key_color + val_card)
            explicitCards[key_color + val_card] = val_card + val_color
            simpleCards = append(simpleCards, key_color + val_card)
        }
    }
}

func shuffleCards(c []string) []string {
    rand.Seed(time.Now().UnixNano())
    rand.Shuffle(len(c), func(i, j int) {
        c[i], c[j] = c[j], c[i]
    })
    return c
}

func generateId() string {
    return strings.Split(uuid.New().String(), "-")[0]
}

func (p *Player) savePlayer() error {
    filename := "p_" + p.Id + ".json"
    p.Cards = explicitCards
    jsonpayload, err := json.MarshalIndent(p, "", "  ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filename, jsonpayload, 0600)
}

func loadPlayer(id string) (*Player, error) {
    var p Player
    filename := "p_" + id + ".json"
    jsonpayload, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    err = json.Unmarshal(jsonpayload, &p)
    if err != nil {
        return nil, err
    }
    return &p, nil
}

func (g *Game) saveGame() error {
    filename := "g_" + g.Id + ".json"
    g.Cards = explicitCards
    jsonpayload, err := json.MarshalIndent(g, "", "  ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filename, jsonpayload, 0600)
}

func loadGame(id string) (*Game, error) {
    var g Game
    filename := "g_" + id + ".json"
    jsonpayload, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    err = json.Unmarshal(jsonpayload, &g)
    if err != nil {
        return nil, err
    }
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
        pNord.DistributedCards[v] = g.Cards[v]
    }
    err = pNord.savePlayer()
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
    for _, v := range shuffledCards[8:16] {
        pSud.DistributedCards[v] = g.Cards[v]
    }
    err = pSud.savePlayer()
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
    for _, v := range shuffledCards[16:24] {
        pEst.DistributedCards[v] = g.Cards[v]
    }
    err = pEst.savePlayer()
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
        pOuest.DistributedCards[v] = g.Cards[v]
    }
    err = pOuest.savePlayer()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/coinche/game/"+gameId, http.StatusFound)
}

func playerViewHandler(w http.ResponseWriter, r *http.Request, id string) {
    p, err := loadPlayer(id)
    if err != nil {
        http.Redirect(w, r, "/coinche/", http.StatusFound)
        return
    }
    playerRenderTemplate(w, "player/view", p)
}
/*
var templates = template.Must(template.ParseFiles("templates/index.html",
                                                  "templates/game/edit.html",
                                                  "templates/game/view.html",
                                                  "templates/player/view.html"))
*/

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

func playerRenderTemplate(w http.ResponseWriter, tmpl string, p *Player) {
    err := templates.ExecuteTemplate(w, tmpl, p)
    if err != nil {
        log.Fatalf("Template execution failed!")
        http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
        return
    }
}

var validUrl = regexp.MustCompile("^/coinche/(game|game/edit|player)/([a-zA-Z0-9]+)$")

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

func main() {
    fmt.Println("Starting...")
    initCards()

    fs := http.FileServer(http.Dir("./assets/"))
    http.Handle("/coinche/assets/", http.StripPrefix("/coinche/assets/", fs))

    http.HandleFunc("/coinche/", indexHandler)
    http.HandleFunc("/coinche/game/", makeHandler(gameViewHandler))
    http.HandleFunc("/coinche/game/edit/", makeHandler(gameEditHandler))
    http.HandleFunc("/coinche/game/save", gameSaveHandler)
    http.HandleFunc("/coinche/game/distribute", gameDistributeHandler)
    http.HandleFunc("/coinche/player/", makeHandler(playerViewHandler))

    log.Fatal(http.ListenAndServe(":8080", nil))
}