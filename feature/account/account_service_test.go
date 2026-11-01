package account

import (
	"context"
	"errors"
	"testing"

	"github.com/deltegui/owl/v3/hash"

	"github.com/deltegui/owl-starter/domain/user"
)

type memUsers map[string]user.User

func (m memUsers) ByEmail(_ context.Context, email string) (user.User, error) {
	u, ok := m[email]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}
func (m memUsers) Create(context.Context, user.User) (int64, error) { return 0, nil }
func (m memUsers) SetPasswordHash(_ context.Context, id int64, h string) error {
	for k, u := range m {
		if u.ID == id {
			u.PasswordHash = h
			m[k] = u
		}
	}
	return nil
}
func (m memUsers) Count(context.Context) (int, error) { return len(m), nil }

func TestAuthenticate(t *testing.T) {
	hasher := hash.Argon2{Memory: 8, Time: 1, Threads: 1} // small, so the test is fast
	hashed, err := hasher.Hash("secret")
	if err != nil {
		t.Fatal(err)
	}
	svc := newService(memUsers{"a@b.c": {ID: 1, Email: "a@b.c", PasswordHash: hashed}}, hasher)

	if u, err := svc.authenticate(context.Background(), "a@b.c", "secret"); err != nil || u.ID != 1 {
		t.Fatalf("right password: %v %v", u, err)
	}
	for name, c := range map[string][2]string{"wrong password": {"a@b.c", "nope"}, "unknown user": {"x@y.z", "secret"}} {
		if _, err := svc.authenticate(context.Background(), c[0], c[1]); !errors.Is(err, errInvalidCredentials) {
			t.Errorf("%s: want errInvalidCredentials, got %v", name, err)
		}
	}
}
