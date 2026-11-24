package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races, filtered by filter and sorted by orderBy.
	List(filter *racing.ListRacesRequestFilter, orderBy []OrderBy) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter, orderBy []OrderBy) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	query, err = r.applyOrder(query, orderBy)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	// Only restrict by visibility when explicitly asked to; the default
	// (false) keeps returning all races. visible is stored as an INTEGER
	// (0 or 1) because SQLite has no boolean type, so the bound value is 1.
	if filter.OnlyVisible {
		clauses = append(clauses, "visible = ?")
		args = append(args, 1)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

// OrderBy is one sort key for List, already validated by the caller.
type OrderBy struct {
	Field string // proto field name, e.g. "advertised_start_time"
	Desc  bool
}

// sortColumns maps each sortable Race field to the SQL used to sort by it.
// It is the whitelist for order_by: only these expressions reach the query,
// never the caller's text. advertised_start_time is stored as text with a
// time zone offset, so julianday() compares it as a point in time rather
// than as a string.
var sortColumns = map[string]string{
	"id":                    "id",
	"meeting_id":            "meeting_id",
	"name":                  "name",
	"number":                "number",
	"visible":               "visible",
	"advertised_start_time": "julianday(advertised_start_time)",
}

// IsSortableField reports whether field can be used in order_by.
func IsSortableField(field string) bool {
	_, ok := sortColumns[field]
	return ok
}

func (r *racesRepo) applyOrder(query string, orderBy []OrderBy) (string, error) {
	var (
		terms []string
		hasID bool
	)

	for _, o := range orderBy {
		column, ok := sortColumns[o.Field]
		if !ok {
			return "", fmt.Errorf("unsupported order field %q", o.Field)
		}

		direction := "ASC"
		if o.Desc {
			direction = "DESC"
		}
		terms = append(terms, column+" "+direction)

		if o.Field == "id" {
			hasID = true
		}
	}

	// Always finish with id so races with equal values keep a stable order.
	if !hasID {
		terms = append(terms, "id ASC")
	}

	return query + " ORDER BY " + strings.Join(terms, ", "), nil
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
