package db

import (
	"database/sql"
	"testing"
	"time"
)

// TestSeed runs Init on an empty in-memory database and checks the dummy
// events it creates.
func TestSeed(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	// Each connection to :memory: gets its own empty database, so keep a
	// single connection to make sure every query sees the same table.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	repo := &eventsRepo{db: db}

	// Start times are stored to the second, so allow one second of rounding.
	earliest := time.Now().AddDate(0, 0, -1).Add(-time.Second)
	if err := repo.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	latest := time.Now().AddDate(0, 0, 2)

	events, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(events) != 100 {
		t.Fatalf("got %d events, want 100", len(events))
	}

	teams := make(map[string]bool, len(mlbTeams))
	for _, team := range mlbTeams {
		teams[team] = true
	}

	for _, e := range events {
		if e.Sport != "Baseball" || e.Competition != "MLB" {
			t.Errorf("event %d is %q %q, want %q %q", e.Id, e.Sport, e.Competition, "Baseball", "MLB")
		}

		if !teams[e.HomeTeam] || !teams[e.AwayTeam] {
			t.Errorf("event %d teams %q and %q are not both MLB teams", e.Id, e.HomeTeam, e.AwayTeam)
		} else if e.HomeTeam == e.AwayTeam {
			t.Errorf("event %d has %q playing itself", e.Id, e.HomeTeam)
		}

		if want := e.AwayTeam + " vs " + e.HomeTeam; e.Name != want {
			t.Errorf("event %d name = %q, want %q", e.Id, e.Name, want)
		}

		start := e.GetAdvertisedStartTime().AsTime()
		if start.Before(earliest) || start.After(latest) {
			t.Errorf("event %d start %v is outside %v to %v", e.Id, start, earliest, latest)
		}
	}
}
