package service

import (
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestRaceStatus checks the rule itself. now is fixed, so the result never
// depends on when the test runs.
func TestRaceStatus(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		start time.Time
		want  racing.Race_Status
	}{
		{"one second ago is closed", now.Add(-time.Second), racing.Race_CLOSED},
		{"exactly now is open", now, racing.Race_OPEN},
		{"one second from now is open", now.Add(time.Second), racing.Race_OPEN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := raceStatus(tt.start, now); got != tt.want {
				t.Errorf("raceStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// racesRepoStub returns a fixed list of races from List.
type racesRepoStub struct {
	races []*racing.Race
}

func (s *racesRepoStub) Init() error { return nil }

func (s *racesRepoStub) List(*racing.ListRacesRequestFilter, []db.OrderBy) ([]*racing.Race, error) {
	return s.races, nil
}

func (s *racesRepoStub) Get(int64) (*racing.Race, error) { return nil, nil }

// TestListRacesSetsStatus checks that ListRaces sets a status on every race.
// The start times are far in the past and far in the future, so the result
// doesn't depend on when the test runs.
func TestListRacesSetsStatus(t *testing.T) {
	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC)
	repo := &racesRepoStub{races: []*racing.Race{
		{Id: 1, AdvertisedStartTime: timestamppb.New(past)},
		{Id: 2, AdvertisedStartTime: timestamppb.New(future)},
	}}

	resp, err := NewRacingService(repo).ListRaces(context.Background(), &racing.ListRacesRequest{})
	if err != nil {
		t.Fatalf("ListRaces() error: %v", err)
	}

	want := map[int64]racing.Race_Status{1: racing.Race_CLOSED, 2: racing.Race_OPEN}
	if len(resp.Races) != len(want) {
		t.Fatalf("got %d races, want %d", len(resp.Races), len(want))
	}
	for _, race := range resp.Races {
		if race.Status != want[race.Id] {
			t.Errorf("race %d status = %v, want %v", race.Id, race.Status, want[race.Id])
		}
	}
}
