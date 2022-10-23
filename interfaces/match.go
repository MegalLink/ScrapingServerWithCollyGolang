package interfaces

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
	LocalTeam   TeamResultInformation `json:"localTeam"`
	VisitorTeam TeamResultInformation `json:"visitorTeam"`
}
