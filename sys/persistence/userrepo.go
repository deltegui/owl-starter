package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/deltegui/owl-starter/domain/user"

	"github.com/deltegui/owl/v3/core"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) UserRepository { return UserRepository{db} }

func (r UserRepository) ByEmail(ctx context.Context, email string) (user.User, error) {
	var u user.User
	var role int64
	err := r.db.QueryRowContext(ctx,
		`select id, name, email, password_hash, role, created_at from users where email = $1`, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	u.Role = core.Role(role)
	return u, err
}

func (r UserRepository) Create(ctx context.Context, u user.User) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`insert into users (name, email, password_hash, role) values ($1, $2, $3, $4) returning id`,
		u.Name, u.Email, u.PasswordHash, int64(u.Role)).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, user.ErrEmailTaken
	}
	return id, err
}

func (r UserRepository) SetPasswordHash(ctx context.Context, id int64, hash string) error {
	_, err := r.db.ExecContext(ctx, `update users set password_hash = $1 where id = $2`, hash, id)
	return err
}

func (r UserRepository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `select count(*) from users`).Scan(&n)
	return n, err
}
