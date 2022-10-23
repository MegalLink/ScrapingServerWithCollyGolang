package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/MegalLink/my/interfaces"
	"github.com/gocolly/colly"
)

type ScraperService interface {
	ScrapTeamInformation(URL string) interfaces.Team
}

type Scraper struct {
	Collector *colly.Collector
}

func NewScraperService(collector *colly.Collector) ScraperService {
	return &Scraper{
		Collector: collector,
	}
}

func (s *Scraper) ScrapTeamInformation(URL string) interfaces.Team {
	log.Println("visiting", URL)

	var teamInformation interfaces.Team
	var matchs = make([]interfaces.MatchInformation, 0)
	var matchInformation interfaces.MatchInformation

	collector := s.Collector

	matchInformationCollector := collector.Clone()

	collector.OnRequest(func(request *colly.Request) {
		requestUrl := request.URL.String()

		fmt.Println("Visiting", requestUrl)
		teamInformation.TeamLink = requestUrl
		requestUrlParts := strings.Split(requestUrl, "/")
		teamInformation.Name = requestUrlParts[len(requestUrlParts)-1]
	})

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	collector.OnHTML(".match-cell a", func(element *colly.HTMLElement) {
		matchLink := element.Attr("href")
		fmt.Println("MatchLink", matchLink)
		matchInformationCollector.Visit(matchLink)
	})

	matchInformationCollector.OnHTML("p[itemprop=homeTeam]", func(element *colly.HTMLElement) {
		matchInformation.LocalTeam.TeamLocationState = "local"
		matchInformation.LocalTeam.TeamName = strings.TrimSpace(element.Text)
	})

	matchInformationCollector.OnHTML("p[itemprop=awayTeam]", func(element *colly.HTMLElement) {
		matchInformation.VisitorTeam.TeamLocationState = "visitor"
		matchInformation.VisitorTeam.TeamName = strings.TrimSpace(element.Text)
	})

	matchInformationCollector.OnHTML(".marker .data", func(element *colly.HTMLElement) {
		matchResult := strings.TrimSpace(element.Text)
		matchInformation.MatchResult = matchResult
		goalsByTeam := strings.Split(matchResult, "-")
		matchInformation.LocalTeam.TotalGoals = strings.TrimSpace(goalsByTeam[0])
		matchInformation.VisitorTeam.TotalGoals = strings.TrimSpace(goalsByTeam[1])
	})

	matchInformationCollector.OnHTML(".compare-data table tbody", func(tableBody *colly.HTMLElement) {
		tableRowsSize := tableBody.DOM.Children().Size()
		//posession percent
		localPosessionPercent := tableBody.DOM.Find(".possession-graph .local p").Text()
		visitorPosessionPercent := tableBody.DOM.Find(".possession-graph .visitor p").Text()
		matchInformation.LocalTeam.PosessionPercent = localPosessionPercent
		matchInformation.VisitorTeam.PosessionPercent = visitorPosessionPercent
		//corners
		cornersContainer := tableBody.DOM.Children().Eq(3)
		localCorners := cornersContainer.Find(".td-num").First().Text()
		visitorCorners := cornersContainer.Find(".td-num").Last().Text()
		matchInformation.LocalTeam.Corners = localCorners
		matchInformation.VisitorTeam.Corners = visitorCorners
		//yellowCards
		yellowCardsContainer := tableBody.DOM.Children().Eq(tableRowsSize - 2)
		localYellowCards := yellowCardsContainer.Children().Eq(0).Text()
		visitorYellowCards := yellowCardsContainer.Children().Eq(2).Text()
		matchInformation.LocalTeam.YellowCards = localYellowCards
		matchInformation.VisitorTeam.YellowCards = visitorYellowCards
	})

	matchInformationCollector.OnScraped(func(r *colly.Response) {
		matchInformation.MatchLink = r.Request.URL.String()
		fmt.Println("Match Information", matchInformation)
		matchs = append(matchs, matchInformation)
	})

	collector.OnScraped(func(r *colly.Response) {
		teamInformation.MatchsInformation = matchs
	})

	collector.Visit(URL)
	collector.Wait()
	return teamInformation
}
