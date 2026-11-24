package db

import (
	"database/sql"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// TestApplyFilter checks the SQL and arguments applyFilter builds.
func TestApplyFilter(t *testing.T) {
	repo := &racesRepo{}
	base := "SELECT id FROM races"

	tests := []struct {
		name      string
		filter    *racing.ListRacesRequestFilter
		wantWhere string
		wantArgs  []interface{}
	}{
		{"nil filter returns all races", nil, "", nil},
		{"empty filter returns all races", &racing.ListRacesRequestFilter{}, "", nil},
		{"only visible", &racing.ListRacesRequestFilter{OnlyVisible: true}, " WHERE visible = ?", []interface{}{1}},
		{"meeting ids and only visible",
			&racing.ListRacesRequestFilter{MeetingIds: []int64{1, 2}, OnlyVisible: true},
			" WHERE meeting_id IN (?,?) AND visible = ?", []interface{}{int64(1), int64(2), 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := repo.applyFilter(base, tt.filter)

			if got := strings.TrimPrefix(query, base); got != tt.wantWhere {
				t.Errorf("where clause = %q, want %q", got, tt.wantWhere)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}

// newTestRepo returns a races repository backed by an in-memory SQLite
// database holding three fixed races:
//
//	id 1: meeting 1, visible
//	id 2: meeting 1, hidden
//	id 3: meeting 2, visible
func newTestRepo(t *testing.T) *racesRepo {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	// Each connection to :memory: gets its own empty database, so keep a
	// single connection to make sure every query sees the same table.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	// Same schema as seed() in db.go. Written out here instead of calling
	// Init(), which would also insert 100 random races. If the schema in
	// db.go changes, update this too.
	if _, err := db.Exec(`CREATE TABLE races (id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, number INTEGER, visible INTEGER, advertised_start_time DATETIME)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	start := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	fixtures := []struct {
		id, meetingID, visible int64
	}{
		{1, 1, 1},
		{2, 1, 0},
		{3, 2, 1},
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

// TestListFiltersByVisibility runs List against a real SQLite database,
// so it also catches mistakes the SQL string check above cannot, such as
// a misspelled column name.
func TestListFiltersByVisibility(t *testing.T) {
	repo := newTestRepo(t)

	tests := []struct {
		name    string
		filter  *racing.ListRacesRequestFilter
		wantIDs []int64
	}{
		{"nil filter returns all races", nil, []int64{1, 2, 3}},
		{"empty filter returns all races", &racing.ListRacesRequestFilter{}, []int64{1, 2, 3}},
		{"only visible", &racing.ListRacesRequestFilter{OnlyVisible: true}, []int64{1, 3}},
		{"meeting 1 and only visible",
			&racing.ListRacesRequestFilter{MeetingIds: []int64{1}, OnlyVisible: true}, []int64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			races, err := repo.List(tt.filter, nil)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}

			var gotIDs []int64
			for _, race := range races {
				gotIDs = append(gotIDs, race.Id)
			}
			// Sorted so this test only checks filtering, not ordering.
			sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })

			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("race IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}
