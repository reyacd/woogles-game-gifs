package game_history_hack

import (
    "fmt"
    "bytes"
    "io/ioutil"
    "net/http"
    "strings"
    
    "github.com/domino14/macondo/gcgio"
    "github.com/domino14/macondo/config"
    pb "github.com/domino14/liwords/rpc/api/proto/game_service"
    macondopb "github.com/domino14/macondo/gen/api/proto/macondo"

    "google.golang.org/protobuf/proto"   
)

func GetGameHistory (id string) (*macondopb.GameHistory, error) {

    gcgRequest := &pb.GCGRequest{ GameId: id }
    req, _ := proto.Marshal(gcgRequest)

    response, err := http.Post("https://woogles.io/twirp/game_service.GameMetadataService/GetGCG", 
                               "application/protobuf", bytes.NewBuffer(req))

    if err != nil {
        return &macondopb.GameHistory{}, fmt.Errorf("HTTP Post Error: %s\n", err)
    } else {
        gcgResponse := &pb.GCGResponse{}
        bytes, _ := ioutil.ReadAll(response.Body)
        if err := proto.Unmarshal(bytes, gcgResponse); err != nil {
            return &macondopb.GameHistory{}, fmt.Errorf("Unmarshaling Error: %s\n%v\n", err, response.Body)
        } else {
            fmt.Println("GCG request successful!\n")
        }

        cfg := &config.Config{ 
            Debug: false,
	    LetterDistributionPath: "../../macondo/data/letterdistributions",
	    StrategyParamsPath: "../../macondo/data/strategy",
	    LexiconPath: "../../macondo/data/lexica",
	    DefaultLexicon: "NWL20",
	    DefaultLetterDistribution: "English",
        } 
    
        reader := strings.NewReader(gcgResponse.Gcg)
        if err != nil {
            return &macondopb.GameHistory{}, fmt.Errorf("Reader Error: %s\n", err)
        }
    
        gameHistory, err := gcgio.ParseGCGFromReader(cfg, reader)
        if err != nil {
            return &macondopb.GameHistory{}, fmt.Errorf("GCG Parse Error: %s\n", err)
        } else {
            fmt.Println("GCG to GameHistory conversion successful!")
            return gameHistory, nil
        }
    }
}
