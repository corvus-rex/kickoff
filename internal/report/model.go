package report

type MatchReport struct {
	MatchID           uint     `json:"match_id"`
	MatchDate         string   `json:"match_date"`
	MatchTime         string   `json:"match_time"`
	HomeTeam          TeamInfo `json:"home_team"`
	AwayTeam          TeamInfo `json:"away_team"`
	HomeScore         int      `json:"home_score"`
	AwayScore         int      `json:"away_score"`
	Result            string   `json:"result"`
	TopScorer         *Scorer  `json:"top_scorer,omitempty"`
	HomeCumulativeWins int     `json:"home_cumulative_wins"`
	AwayCumulativeWins int     `json:"away_cumulative_wins"`
}

type TeamInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Scorer struct {
	PlayerID   uint   `json:"player_id"`
	PlayerName string `json:"player_name"`
	Goals      int    `json:"goals"`
}
