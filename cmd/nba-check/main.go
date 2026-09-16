// Command nba-check makes one small live request to nba-api-go.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/n-ae/nba-api-go/v3/pkg/live"
	liveendpoints "github.com/n-ae/nba-api-go/v3/pkg/live/endpoints"
	"github.com/n-ae/nba-api-go/v3/pkg/stats"
	statsendpoints "github.com/n-ae/nba-api-go/v3/pkg/stats/endpoints"
	"github.com/n-ae/nba-api-go/v3/pkg/stats/parameters"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "box-score" {
		probeGameBoxScore()
		return
	}

	probeLiveScoreboard()
	probeTeamRoster()
	probePlayerGameLog()
	probeGameBoxScore()
}

func probeLiveScoreboard() {
	ctx, cancel := requestContext()
	defer cancel()

	client := live.NewDefaultClient()
	resp, err := liveendpoints.Scoreboard(ctx, client)
	if err != nil {
		fmt.Printf("scoreboard: unavailable: %v\n", err)
		return
	}

	fmt.Printf("NBA scoreboard for %s: %d game(s)\n", resp.Data.Scoreboard.GameDate, len(resp.Data.Scoreboard.Games))
	for _, game := range resp.Data.Scoreboard.Games {
		fmt.Printf("%s @ %s — %s\n", game.AwayTeam.TeamTricode, game.HomeTeam.TeamTricode, game.GameStatusText)
	}
}

func probeTeamRoster() {
	ctx, cancel := requestContext()
	defer cancel()

	client := stats.NewDefaultClient()
	season := parameters.NewSeason(2024)
	leagueID := parameters.LeagueIDNBA
	resp, err := statsendpoints.GetCommonTeamRoster(ctx, client, statsendpoints.CommonTeamRosterRequest{
		TeamID:   "1610612747", // Los Angeles Lakers
		Season:   &season,
		LeagueID: &leagueID,
	})
	if err != nil {
		fmt.Printf("team roster: unavailable: %v\n", err)
		return
	}

	fmt.Printf("team roster: available: %d player(s)\n", len(resp.Data.CommonTeamRoster))
}

func probePlayerGameLog() {
	ctx, cancel := requestContext()
	defer cancel()

	client := stats.NewDefaultClient()
	resp, err := statsendpoints.PlayerGameLog(ctx, client, statsendpoints.PlayerGameLogRequest{
		PlayerID:   "2544", // LeBron James
		Season:     parameters.NewSeason(2024),
		SeasonType: parameters.SeasonTypeRegular,
		LeagueID:   parameters.LeagueIDNBA,
	})
	if err != nil {
		fmt.Printf("player game log: unavailable: %v\n", err)
		return
	}

	fmt.Printf("player game log: available: %d game(s)\n", len(resp.Data.PlayerGameLog))
}

func probeGameBoxScore() {
	ctx, cancel := requestContext()
	defer cancel()

	client := stats.NewDefaultClient()
	resp, err := statsendpoints.GetBoxScoreTraditionalV2(ctx, client, statsendpoints.BoxScoreTraditionalV2Request{
		GameID: "0022401230", // A completed 2024-25 regular-season game.
	})
	if err != nil {
		fmt.Printf("game box score: unavailable: %v\n", err)
		return
	}

	fmt.Printf("game box score: available: %d player stat lines\n", len(resp.Data.PlayerStats))
}

func requestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
