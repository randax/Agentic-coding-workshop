package identity

import (
	"context"
	"errors"
	"testing"

	"saltcrm/internal/agent"
)

// fakeUsers is an in-memory UserRepository keyed by email.
type fakeUsers struct{ byEmail map[string]agent.Agent }

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (agent.Agent, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return agent.Agent{}, agent.ErrNotFound
}
func (f *fakeUsers) FindByID(_ context.Context, id uint) (agent.Agent, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return agent.Agent{}, agent.ErrNotFound
}

// fakeSessions is an in-memory SessionRepository.
type fakeSessions struct{ byToken map[string]uint }

func (f *fakeSessions) Create(_ context.Context, s *Session) error {
	f.byToken[s.Token] = s.UserID
	return nil
}
func (f *fakeSessions) FindUserID(_ context.Context, token string) (uint, error) {
	if id, ok := f.byToken[token]; ok {
		return id, nil
	}
	return 0, ErrNoSession
}
func (f *fakeSessions) Delete(_ context.Context, token string) error {
	delete(f.byToken, token)
	return nil
}

func newService(t *testing.T) (*Service, agent.Agent) {
	t.Helper()
	hash, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	u := agent.Agent{ID: 1, Name: "Sam Carter", Email: "sam@isp.example", PasswordHash: hash, Role: agent.RoleAgent}
	users := &fakeUsers{byEmail: map[string]agent.Agent{u.Email: u}}
	sessions := &fakeSessions{byToken: map[string]uint{}}
	return NewService(users, sessions), u
}

func TestAuthenticateAcceptsCorrectPassword(t *testing.T) {
	svc, u := newService(t)

	got, err := svc.Authenticate(context.Background(), u.Email, "s3cret")
	if err != nil {
		t.Fatalf("Authenticate returned error: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("authenticated user = %+v, want id %d", got, u.ID)
	}
}

func TestAuthenticateRejectsWrongPasswordAndUnknownEmail(t *testing.T) {
	svc, u := newService(t)

	if _, err := svc.Authenticate(context.Background(), u.Email, "nope"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("wrong password error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := svc.Authenticate(context.Background(), "ghost@isp.example", "s3cret"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("unknown email error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginThenCurrentUserResolvesSession(t *testing.T) {
	svc, u := newService(t)
	ctx := context.Background()

	loggedIn, token, err := svc.Login(ctx, u.Email, "s3cret")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if token == "" || loggedIn.ID != u.ID {
		t.Fatalf("Login = (%+v, %q), want a token and the user", loggedIn, token)
	}

	current, err := svc.CurrentUser(ctx, token)
	if err != nil {
		t.Fatalf("CurrentUser returned error: %v", err)
	}
	if current.ID != u.ID {
		t.Errorf("CurrentUser = %+v, want id %d", current, u.ID)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	svc, u := newService(t)
	ctx := context.Background()
	_, token, _ := svc.Login(ctx, u.Email, "s3cret")

	if err := svc.Logout(ctx, token); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
	if _, err := svc.CurrentUser(ctx, token); !errors.Is(err, ErrNoSession) {
		t.Errorf("CurrentUser after logout error = %v, want ErrNoSession", err)
	}
}
