package explorerlib

import (
	"fmt"
	"github.com/coveo/go-coveo/search"
	"github.com/coveo/uabot/scenariolib"
	"github.com/jmcvetta/randutil"
	"github.com/satori/go.uuid"
	"math"
	"time"
)

type Index struct {
	Client search.Client
}

var (
	t1                time.Time
	t2                time.Time
	t3                time.Time
	dt1               time.Duration
	dt2               time.Duration
	dt3               time.Duration
	numberOfActiveBot int = 0
	throttle          time.Duration
)

func NewIndex(endpoint string, searchToken string) (Index, error) {
	client, err := search.NewClient(search.Config{
		Endpoint:  endpoint,
		Token:     searchToken,
		UserAgent: "",
	})
	return Index{Client: client}, err
}

func (index *Index) FetchLanguages() ([]string, error) {
	languageFacetValues, err := index.Client.ListFacetValues("@syslanguage", math.MaxInt16)
	languages := []string{}
	for _, value := range languageFacetValues.Values {
		languages = append(languages, value.Value)
	}
	return languages, err
}

func (index *Index) FetchFieldValues(field string) (*search.FacetValues, error) {
	return index.Client.ListFacetValues(field, 1000)
}

func (index *Index) FindTotalCountFromQuery(query search.Query) (int, error) {
	response, status := index.Client.Query(query)
	return response.TotalCount, status
}

func (index *Index) FetchResponse(queryExpression string, numberOfResults int) (*search.Response, error) {
	return index.Client.Query(search.Query{
		AQ:              queryExpression,
		NumberOfResults: numberOfResults,
	})
}

func (index *Index) BuildGoodQueries(wordCountsByLanguage map[string]WordCounts, numberOfQueryByLanguage int, averageNumberOfWords int, minTime time.Duration, botId uuid.UUID) (map[string][]string, error) {

	numberOfActiveBot++
	throttle = (minTime * time.Millisecond) * time.Duration(numberOfActiveBot)
	scenariolib.Info.Printf("Throttled at : %v", throttle)

	queriesInLanguage := make(map[string][]string)
	scenariolib.Info.Println("Building queries and calling the index to validate that they return results")

	for language, wordCounts := range wordCountsByLanguage {
		words := []string{}

		choices := make([]randutil.Choice, 0, wordCounts.TotalCount)
		for _, wordCount := range wordCounts.Words {
			choices = append(choices, randutil.Choice{wordCount.Count, wordCount.Word})
		}

		t2 = time.Now()
		for i := 0; i < numberOfQueryByLanguage; {
			word := wordCounts.PickExpNWordsWeighted(choices, averageNumberOfWords)

			dt2 = time.Since(t2)
			if dt2 < throttle {
				time.Sleep(throttle - dt2)
			}
			t2 = time.Now()
			response, err := index.FetchResponse(word, 10)

			if err != nil {
				return nil, err
			}

			//todo fix this display fonction when multiple bot are working
			if len(response.Results) > 0 && !contains(words, word) {
				words = append(words, word)
				i++
				fmt.Printf("\rBot %v : Building and validating queries: %.0f %% completed for language %s", botId, (float32(i)/float32(numberOfQueryByLanguage))*100, language)
			}
		}
		fmt.Printf("\n")
		scenariolib.Info.Printf("Total number of good queries in %v: %v", language, len(words))
		queriesInLanguage[language] = words

	}
	fmt.Printf("\n")
	numberOfActiveBot--
	return queriesInLanguage, nil
}
