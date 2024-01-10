package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "gRPC-server/iternal/app"
	"gRPC-server/iternal/domain/models"
	"gRPC-server/iternal/storage"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
)

type Storage struct {
	db *sql.DB
}

//func (s *Storage) isAdmin(ctx context.Context, userID int64) (bool, error) {
//	//TODO implement me
//	panic("implement me")
//}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	slog.Info("trying connect to db ")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	slog.Info("connected")
	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"
	stmt, err := s.db.Prepare("insert into users(email, pass_hash) values(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"
	stmt, err := s.db.Prepare("select id, email, pass_hash from users where email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)

	}
	row := stmt.QueryRowContext(ctx, email)
	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	stmt, err := s.db.Prepare("select is_admin from main.users where id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)
	var isAdmin bool

	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("select id, name, secret from apps where id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	row := stmt.QueryRowContext(ctx, appID)
	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil

}
