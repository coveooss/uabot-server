package autobot

import (
	"github.com/coveo/uabot-server/explorerlib"
	"github.com/coveo/uabot/scenariolib"
	"math/rand"
)

type Autobot struct {
	config *explorerlib.Config
	random *rand.Rand
}

func NewAutobot(_config *explorerlib.Config, _random *rand.Rand) *Autobot {
	return &Autobot{
		config: _config,
		random: _random,
	}
}

func (bot *Autobot) Run(quitChannel chan bool) error {
	index, status := explorerlib.NewIndex(bot.config.SearchEndpoint, bot.config.SearchToken)

	wordCountsByLanguage, status := explorerlib.FindWordsByLanguageInIndex(
		index,
		bot.config.FieldsToExploreEqually,
		bot.config.DocumentsExplorationPercentage,
		bot.config.FetchNumberOfResults)
	if status != nil {
		return status
	}
	scenariolib.Info.Println("words count by language")

	languages, status := index.Client.ListFacetValues("@language", 1000)
	if status != nil {
		return status
	}
	scenariolib.Info.Println("language")

	goodQueries, status := index.BuildGoodQueries(wordCountsByLanguage, bot.config.NumberOfQueryByLanguage, bot.config.AverageNumberOfWordsPerQuery)
	if status != nil {
		return status
	}
	scenariolib.Info.Println("good queries")

	taggedLanguages := make([]string, 0)
	scenarios := []*scenariolib.Scenario{}

	originLevels := bot.config.OriginLevels

	for originLevel1, originLevels2 := range originLevels {
		for _, originLevel2 := range originLevels2 {
			for _, lang := range languages.Values {
				scenariolib.Info.Println("begin building scenarios")
				taggedLanguage := explorerlib.LanguageToTag(lang.Value)
				taggedLanguages = append(taggedLanguages, taggedLanguage)
				scenario := explorerlib.NewScenarioBuilder().WithName("1 search and click in " + lang.Value).WithWeight(lang.NumberOfResults).WithLanguage(taggedLanguage).WithEvent(explorerlib.NewSetOriginLevels(originLevel1, originLevel2)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).Build()
				scenarios = append(scenarios, scenario)

				scenario = explorerlib.NewScenarioBuilder().WithName("2 search and click in " + lang.Value).WithWeight(lang.NumberOfResults).WithLanguage(taggedLanguage).WithEvent(explorerlib.NewSetOriginLevels(originLevel1, originLevel2)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).Build()
				scenarios = append(scenarios, scenario)

				scenario = explorerlib.NewScenarioBuilder().WithName("3 search and click in " + lang.Value).WithWeight(lang.NumberOfResults).WithLanguage(taggedLanguage).WithEvent(explorerlib.NewSetOriginLevels(originLevel1, originLevel2)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).Build()
				scenarios = append(scenarios, scenario)

				scenario = explorerlib.NewScenarioBuilder().WithName("4 search and click in " + lang.Value).WithWeight(lang.NumberOfResults).WithLanguage(taggedLanguage).WithEvent(explorerlib.NewSetOriginLevels(originLevel1, originLevel2)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).Build()
				scenarios = append(scenarios, scenario)

				scenario = explorerlib.NewScenarioBuilder().WithName("5 search and click in " + lang.Value).WithWeight(lang.NumberOfResults).WithLanguage(taggedLanguage).WithEvent(explorerlib.NewSetOriginLevels(originLevel1, originLevel2)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).WithEvent(explorerlib.NewSearchEvent(true)).WithEvent(explorerlib.NewClickEvent(0.5)).WithEvent(explorerlib.NewClickEvent(0.8)).Build()
				scenarios = append(scenarios, scenario)

				viewScenarioBuilder := explorerlib.NewScenarioBuilder().WithName("views in " + lang.Value).WithWeight(lang.NumberOfResults).WithLanguage(taggedLanguage).WithEvent(explorerlib.NewSetOriginLevels(originLevel1, originLevel2)).WithEvent(explorerlib.NewSearchEvent(false))
				for i := 0; i < 20; i++ {
					viewScenarioBuilder.WithEvent(explorerlib.NewViewEvent())
				}
				scenarios = append(scenarios, viewScenarioBuilder.Build())
				scenariolib.Info.Println("finished building scenarios")
			}
		}
	}

	err := explorerlib.NewBotConfigurationBuilder().WithOrgName(bot.config.Org).WithSearchEndpoint(bot.config.SearchEndpoint).WithAnalyticsEndpoint(bot.config.AnalyticsEndpoint).AllAnonymous().WithLanguages(taggedLanguages).WithGoodQueryByLanguage(goodQueries).WithTimeBetweenActions(1).WithTimeBetweenVisits(5).WithScenarios(scenarios).NoWait().Save(bot.config.OutputFilePath)
	if err != nil {
		return err
	}
	scenariolib.Info.Println("bot configuration completed")

	uabot := scenariolib.NewUabot(true, bot.config.OutputFilePath, bot.config.SearchToken, bot.config.AnalyticsToken, bot.random)

	scenariolib.Info.Println("Running Bot")
	err = uabot.Run(quitChannel)
	return err
}

func (bot *Autobot) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"searchEndpoint":                 bot.config.SearchEndpoint,
		"analyticsEndpoint":              bot.config.AnalyticsEndpoint,
		"averageNumberOfWordsPerQuery":   bot.config.AverageNumberOfWordsPerQuery,
		"documentsExplorationPercentage": bot.config.DocumentsExplorationPercentage,
		"fieldsToExploreEqually":         bot.config.FieldsToExploreEqually,
		"org":                      bot.config.Org,
		"outputFilepath":           bot.config.OutputFilePath,
		"numberOfQueryPerLanguage": bot.config.NumberOfQueryByLanguage,
		"numberOfResultsPerQuery":  bot.config.FetchNumberOfResults,
		"originLevels":             bot.config.OriginLevels,
	}
}
