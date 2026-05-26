package agent

import (
	"context"
	"errors"
	"testing"
)

// fakeRepo is an in-memory Repository for unit-testing the service in isolation.
type fakeRepo struct {
	agents []Agent
	err    error
}

func (f *fakeRepo) List(ctx context.Context) ([]Agent, error) {
	return f.agents, f.err
}

func TestServiceListReturnsAgentsFromRepository(t *testing.T) {
	want := []Agent{
		{ID: 1, Name: "Sam Carter", Email: "sam@isp.example"},
		{ID: 2, Name: "Robin Diaz", Email: "robin@isp.example"},
	}
	svc := NewService(&fakeRepo{agents: want})

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("List returned %d agents, want %d", len(got), len(want))
	}
	if got[0].Name != "Sam Carter" {
		t.Errorf("agent[0].Name = %q, want %q", got[0].Name, "Sam Carter")
	}
}

func TestServiceListPropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	svc := NewService(&fakeRepo{err: repoErr})

	_, err := svc.List(context.Background())
	if !errors.Is(err, repoErr) {
		t.Fatalf("List error = %v, want %v", err, repoErr)
	}
}
