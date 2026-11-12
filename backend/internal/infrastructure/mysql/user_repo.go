package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"kanban/internal/domain/identity"
)

type UserRepository struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Add(ctx context.Context, u *identity.User, hash string) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO users (email, name, password_hash) VALUES (?,?,?)`, u.Email, u.Name, hash)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return identity.ErrEmailTaken
		}
		return mapErr(err)
	}
	id, _ := res.LastInsertId()
	u.ID = int(id)
	return nil
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (*identity.User, string, error) {
	var u identity.User
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT id, email, name, password_hash FROM users WHERE email=?`, email).
		Scan(&u.ID, &u.Email, &u.Name, &hash)
	if err != nil {
		return nil, "", mapErr(err)
	}
	return &u, hash, nil
}

func (r *UserRepository) ByID(ctx context.Context, id int) (*identity.User, error) {
	var u identity.User
	err := r.db.QueryRowContext(ctx, `SELECT id, email, name FROM users WHERE id=?`, id).Scan(&u.ID, &u.Email, &u.Name)
	if err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *UserRepository) List(ctx context.Context) ([]*identity.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, email, name FROM users ORDER BY name`)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []*identity.User{}
	for rows.Next() {
		var u identity.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, rows.Err()
}
