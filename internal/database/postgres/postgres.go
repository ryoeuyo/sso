package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ryoeuyo/sso/internal/database"
	"github.com/ryoeuyo/sso/internal/domain/entity"
)

type Database struct {
	db *sql.DB
}

func New(db *sql.DB) *Database {
	return &Database{
		db: db,
	}
}

func (d *Database) Stop() error {
	return d.db.Close()
}

func (d *Database) Save(ctx context.Context, login string, passHash []byte) (int64, error) {
	const fn = "postgres.Save"

	stmt, err := d.db.Prepare("INSERT INTO users(login, passHash) VALUES (?, ?) RETURNS id")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	res, err := stmt.ExecContext(ctx, login, passHash)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique violation
				return 0, fmt.Errorf("%s: %w", fn, database.ErrLoginIsExists)
			default:
				return 0, fmt.Errorf("%s: PostgresSQL error %s (%s): %s", fn, pgErr.Message, pgErr.Code, err)
			}
		}

		// if err is not postgres error, returns common error
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (d *Database) User(ctx context.Context, login string) (*entity.User, error) {
	const fn = "postgres.User"

	stmt, err := d.db.Prepare("SELECT id, login, passHash FROM users WHERE login = ?")
	if err != nil {
		return &entity.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	row := stmt.QueryRowContext(ctx, login)

	var user entity.User
	err = row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &entity.User{}, fmt.Errorf("%s: %w", fn, database.ErrUserIsNotExists)
		}

		return &entity.User{}, fmt.Errorf("%s: %w", fn, err)
	}

	return &user, nil
}
