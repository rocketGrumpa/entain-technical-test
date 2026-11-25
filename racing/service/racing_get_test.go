package service

import (
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// getRepoStub returns race and err from Get, and records whether Get was called.
type getRepoStub struct {
	race      *racing.Race
	err       error
	getCalled bool
}

func (s *getRepoStub) Init() error { return nil }

func (s *getRepoStub) List(*racing.ListRacesRequestFilter, []db.OrderBy) ([]*racing.Race, error) {
	return nil, nil
}

func (s *getRepoStub) Get(int64) (*racing.Race, error) {
	s.getCalled = true
	return s.race, s.err
}

func TestGetRace(t *testing.T) {
	past := timestamppb.New(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name         string
		id           int64
		repo         *getRepoStub
		wantCode     codes.Code
		wantStatus   racing.Race_Status
		wantRepoCall bool
	}{
		{"found race gets a status", 1, &getRepoStub{race: &racing.Race{Id: 1, AdvertisedStartTime: past}},
			codes.OK, racing.Race_CLOSED, true},
		{"missing race is not found", 99, &getRepoStub{err: db.ErrNotFound},
			codes.NotFound, racing.Race_STATUS_UNSPECIFIED, true},
		{"zero id is invalid", 0, &getRepoStub{},
			codes.InvalidArgument, racing.Race_STATUS_UNSPECIFIED, false},
		{"negative id is invalid", -1, &getRepoStub{},
			codes.InvalidArgument, racing.Race_STATUS_UNSPECIFIED, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			race, err := NewRacingService(tt.repo).GetRace(context.Background(), &racing.GetRaceRequest{Id: tt.id})

			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("status code = %v, want %v", got, tt.wantCode)
			}
			if tt.repo.getCalled != tt.wantRepoCall {
				t.Errorf("repository called = %v, want %v", tt.repo.getCalled, tt.wantRepoCall)
			}
			if err == nil && race.Status != tt.wantStatus {
				t.Errorf("race status = %v, want %v", race.Status, tt.wantStatus)
			}
		})
	}
}
