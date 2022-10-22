package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
)

type Team struct {
	Name              string             `json:"teamName"`
	TeamLink          string             `json:"teamLink"`
	MatchsInformation []MatchInformation `json:"matchsInformation"`
}

type TeamResultInformation struct {
	TeamLocationState string `json:"teamLocationState"`
	TeamName          string `json:"teamName"`
	TotalGoals        string `json:"totalGoals"`
	PosessionPercent  string `json:"posessionPercent"`
	Corners           string `json:"corners"`
	YellowCards       string `json:"yellowCards"`
}

type MatchInformation struct {
	MatchLink   string                `json:"matchLink"`
	MatchResult string                `json:"matchResult"`
	IsWinner    bool                  `json:"isWinner"`
	LocalTeam   TeamResultInformation `json:"localTeam"`
	VisitorTeam TeamResultInformation `json:"visitorTeam"`
}

func handler(w http.ResponseWriter, r *http.Request) {

	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("visiting", URL)

	var teamInformation Team
	var matchs = make([]MatchInformation, 0)
	var matchInformation MatchInformation

	collector := colly.NewCollector(
		colly.AllowedDomains("es.besoccer.com", "www.es.besoccer.com"),
	)

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
		matchs = append(matchs, matchInformation)
	})

	//collector.Visit("https://es.besoccer.com/equipo/real-madrid")
	collector.Visit(URL)

	teamInformation.MatchsInformation = matchs
	data := writeJSON(teamInformation)
	fmt.Println("Done")
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func main() {
	// example usage: curl -s 'http://127.0.0.1:7171/?url=https://es.besoccer.com/equipo/real-madrid'
	addr := ":7171"

	http.HandleFunc("/", handler)

	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func writeJSON(data Team) []byte {
	byteData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println("Unable to create json file")
		return nil
	}

	return byteData
}
