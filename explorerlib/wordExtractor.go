package explorerlib

import (
	"github.com/coveo/go-coveo/search"
	"regexp"
	"strings"
)

func ExtractWordsFromResponse(response search.Response) WordCounts {
	titleWords := ExtractWordCountsFromTitlesInResponse(response)
	concepts := ExtractWordCountsFromConceptsInResponse(response)

	return titleWords.Extend(concepts)
}

func ExtractWordCountsFromTitlesInResponse(response search.Response) WordCounts {
	var text string
	for _, result := range response.Results {
		text = strings.Join([]string{text, result.Title}, " ")
	}
	text = CleanText(text)
	reg, _ := regexp.Compile("([\\w\\p{L}\\p{Nd}']+)")

	words := []string{}
	for _, word := range reg.FindAllString(text, -1) {
		if len(word) > 2 {
			words = append(words, word)
		}
	}

	results := CountWordOccurence(words)
	return results
}

func ExtractWordCountsFromConceptsInResponse(response search.Response) WordCounts {
	conceptsList := WordCounts{}
	for _, groupBy := range response.GroupByResults {
		for _, concepts := range groupBy.Values {
			cleanValue := CleanText(concepts.Value)
			if len(cleanValue) > 2 {
				conceptsList = conceptsList.Add(WordCount{cleanValue, concepts.NumberOfResults})
			}
		}
	}
	return conceptsList
}
