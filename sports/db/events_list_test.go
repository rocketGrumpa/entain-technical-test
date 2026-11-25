package db

import (
	"database/sql"
	"testing"
	"time"
)

// newListTestRepo returns an events repository backed by an in-memory SQLite
// database with two events. Their start times use different UTC offsets.
func newListTestRepo(t *testing.T) *eventsRepo {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	// Each connection to :memory: gets its own empty database, so keep a
	// single connection to make sure every query sees the same table.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	// Same schema as seed() in db.go. If the schema in db.go changes,
	// update this too.
	if _, err := db.Exec(`CREATE TABLE events (id INTEGER PRIMARY KEY, name TEXT, advertised_start_time DATETIME, sport TEXT, competition TEXT, home_team TEXT, away_team TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	fixtures := []struct {
		id                int64
		start, home, away string
	}{
		{1, "2026-11-23T12:00:00Z", "Atlanta Braves", "New York Mets"},
		{2, "2026-11-24T09:30:00+10:00", "Texas Rangers", "Seattle Mariners"},
	}
	for _, f := range fixtures {
		if _, err := db.Exec(
			`INSERT INTO events(id, name, advertised_start_time, sport, competition, home_team, away_team) VALUES (?,?,?,?,?,?,?)`,
			f.id, f.away+" vs "+f.home, f.start, "Baseball", "MLB", f.home, f.away,
		); err != nil {
			t.Fatalf("insert event %d: %v", f.id, err)
		}
	}

	return &eventsRepo{db: db}
}

// TestList runs List against an in-memory SQLite database and checks that
// every column comes back in the right field.
func TestList(t *testing.T) {
	repo := newListTestRepo(t)

	events, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	want := []struct {
		id               int64
		name, home, away string
		start            time.Time
	}{
		{1, "New York Mets vs Atlanta Braves", "Atlanta Braves", "New York Mets",
			time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)},
		{2, "Seattle Mariners vs Texas Rangers", "Texas Rangers", "Seattle Mariners",
			time.Date(2026, 11, 23, 23, 30, 0, 0, time.UTC)},
	}
	if len(events) != len(want) {
		t.Fatalf("got %d events, want %d", len(events), len(want))
	}
	for i, w := range want {
		got := events[i]
		if got.Id != w.id || got.Name != w.name || got.Sport != "Baseball" || got.Competition != "MLB" ||
			got.HomeTeam != w.home || got.AwayTeam != w.away {
			t.Errorf("event %d = {%d, %q, %q, %q, home %q, away %q}, want {%d, %q, %q, %q, home %q, away %q}",
				i, got.Id, got.Name, got.Sport, got.Competition, got.HomeTeam, got.AwayTeam,
				w.id, w.name, "Baseball", "MLB", w.home, w.away)
		}
		if start := got.GetAdvertisedStartTime().AsTime(); !start.Equal(w.start) {
			t.Errorf("event %d start = %v, want %v", w.id, start, w.start)
		}
	}
}
