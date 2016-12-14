package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coveo/uabot-server/explorerlib"
	"github.com/coveo/uabot/scenariolib"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"math/rand"
	"net/http"
	"time"
)

const (
	MINIMUMTIMETOLIVE int = 1
	MAXIMUMTIMETOLIVE int = 120
	DEFAULTIMETOLIVE  int = 2

	MINIMUMNUMBERWORDSPERQUERY int = 1
	MAXIMUMNUMBERWORDSPERQUERY int = 20
	DEFAULTNUMBERWORDSPERQUERY int = 2

	MINIMUMDOCUMENTEXPLORATIONPERCENT float64 = 0.001
	MAXIMUMDOCUMENTEXPLORATIONPERCENT float64 = 1
	DEFAULTDOCUMENTEXPLORATIONPERCENT float64 = 0.01

	MINIMUMNUMBEROFQUERYPERLANGUAGE int = 0
	MAXIMUMNUMBEROFQUERYPERLANGUAGE int = 200
	DEFAULTNUMBEROFQUERYPERLANGUAGE int = 10

	MINIMUMFETCHNUMBEROFRESULTS int = 0
	MAXIMUMFETCHNUMBEROFRESULTS int = 2000
	DEFAULTFETCHNUMBEROFRESULTS int = 1000
)

var (
	quitChannels map[uuid.UUID]chan bool
	random       *rand.Rand
	workPool     *WorkPool
)

func Init(_workPool *WorkPool, _random *rand.Rand) {
	workPool = _workPool
	quitChannels = make(map[uuid.UUID]chan bool)
	random = _random
}

func Start(writter http.ResponseWriter, request *http.Request) {
	config, err := DecodeConfig(request.Body)
	if err != nil {
		http.Error(writter, err.Error(), http.StatusTeapot)
		return
	}

	config.Id = uuid.NewV4()

	err = validateConfig(config)
	if err != nil {
		scenariolib.Error.Print(err.Error())
		http.Error(writter, err.Error(), http.StatusBadRequest)
		return
	}
	//Format the Config into a JSON for display purpose
	out, err := json.MarshalIndent(config,"","	")
	if err != nil {
		http.Error(writter, err.Error(), http.StatusTeapot)
		return
	}
	scenariolib.Info.Println("Current Configuration : \n" + string(out))

	timer := time.NewTimer(time.Duration(config.TimeToLive) * time.Minute)
	quitChannel := make(chan bool)
	go func() {
		<-timer.C
		scenariolib.Info.Printf("Timer Timed Out")
		close(quitChannel)
	}()
	quitChannels[config.Id] = quitChannel
	worker := NewWorker(config, quitChannel, random, config.Id)
	err = workPool.PostWork(&worker)
	if err != nil {
		scenariolib.Error.Printf("Error : %v\n", err)
	}
	json.NewEncoder(writter).Encode(map[string]interface{}{
		"workerID": config.Id,
	})
}

