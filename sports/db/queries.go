package db

const (
	eventsList = "list"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventsList: `
			SELECT
				id,
				name,
				advertised_start_time,
				sport,
				competition,
				home_team,
				away_team
			FROM events
		`,
	}
}
