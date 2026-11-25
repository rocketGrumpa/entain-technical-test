package service

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// eventsRepoStub returns events and err from List.
type eventsRepoStub struct {
	events []*sports.Event
	err    error
}

func (s *eventsRepoStub) Init() error { return nil }

func (s *eventsRepoStub) List() ([]*sports.Event, error) {
	return s.events, s.err
}

func TestListEventsReturnsEvents(t *testing.T) {
	repo := &eventsRepoStub{events: []*sports.Event{
		{Id: 1, Name: "New York Mets vs Atlanta Braves", Sport: "Baseball"},
		{Id: 2, Name: "Seattle Mariners vs Texas Rangers", Sport: "Baseball"},
	}}

	resp, err := NewSportsService(repo).ListEvents(context.Background(), &sports.ListEventsRequest{})
	if err != nil {
		t.Fatalf("ListEvents() error: %v", err)
	}
	if len(resp.Events) != 2 || resp.Events[0].Id != 1 || resp.Events[1].Id != 2 {
		t.Errorf("ListEvents() returned %v, want events 1 and 2", resp.Events)
	}
}

// TestListEventsHidesDatabaseErrors checks that a repository error reaches
// the client as INTERNAL with a general message, not the database's own text.
func TestListEventsHidesDatabaseErrors(t *testing.T) {
	repo := &eventsRepoStub{err: errors.New("no such table: events")}

	_, err := NewSportsService(repo).ListEvents(context.Background(), &sports.ListEventsRequest{})

	if got := status.Code(err); got != codes.Internal {
		t.Errorf("status code = %v, want %v", got, codes.Internal)
	}
	if msg := status.Convert(err).Message(); strings.Contains(msg, "no such table") {
		t.Errorf("error message %q shows the database error to the client", msg)
	}
}
