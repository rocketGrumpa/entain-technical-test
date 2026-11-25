package service

import (
	"reflect"
	"testing"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParseOrderBy(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []db.OrderBy
		wantErr bool
	}{
		{"empty uses the default order", "", []db.OrderBy{{Field: "advertised_start_time"}}, false},
		{"blank uses the default order", "   ", []db.OrderBy{{Field: "advertised_start_time"}}, false},
		{"single field ascending", "name", []db.OrderBy{{Field: "name"}}, false},
		{"single field descending", "name desc", []db.OrderBy{{Field: "name", Desc: true}}, false},
		{"several fields with extra spaces", " meeting_id , advertised_start_time desc ",
			[]db.OrderBy{{Field: "meeting_id"}, {Field: "advertised_start_time", Desc: true}}, false},
		{"unknown field", "price", nil, true},
		{"uppercase direction", "name DESC", nil, true},
		{"asc is not accepted", "name asc", nil, true},
		{"extra word", "name desc extra", nil, true},
		{"empty entry", "name,,id", nil, true},
		{"trailing comma", "name,", nil, true},
		{"repeated field", "id, id desc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOrderBy(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOrderBy(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseOrderBy(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// fakeRepo records whether List was called, so the test can check that an
// invalid order_by never reaches the repository.
type fakeRepo struct {
	listCalled bool
}

func (f *fakeRepo) Init() error { return nil }

func (f *fakeRepo) List(*racing.ListRacesRequestFilter, []db.OrderBy) ([]*racing.Race, error) {
	f.listCalled = true
	return nil, nil
}

func (f *fakeRepo) Get(int64) (*racing.Race, error) { return nil, nil }

func TestListRacesRejectsInvalidOrderBy(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewRacingService(repo)

	_, err := svc.ListRaces(context.Background(), &racing.ListRacesRequest{OrderBy: "price"})

	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("status code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
	if repo.listCalled {
		t.Error("List was called with an invalid order_by")
	}
}
