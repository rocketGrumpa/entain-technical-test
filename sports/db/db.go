package db

import (
	"fmt"
	"math/rand"
	"time"

	"syreclabs.com/go/faker"
)

// mlbTeams are the 30 Major League Baseball teams, used to make up event names.
var mlbTeams = []string{
	"Arizona Diamondbacks", "Atlanta Braves", "Baltimore Orioles", "Boston Red Sox",
	"Chicago Cubs", "Chicago White Sox", "Cincinnati Reds", "Cleveland Guardians",
	"Colorado Rockies", "Detroit Tigers", "Houston Astros", "Kansas City Royals",
	"Los Angeles Angels", "Los Angeles Dodgers", "Miami Marlins", "Milwaukee Brewers",
	"Minnesota Twins", "New York Mets", "New York Yankees", "Athletics",
	"Philadelphia Phillies", "Pittsburgh Pirates", "San Diego Padres", "San Francisco Giants",
	"Seattle Mariners", "St. Louis Cardinals", "Tampa Bay Rays", "Texas Rangers",
	"Toronto Blue Jays", "Washington Nationals",
}

// seed creates the events table and fills it with 100 dummy MLB games.
// It stops at the first error, so a failed insert isn't hidden by the ones after it.
func (r *eventsRepo) seed() error {
	if _, err := r.db.Exec(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, name TEXT, advertised_start_time DATETIME, sport TEXT, competition TEXT, home_team TEXT, away_team TEXT)`); err != nil {
		return err
	}

	now := time.Now()
	for i := 1; i <= 100; i++ {
		h := rand.Intn(len(mlbTeams))
		// Pick any team except the home team, so a team never plays itself.
		a := (h + 1 + rand.Intn(len(mlbTeams)-1)) % len(mlbTeams)
		home, away := mlbTeams[h], mlbTeams[a]

		if _, err := r.db.Exec(
			`INSERT OR IGNORE INTO events(id, name, advertised_start_time, sport, competition, home_team, away_team) VALUES (?,?,?,?,?,?,?)`,
			i,
			fmt.Sprintf("%s vs %s", away, home),
			faker.Time().Between(now.AddDate(0, 0, -1), now.AddDate(0, 0, 2)).Format(time.RFC3339),
			"Baseball",
			"MLB",
			home,
			away,
		); err != nil {
			return err
		}
	}

	return nil
}
