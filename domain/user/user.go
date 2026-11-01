// Package user is the user of the app, shared by every feature that needs to
// know who someone is. The storage is in sys/persistence.
package user

import (
	"context"
	"errors"
	"time"

	"github.com/deltegui/owl/v3/core"
	"github.com/deltegui/owl/v3/session"
)

var ErrNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already registered")

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         core.Role
	CreatedAt    time.Time
}

type Repository interface {
	// ByEmail returns ErrNotFound if there is no user.
	ByEmail(ctx context.Context, email string) (User, error)
	// Create returns ErrEmailTaken if the email is already registered.
	Create(ctx context.Context, u User) (int64, error)
	SetPasswordHash(ctx context.Context, id int64, hash string) error
	Count(ctx context.Context) (int, error)
}

// ClaimName is the key of the session claim that holds the name of the user.
const ClaimName session.Claim = "Name"
