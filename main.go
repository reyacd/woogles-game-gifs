package main

import (
    "bytes"
    "context"
    "fmt"
    "image"
    "image/color/palette"
    "image/draw"
    "image/gif"
    "image/png"
    "io/ioutil"
    "net/http"
    "os"
    "strings"

    "github.com/domino14/macondo/gcgio"
    "github.com/domino14/macondo/config"

    pb "github.com/domino14/liwords/rpc/api/proto/game_service"
    macondopb "github.com/domino14/macondo/gen/api/proto/macondo"
)

var boardConfig = []string{
    "=  '   =   '  =",
    " -   \"   \"   - ",
    "  -   ' '   -  ",
    "'  -   '   -  '",
    "    -     -    ",
    " \"   \"   \"   \" ",
    "  '   ' '   '  ",
    "=  '   *   '  =",
    "  '   ' '   '  ",
    " \"   \"   \"   \" ",
    "    -     -    ",
    "'  -   '   -  '",
    "  -   ' '   -  ",
    " -   \"   \"   - ",
    "=  '   =   '  =",
}

var tileSrc = map[byte][2]int{
    'A': {0, 0}, 'B': {0, 1}, 'C': {0, 2}, 'D': {0, 3}, 'E': {0, 4},
    'F': {0, 5}, 'G': {0, 6}, 'H': {0, 7}, 'I': {0, 8}, 'J': {0, 9},
    'K': {1, 0}, 'L': {1, 1}, 'M': {1, 2}, 'N': {1, 3}, 'O': {1, 4},
    'P': {1, 5}, 'Q': {1, 6}, 'R': {1, 7}, 'S': {1, 8}, 'T': {1, 9},
    'U': {2, 0}, 'V': {2, 1}, 'W': {2, 2}, 'X': {2, 3}, 'Y': {2, 4},
    'Z': {2, 5}, 'a': {2, 6}, 'b': {2, 7}, 'c': {2, 8}, 'd': {2, 9},
    'e': {3, 0}, 'f': {3, 1}, 'g': {3, 2}, 'h': {3, 3}, 'i': {3, 4},
    'j': {3, 5}, 'k': {3, 6}, 'l': {3, 7}, 'm': {3, 8}, 'n': {3, 9},
    'o': {4, 0}, 'p': {4, 1}, 'q': {4, 2}, 'r': {4, 3}, 's': {4, 4},
    't': {4, 5}, 'u': {4, 6}, 'v': {4, 7}, 'w': {4, 8}, 'x': {4, 9},
    'y': {5, 0}, 'z': {5, 1}, '?': {5, 2},
}

var boardSrc = map[byte][2]int{
    '-': {5, 3}, '=': {5, 4},
    '\'': {5, 5}, '"': {5, 6}, '*': {5, 7}, ' ': {5, 8},
}

// Doubled because of retina screen.
const squareDim = 2 * 34

func loadTilesImg() (image.Image, error) {

    tilesBytes, err := ioutil.ReadFile("data/tiles.png")
    if err != nil {
        return nil, err
    }
    img, err := png.Decode(bytes.NewReader(tilesBytes))
    if err != nil {
        return nil, err
    }
    bounds := img.Bounds()
    expectedX := 10 * squareDim
    expectedY := 6 * squareDim
    if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Dx() != expectedX || bounds.Dy() != expectedY {
        return nil, fmt.Errorf("unexpected size: %s vs %s", bounds.String(), image.Pt(expectedX, expectedY))
    }
    return img, nil
}

func AnimateGame(tilesImg image.Image, boardConfig []string, hist *macondopb.GameHistory) (*gif.GIF, error) {

    img := image.NewPaletted(image.Rect(0, 0, 15*squareDim, 15*squareDim), palette.Plan9)
    gameGif := &gif.GIF{}

    // Draw the board.
    for r := 0; r < 15; r++ {
        y := r * squareDim
	for c := 0; c < 15; c++ {
	    x := c * squareDim
            b := boardConfig[r][c]
            srcPt, ok := boardSrc[b]
            if !ok {
                srcPt = boardSrc[' ']
            }
            draw.Draw(img, image.Rect(x, y, x+squareDim, y+squareDim), tilesImg,
                      image.Pt(srcPt[1]*squareDim, srcPt[0]*squareDim), draw.Over)
        }
    }
    gameGif.Image = append(gameGif.Image, img) 
    gameGif.Delay = append(gameGif.Delay, 100)

    prevImg := img
    for i := range hist.Events {
	evtImg := image.NewPaletted(prevImg.Bounds(), palette.Plan9)
        draw.Draw(evtImg, evtImg.Bounds(), prevImg, image.Pt(0, 0), draw.Over)
        removePhony, err := drawEvent(*hist.Events[i], evtImg, tilesImg)
        if err != nil {
            return &gif.GIF{}, fmt.Errorf("Error drawing event: %v", err)
        } 
        if removePhony {
            draw.Draw(evtImg, evtImg.Bounds(), gameGif.Image[i-1], image.Pt(0, 0), draw.Over)
        }
        gameGif.Image = append(gameGif.Image, evtImg)
        gameGif.Delay = append(gameGif.Delay, 100)
        prevImg = evtImg
    }
    return gameGif, nil
}

