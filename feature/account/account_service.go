package account

import (
	"context"
	"errors"

	"github.com/deltegui/owl/v3/core"

	"github.com/deltegui/owl-starter/domain/user"
)

var errInvalidCredentials = errors.New("invalid credentials")

type service struct {
	users  user.Repository
	hasher core.Hasher
}

func newService(users user.Repository, hasher core.Hasher) service {
	return service{users, hasher}
}

// authenticate gives the same answer, and takes the same time, for a user that
// does not exist and for a wrong password.
func (s service) authenticate(ctx context.Context, email, password string) (user.User, error) {
	u, err := s.users.ByEmail(ctx, email)
	if errors.Is(err, user.ErrNotFound) {
		s.hasher.Mimic(password)
		return user.User{}, errInvalidCredentials
	}
	if err != nil {
		return user.User{}, err
	}
	ok, err := s.hasher.Check(u.PasswordHash, password)
	if err != nil {
		return user.User{}, err // the stored hash is unreadable
	}
	if !ok {
		return user.User{}, errInvalidCredentials
	}
	if s.hasher.NeedsRehash(u.PasswordHash) {
		if fresh, err := s.hasher.Hash(password); err == nil {
			_ = s.users.SetPasswordHash(ctx, u.ID, fresh)
		}
	}
	return u, nil
}