func validateConfig(config *explorerlib.Config) error {
	if config.OriginLevels == nil {
		return errors.New("Origin Level 1 Missing")
	} else {
		for originLevel1, originLevel2 := range config.OriginLevels {
			if len(originLevel2) == 0 {
				return errors.New("Origin Level 2 Missing for originLevel1: " + originLevel1)
			}
		}
	}
	if config.SearchEndpoint == "" {
		return errors.New("searchEndpoint Missing")
	}
	if config.SearchToken == "" {
		return errors.New("searchToken Missing")
	}
	if config.AnalyticsEndpoint == "" {
		return errors.New("analyticsEndpoint Missing")
	}
	if config.AnalyticsToken == "" {
		return errors.New("analyticsToken Missing")
	}
	if config.TimeToLive < MINIMUMTIMETOLIVE || config.TimeToLive > MAXIMUMTIMETOLIVE {
		scenariolib.Warning.Print("TimeToLive is out of bounds, should be in [%v,%v], will use default value of %v ",MINIMUMTIMETOLIVE,MAXIMUMTIMETOLIVE,DEFAULTIMETOLIVE)
		config.TimeToLive = DEFAULTIMETOLIVE
	}
	if config.AverageNumberOfWordsPerQuery < MINIMUMNUMBERWORDSPERQUERY || config.AverageNumberOfWordsPerQuery > MAXIMUMNUMBERWORDSPERQUERY {
		scenariolib.Warning.Print("AverageNumberOfWordsPerQuery is out of bounds, should be in [%v,%v], will use default value of %v ",MINIMUMNUMBERWORDSPERQUERY,MAXIMUMNUMBERWORDSPERQUERY,DEFAULTNUMBERWORDSPERQUERY)
		config.AverageNumberOfWordsPerQuery = DEFAULTNUMBERWORDSPERQUERY
	}
	if config.DocumentsExplorationPercentage < MINIMUMDOCUMENTEXPLORATIONPERCENT || config.DocumentsExplorationPercentage > MAXIMUMDOCUMENTEXPLORATIONPERCENT {
		scenariolib.Warning.Print("DocumentsExplorationPercentage is out of bounds, should be in [0%,100%], will use default value of %f %%",DEFAULTDOCUMENTEXPLORATIONPERCENT * 100)
		config.DocumentsExplorationPercentage = DEFAULTDOCUMENTEXPLORATIONPERCENT
	}
	if config.NumberOfQueryByLanguage < MINIMUMNUMBEROFQUERYPERLANGUAGE || config.NumberOfQueryByLanguage > MAXIMUMNUMBEROFQUERYPERLANGUAGE {
		scenariolib.Warning.Print("NumberOfQueryByLanguage is out of bounds, should be in [%v,%v], will use default value of %v ",MINIMUMNUMBEROFQUERYPERLANGUAGE,MAXIMUMNUMBEROFQUERYPERLANGUAGE,DEFAULTNUMBEROFQUERYPERLANGUAGE)
		config.NumberOfQueryByLanguage = DEFAULTNUMBEROFQUERYPERLANGUAGE
	}
	if config.FetchNumberOfResults < MINIMUMFETCHNUMBEROFRESULTS || config.FetchNumberOfResults > MAXIMUMFETCHNUMBEROFRESULTS {
		scenariolib.Warning.Print("FetchNumberOfResults is out of bounds, should be in [%v,%v], will use default value of %v ",MINIMUMFETCHNUMBEROFRESULTS,MAXIMUMFETCHNUMBEROFRESULTS,DEFAULTFETCHNUMBEROFRESULTS)
		config.FetchNumberOfResults = DEFAULTFETCHNUMBEROFRESULTS
	}
	if config.FieldsToExploreEqually == nil || len(config.FieldsToExploreEqually) == 0 {
		scenariolib.Warning.Print("FieldsToExploreEqually out of bounds, will be set to default value of @syssource")
		config.FieldsToExploreEqually = []string{"@syssource"}
	}
	if config.OutputFilePath == "" {
		scenariolib.Warning.Print("OutputFilePath undefined, will be set to ", config.Id.String()+".json")
		config.OutputFilePath = config.Id.String() + ".json"
	}
	if config.Org == "" {
		return errors.New("Org Missing")
	}
	return nil
}

func Stop(writter http.ResponseWriter, request *http.Request) {
	Vars := mux.Vars(request)
	id, _ := uuid.FromString(Vars["id"])
	close(quitChannels[id])
	delete(quitChannels, id)
}

func GetInfo(writter http.ResponseWriter, request *http.Request) {
	infos := map[string]interface{}{
		"status":         "UP",
		"botWorkerInfos": workPool.getInfo(),
		"activeRoutines": fmt.Sprintf("%v/%v", workPool.ActiveRoutines(), workPool.NumberConcurrentRoutine),
		"queuedWork":     fmt.Sprintf("%v/%v", workPool.QueuedWork(), workPool.QueueLength),
	}
	writter.Header().Add("Content-Type", "application/json")
	json.NewEncoder(writter).Encode(infos)
}
