package db

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
)

// TestApplyOrder checks the ORDER BY clause applyOrder builds.
func TestApplyOrder(t *testing.T) {
	repo := &racesRepo{}
	base := "SELECT id FROM races"

	tests := []struct {
		name      string
		orderBy   []OrderBy
		wantOrder string
		wantErr   bool
	}{
		{"no keys still orders by id", nil, " ORDER BY id ASC", false},
		{"start time is compared with julianday", []OrderBy{{Field: "advertised_start_time"}},
			" ORDER BY julianday(advertised_start_time) ASC, id ASC", false},
		{"several keys, then id", []OrderBy{{Field: "meeting_id"}, {Field: "advertised_start_time", Desc: true}},
			" ORDER BY meeting_id ASC, julianday(advertised_start_time) DESC, id ASC", false},
		{"id is not added twice", []OrderBy{{Field: "id", Desc: true}}, " ORDER BY id DESC", false},
		{"unknown field", []OrderBy{{Field: "price"}}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := repo.applyOrder(base, tt.orderBy)
			if (err != nil) != tt.wantErr {
				t.Fatalf("applyOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got := strings.TrimPrefix(query, base); got != tt.wantOrder {
				t.Errorf("order clause = %q, want %q", got, tt.wantOrder)
			}
		})
	}
}

// newOrderTestRepo returns a races repository backed by an in-memory SQLite
// database with start times in different time zone offsets:
//
//	id 1: 2026-11-23T12:00:00+11:00  (01:00 UTC)
//	id 2: 2026-11-23T11:30:00+10:00  (01:30 UTC)
//	id 3: 2026-11-23T01:15:00Z       (01:15 UTC)
//	id 4: 2026-11-23T12:00:00+11:00  (01:00 UTC, same as id 1)
//
// Sorted as text these would come back 3, 2, 1, 4. Sorted as points in time
// they come back 1, 4, 3, 2, with id breaking the tie between 1 and 4.
func newOrderTestRepo(t *testing.T) *racesRepo {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	// Each connection to :memory: gets its own empty database.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	// Same schema as seed() in db.go. If the schema changes, update this too.
	if _, err := db.Exec(`CREATE TABLE races (id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, number INTEGER, visible INTEGER, advertised_start_time DATETIME)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	fixtures := []struct {
		id    int64
		start string
	}{
		{1, "2026-11-23T12:00:00+11:00"},
		{2, "2026-11-23T11:30:00+10:00"},
		{3, "2026-11-23T01:15:00Z"},
		{4, "2026-11-23T12:00:00+11:00"},
	}
	for _, f := range fixtures {
		if _, err := db.Exec(
			`INSERT INTO races(id, meeting_id, name, number, visible, advertised_start_time) VALUES (?,?,?,?,?,?)`,
			f.id, 1, "Test race", 1, 1, f.start,
		); err != nil {
			t.Fatalf("insert race %d: %v", f.id, err)
		}
	}

	return &racesRepo{db: db}
}

// TestListOrdersByStartTime runs List against a real SQLite database to check
// that start times are compared as points in time, and that id breaks ties.
func TestListOrdersByStartTime(t *testing.T) {
	repo := newOrderTestRepo(t)

	tests := []struct {
		name    string
		orderBy []OrderBy
		wantIDs []int64
	}{
		{"earliest first, id breaks the tie", []OrderBy{{Field: "advertised_start_time"}}, []int64{1, 4, 3, 2}},
		{"latest first, id still breaks the tie", []OrderBy{{Field: "advertised_start_time", Desc: true}}, []int64{2, 3, 1, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			races, err := repo.List(nil, tt.orderBy)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}

			var gotIDs []int64
			for _, race := range races {
				gotIDs = append(gotIDs, race.Id)
			}
			// Not sorted before comparing: the order is what is being tested.
			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("race IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}