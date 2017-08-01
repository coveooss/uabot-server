package explorerlib

import (
	"github.com/coveo/go-coveo/search"
	"github.com/coveo/uabot/scenariolib"
	"time"
)

func FindWordsByLanguageInIndex(index Index, fields []string, documentsExplorationPercentage float64, fetchNumberOfResults int, minTime time.Duration) (map[string]WordCounts, error) {

	numberOfActiveBot++
	throttle = (minTime * time.Millisecond) * time.Duration(numberOfActiveBot)
	scenariolib.Info.Printf("Throttled at : %v", throttle)

	scenariolib.Info.Printf("Number of active bot : %v", numberOfActiveBot)
	wordCountsByLanguage := make(map[string]WordCounts)
	wordsByFieldValueByLanguage := map[string][]WordsByFieldValue{}
	languages, status := index.FetchLanguages()
	if status != nil {
		return nil, status
	}
	// for each language
	for _, language := range languages {
		// discover Words
		// for every fields provided
		for _, field := range fields {
			values, status := index.FetchFieldValues(field)
			if status != nil {
				return nil, status
			}
			t1 = time.Now()
			// for all values of the field
			for _, value := range values.Values {

				wordCounts := WordCounts{}

				dt1 = time.Since(t1)
				if dt1 < throttle {
					time.Sleep(throttle - dt1)
				}
				t1 = time.Now()

				totalCount, status := index.FindTotalCountFromQuery(search.Query{
					AQ: "@syslanguage=\"" + language + "\" " + field + "=\"" + value.Value + "\"",
				})
				if status != nil {
					return nil, status
				}

				var queryNumber int
				if tempQueryNumber := (int(float64(totalCount)*documentsExplorationPercentage) / fetchNumberOfResults); tempQueryNumber > 0 {
					queryNumber = tempQueryNumber
				} else {
					queryNumber = 1
				}
				randomWord := ""

				t3 = time.Now()
				for i := 0; i < queryNumber; i++ {

					// build A query from the word counts in the appropriate language with a filter on the field value
					queryExpression := randomWord +
						" @syslanguage=\"" + language + "\" " +
						field + "=\"" + value.Value + "\" "

					dt3 = time.Since(t3)
					if dt3 < throttle {
						time.Sleep(throttle - dt3)
					}
					t3 = time.Now()
					response, status := index.FetchResponse(queryExpression, fetchNumberOfResults)
					if status != nil {
						return nil, status
					}

					// extract words from the response
					newWordCounts := ExtractWordsFromResponse(*response)
					// update word counts
					wordCounts = wordCounts.Extend(newWordCounts)
					// pick a random word (Probability by popularity, or constant)
					randomWord = wordCounts.PickRandomWord()
				}
				taggedLanguage := LanguageToTag(language)
				wordsByFieldValueByLanguage[taggedLanguage] = append(wordsByFieldValueByLanguage[taggedLanguage], WordsByFieldValue{
					FieldName:  field,
					FieldValue: value.Value,
					Words:      wordCounts,
				})
			}
		}
	}
	// collapse results from all fields
	for language, wordCountsInLanguage := range wordsByFieldValueByLanguage {
		wordCounts := WordCounts{}
		for _, wordCountsByFields := range wordCountsInLanguage {
			wordCounts = wordCounts.Extend(wordCountsByFields.Words)
		}
		RankByWordCount(wordCounts)
		wordCountsByLanguage[language] = wordCounts
		scenariolib.Info.Print("language : ", language, " : Total words count ", len(wordCounts.Words))
	}
	numberOfActiveBot--
	return wordCountsByLanguage, nil
}
