package service

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
)

// Sports is the gRPC service for sports events.
type Sports interface {
	// ListEvents will return a collection of all events.
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
}

// sportsService implements the Sports interface.
type sportsService struct {
	eventsRepo db.EventsRepo
}

// NewSportsService instantiates and returns a new sportsService.
func NewSportsService(eventsRepo db.EventsRepo) Sports {
	return &sportsService{eventsRepo}
}

func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	events, err := s.eventsRepo.List()
	if err != nil {
		// Log the real error for us, and send the client a general message,
		// so database details don't leak out of the service.
		log.Printf("list events: %v", err)
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	return &sports.ListEventsResponse{Events: events}, nil
}
