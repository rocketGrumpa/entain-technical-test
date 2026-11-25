package db

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// EventsRepo provides repository access to sports events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// List will return all events.
	List() ([]*sports.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init creates the events table and its dummy data. It only runs once.
func (r *eventsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = r.seed()
	})

	return err
}

// List returns every event. The order isn't specified.
func (r *eventsRepo) List() ([]*sports.Event, error) {
	rows, err := r.db.Query(getEventQueries()[eventsList])
	if err != nil {
		return nil, err
	}
	// Close rows on every return path, so an error part way through scanning
	// doesn't leave the connection checked out.
	defer rows.Close()

	var events []*sports.Event
	for rows.Next() {
		var (
			event sports.Event
			start time.Time
		)
		if err := rows.Scan(&event.Id, &event.Name, &start, &event.Sport, &event.Competition, &event.HomeTeam, &event.AwayTeam); err != nil {
			return nil, err
		}
		event.AdvertisedStartTime = timestamppb.New(start)

		events = append(events, &event)
	}

	// rows.Err reports an error that stopped the loop early.
	return events, rows.Err()
}
