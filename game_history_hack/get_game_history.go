package game_history_hack

import (
    "fmt"
    "bytes"
    "io/ioutil"
    "net/http"
    
    pb "github.com/domino14/liwords/rpc/api/proto/game_service"
    macondopb "github.com/domino14/macondo/gen/api/proto/macondo"

    "google.golang.org/protobuf/proto"   
)

func GetGameHistory (id string) (*macondopb.GameHistory, error) {

    gameHistoryRequest := &pb.GameHistoryRequest{ GameId: id }
    req, _ := proto.Marshal(gameHistoryRequest)

    response, err := http.Post("https://woogles.io/twirp/game_service.GameMetadataService/GetGameHistory", 
                               "application/protobuf", bytes.NewBuffer(req))

    if err != nil {
        return &macondopb.GameHistory{}, fmt.Errorf("HTTP Post Error: %s\n", err)
    } else {
        gameHistoryResponse := &pb.GameHistoryResponse{}
        bytes, _ := ioutil.ReadAll(response.Body)
        if err := proto.Unmarshal(bytes, gameHistoryResponse); err != nil {
            return &macondopb.GameHistory{}, fmt.Errorf("Unmarshaling Error: %s\n%v\n", err, response.Body)
        } else {
            fmt.Println("Game History request successful!\n")
        }
        return gameHistoryResponse.History, nil
    }
}