func drawPlay(evt macondopb.GameEvent, boardImg *image.Paletted, tilesImg image.Image) {

    var right int
    var down int

    if evt.Direction ==  macondopb.GameEvent_HORIZONTAL {
        right = len(evt.PlayedTiles)
        down = 1
    } else {
        right = 1
        down = len(evt.PlayedTiles)
    }

    row := int(evt.GetRow())
    column := int(evt.GetColumn())
    nRows := row + down
    nCols := column + right

    idx := 0
    fmt.Printf("%s %s\n", evt.Position, evt.PlayedTiles)
    for i := row; i < nRows; i++ {
        y := i * squareDim
        for j := column; j < nCols; j++ {
            x := j * squareDim 
            letter := evt.PlayedTiles[idx] 
            if letter != '.' {
                srcPt := tileSrc[letter]
                draw.Draw(boardImg, image.Rect(x, y, x+squareDim, y+squareDim), tilesImg,
                          image.Pt(srcPt[1]*squareDim, srcPt[0]*squareDim), draw.Over)
            } 
            idx++
        } 
    }
}

func drawEvent(evt macondopb.GameEvent, boardImg *image.Paletted, tilesImg image.Image) (bool, error) {

    evtType := evt.GetType()
    removePhony := false

    switch evtType {
    case macondopb.GameEvent_TILE_PLACEMENT_MOVE:
        fmt.Printf("Tile Placement Play ") 
        drawPlay(evt, boardImg, tilesImg)
    case macondopb.GameEvent_PHONY_TILES_RETURNED:
        fmt.Printf("Phony tiles returned!\n") 
        removePhony = true
    case macondopb.GameEvent_PASS:
        fmt.Printf("Pass.\n") 
    case macondopb.GameEvent_CHALLENGE_BONUS:
        fmt.Printf("Challenge bonus.\n") 
    case macondopb.GameEvent_END_RACK_PTS:
        fmt.Printf("End rack points.\n") 
    case macondopb.GameEvent_EXCHANGE:
        fmt.Printf("Exchange %v\n", evt.Exchanged) 
    case macondopb.GameEvent_END_RACK_PENALTY:
        fmt.Printf("End rack penalty.\n") 
    case macondopb.GameEvent_TIME_PENALTY:
        fmt.Printf("Time penalty.\n") 
    case macondopb.GameEvent_UNSUCCESSFUL_CHALLENGE_TURN_LOSS:
        fmt.Printf("Unsuccessful challenge turn loss.\n") 

    default:
        return removePhony, fmt.Errorf("event type %v not supported", evtType)

    }
    return removePhony, nil
}

func GetGameHistory(id string) (*macondopb.GameHistory, error) {

    client := pb.NewGameMetadataServiceProtobufClient("https://woogles.io", &http.Client{})
    history, err := client.GetGameHistory(context.Background(), &pb.GameHistoryRequest{GameId: id})

    if err != nil {
        return &macondopb.GameHistory{}, err
    }
    return history.History, nil
}

func GetCrossTablesGcg(id string) (string, error) {
    crossTablesUrl := fmt.Sprintf("https://cross-tables.com/annotated/selfgcg/%v/anno%v.gcg", id[:3], id) 
    response, err := http.Get(crossTablesUrl)
    gcg, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", fmt.Errorf("Error reading response from cross-tables.")
    }
    return string(gcg), nil
}

func main() {

    // Cache this.
    tilesImg, err := loadTilesImg()
    if err != nil {
        panic(err)
    }

    gameHistory := &macondopb.GameHistory{}
    var id string
    if strings.Contains(os.Args[1], "ct:") {
        id = strings.Split(os.Args[1], ":")[1]
        gcg, err := GetCrossTablesGcg(id)
        if err != nil {
            fmt.Errorf("Unable to look up cross-tables GCG for provided ID: %v. Received error: %v", id, err)
        }
        cfg := &config.Config{ StrategyParamsPath: "data/strategy",
                               LexiconPath: "data/lexica",
                               LetterDistributionPath: "data/letterdistributions",
                               DefaultLexicon: "NWL20",
                               DefaultLetterDistribution: "English" }
        
        gameHistory, err = gcgio.ParseGCGFromReader(cfg, strings.NewReader(gcg)) 
        if err != nil {
            fmt.Errorf("Error encountered parsing gcg from reader: %v", err)
        }
        
    } else {

        id = string(os.Args[1])
        gameHistory, err = GetGameHistory(id) 
        if err != nil {
            fmt.Println("Caught Error", err)
        }
    }
    gameGif, err := AnimateGame(tilesImg, boardConfig, gameHistory)

    f, err := os.OpenFile(id + ".gif", os.O_RDWR|os.O_CREATE, 0755)
    if err != nil {
        panic(err)
    }
    gif.EncodeAll(f, gameGif)
}
