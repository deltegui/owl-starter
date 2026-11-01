package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/deltegui/owl/v3/session"
)

// SessionStore keeps the sessions of OWL's session.Manager in Postgres. The
// Manager checks the expiration, so expired rows only need to be swept now and
// then (Sweep).
type SessionStore struct{ db *sql.DB }

func NewSessionStore(db *sql.DB) SessionStore { return SessionStore{db} }

func (s SessionStore) Save(ctx context.Context, e session.Entry) error {
	value, err := json.Marshal(e.User)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`insert into sessions (key, user_id, value, created_at, expires_at) values ($1, $2, $3, $4, $5)
		 on conflict (key) do update set value = excluded.value, expires_at = excluded.expires_at`,
		e.Key, e.User.Id, value, e.Created, e.Expires)
	return err
}

func (s SessionStore) Get(ctx context.Context, key string) (session.Entry, error) {
	e := session.Entry{Key: key}
	var value []byte
	err := s.db.QueryRowContext(ctx, `select value, created_at, expires_at from sessions where key = $1`, key).
		Scan(&value, &e.Created, &e.Expires)
	if errors.Is(err, sql.ErrNoRows) {
		return session.Entry{}, session.ErrNotFound
	}
	if err != nil {
		return session.Entry{}, err
	}
	return e, json.Unmarshal(value, &e.User)
}

func (s SessionStore) Delete(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, `delete from sessions where key = $1`, key)
	return err
}

func (s SessionStore) DeleteUser(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `delete from sessions where user_id = $1`, userID)
	return err
}

// Sweep removes the expired sessions.
func (s SessionStore) Sweep(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `delete from sessions where expires_at <= now()`)
	return err
}
