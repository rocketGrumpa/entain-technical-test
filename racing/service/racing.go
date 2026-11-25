package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Racing interface {
	// ListRaces will return a collection of races.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
	// GetRace will return a single race by its id.
	GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.Race, error)
}

// racingService implements the Racing interface.
type racingService struct {
	racesRepo db.RacesRepo
}

// NewRacingService instantiates and returns a new racingService.
func NewRacingService(racesRepo db.RacesRepo) Racing {
	return &racingService{racesRepo}
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	// Validate order_by here, before the repository, so a bad value is
	// reported to the caller as INVALID_ARGUMENT instead of being ignored.
	orderBy, err := parseOrderBy(in.GetOrderBy())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	races, err := s.racesRepo.List(in.Filter, orderBy)
	if err != nil {
		return nil, err
	}

	// Read the clock once, so every race in this response is judged at the
	// same moment, even if the loop runs across a second boundary.
	now := time.Now()
	for _, race := range races {
		race.Status = raceStatus(race.GetAdvertisedStartTime().AsTime(), now)
	}

	return &racing.ListRacesResponse{Races: races}, nil
}

func (s *racingService) GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.Race, error) {
	// Zero is the proto3 default, so an id of zero usually means the caller
	// forgot to set it. Report it as a bad request rather than NOT_FOUND, so
	// the caller can tell a mistake in the request from a missing race.
	if in.GetId() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id must be greater than zero, got %d", in.GetId())
	}

	race, err := s.racesRepo.Get(in.GetId())
	if errors.Is(err, db.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "race %d not found", in.GetId())
	}
	if err != nil {
		return nil, err
	}

	race.Status = raceStatus(race.GetAdvertisedStartTime().AsTime(), time.Now())

	return race, nil
}

// raceStatus returns CLOSED when start is in the past and OPEN otherwise.
// A race that starts exactly at now is still OPEN, because it isn't in the
// past yet.
func raceStatus(start, now time.Time) racing.Race_Status {
	if start.Before(now) {
		return racing.Race_CLOSED
	}
	return racing.Race_OPEN
}

// parseOrderBy turns an order_by value such as
// "meeting_id, advertised_start_time desc" into validated sort keys,
// following the AIP-132 format. An empty value means the default order:
// earliest advertised start time first.
func parseOrderBy(orderBy string) ([]db.OrderBy, error) {
	if strings.TrimSpace(orderBy) == "" {
		return []db.OrderBy{{Field: "advertised_start_time"}}, nil
	}

	var (
		keys []db.OrderBy
		seen = make(map[string]bool)
	)

	for _, part := range strings.Split(orderBy, ",") {
		words := strings.Fields(part)
		if len(words) == 0 || len(words) > 2 {
			return nil, fmt.Errorf("invalid order_by entry %q", strings.TrimSpace(part))
		}

		field := words[0]
		if !db.IsSortableField(field) {
			return nil, fmt.Errorf("unsupported order_by field %q", field)
		}
		// A repeated field is ambiguous ("id, id desc") and also keeps the
		// number of sort keys bounded by the number of sortable fields.
		if seen[field] {
			return nil, fmt.Errorf("order_by field %q is listed more than once", field)
		}
		seen[field] = true

		desc := false
		if len(words) == 2 {
			if words[1] != "desc" {
				return nil, fmt.Errorf("unsupported order_by direction %q: only \"desc\" is allowed", words[1])
			}
			desc = true
		}

		keys = append(keys, db.OrderBy{Field: field, Desc: desc})
	}

	return keys, nil
}
