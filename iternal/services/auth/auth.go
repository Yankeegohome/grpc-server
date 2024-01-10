package auth

import (
	"context"
	"errors"
	"fmt"
	"gRPC-server/iternal/domain/models"
	"gRPC-server/iternal/lib/jwt"
	"gRPC-server/iternal/storage"
	"gRPC-server/iternal/storage/sqlite"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}
type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}
type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrInvalidAppID       = errors.New("invalid application id")
	ErrUserExists         = errors.New("user already exists")
)

// New создает новый инстанс для сервиса Auth
func New(log *slog.Logger, userSaver *sqlite.Storage, userProvider *sqlite.Storage, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}

}

func (a *Auth) Login(
	ctx context.Context,
	email, password string,
	appID int,
) (string, error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)
	log.Info("attempting to login user")
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", err)
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		a.log.Error("failed to get user", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", err)
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return token, nil
}

// RegisterNewUser регистрирует новых польвзателей и возращает уникальный ID
// если пользователь уже зарегистрирован выкидывает ошибку
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email, password string,
) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)
	log.Info("registering user")
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", err)
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("#{op}: #{err}")
	}
	log.Info("user succeed register ")
	return id, nil
}

// IsAdmin проверяет является ли пользователь админом
func (a *Auth) IsAdmin(
	ctx context.Context,
	appID int64,
) (bool, error) {
	const op = "Auth.IsAdmin"
	log := a.log.With(
		slog.String("op", op),
		slog.String("isAdmin", string(appID)),
	)
	log.Info("check if user is admin")

	admin, err := a.usrProvider.IsAdmin(ctx, int64(appID))
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", err)
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)

		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("...", slog.Bool("is_admin", admin))
	return admin, nil

}
