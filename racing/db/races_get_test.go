package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// newGetTestRepo returns a races repository backed by an in-memory SQLite
// database with two races: id 1 is visible, id 2 is not.
func newGetTestRepo(t *testing.T) *racesRepo {
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
	if _, err := db.Exec(`CREATE TABLE races (id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, number INTEGER, visible INTEGER, advertised_start_time DATETIME)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	start := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	fixtures := []struct {
		id, meetingID, visible int64
	}{
		{1, 1, 1},
		{2, 3, 0},
	}
	for _, f := range fixtures {
		if _, err := db.Exec(
			`INSERT INTO races(id, meeting_id, name, number, visible, advertised_start_time) VALUES (?,?,?,?,?,?)`,
			f.id, f.meetingID, "Test race", 1, f.visible, start,
		); err != nil {
			t.Fatalf("insert race %d: %v", f.id, err)
		}
	}

	return &racesRepo{db: db}
}

// TestGet runs Get against an in-memory SQLite database.
func TestGet(t *testing.T) {
	repo := newGetTestRepo(t)

	t.Run("returns the race even when it is not visible", func(t *testing.T) {
		race, err := repo.Get(2)
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		if race.Id != 2 || race.MeetingId != 3 || race.Visible {
			t.Errorf("Get(2) = id %d, meeting %d, visible %v; want id 2, meeting 3, visible false",
				race.Id, race.MeetingId, race.Visible)
		}
		wantStart := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
		if got := race.GetAdvertisedStartTime().AsTime(); !got.Equal(wantStart) {
			t.Errorf("advertised start time = %v, want %v", got, wantStart)
		}
	})

	t.Run("missing id returns ErrNotFound", func(t *testing.T) {
		if _, err := repo.Get(99); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get(99) error = %v, want ErrNotFound", err)
		}
	})
}
