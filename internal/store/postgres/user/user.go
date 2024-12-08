package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/core"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/postgres"
	constraints "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/store/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type store struct {
	*postgres.Postgres
}

func New(pg *postgres.Postgres) core.UserStore {
	return &store{pg}
}

func (s *store) GetUserByEmail(ctx context.Context, email string) (user *core.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user = new(core.User)

	stmt := `SELECT
	id,
	username,
	email,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	telephone,
	password_hash,
	is_deleted,
	created_at,
	updated_at
	FROM users WHERE email = $1`
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Pseudonym,
		&user.Telephone,
		&user.PasswordHash,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *store) GetUserByTelephone(ctx context.Context, telephone string) (user *core.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user = new(core.User)

	stmt := `SELECT
	id,
	username,
	email,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	telephone,
	password_hash,
	is_deleted,
	created_at,
	updated_at
	FROM users WHERE telephone = $1`
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		telephone).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Pseudonym,
		&user.Telephone,
		&user.PasswordHash,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *store) GetUserByUsername(ctx context.Context, username string) (user *core.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user = new(core.User)

	stmt := `SELECT
	id,
	username,
	email,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	telephone,
	password_hash,
	is_deleted,
	created_at,
	updated_at
	FROM users WHERE username = $1`
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Pseudonym,
		&user.Telephone,
		&user.PasswordHash,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *store) GetUserByID(ctx context.Context, userID int) (user *core.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	user = new(core.User)

	stmt := `SELECT
	id,
	username,
	email,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	telephone,
	password_hash,
	is_deleted,
	created_at,
	updated_at
	FROM users WHERE id = $1`
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Pseudonym,
		&user.Telephone,
		&user.PasswordHash,
		&user.IsDeleted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *store) AddUser(ctx context.Context, user core.User) (userID int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO users
	(username,
	email,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	telephone,
	password_hash)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Pseudonym,
		user.Telephone,
		user.PasswordHash,
	).Scan(&userID)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		var pg *pgconn.PgError
		if ok := errors.As(err, &pg); ok && pg.Code == pgerrcode.UniqueViolation {
			switch pg.ConstraintName {
			case constraints.UniqueUsernameConstraint:
				return 0, core.ErrUsernameAlreadyExists
			case constraints.UniqueEmailConstraint:
				return 0, core.ErrEmailAlreadyExists
			case constraints.UniqueTelephoneConstraint:
				return 0, core.ErrTelephoneAlreadyExists
			default:
				return 0, core.ErrAlreadyExists
			}
		}
		return 0, err
	}

	return userID, err
}

func (s *store) UpdateUser(ctx context.Context, user core.UpdateUser) (retUser *core.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var password *string
	if user.Password != nil {
		password = &user.Password.NewPassword
	}

	retUser = new(core.User)
	stmt := `UPDATE users SET
	password_hash = COALESCE($1, password_hash),
	username = COALESCE($2, username),
	email = COALESCE($3, email),
	first_name = COALESCE($4, first_name),
	last_name = COALESCE($5, last_name),
	middle_name = COALESCE($6, middle_name),
	pseudonym = COALESCE($7, pseudonym),
	telephone = COALESCE($8, telephone),
	updated_at = DEFAULT
	WHERE id = $9
	RETURNING
	id,
	username,
	email,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	telephone,
	password_hash,
	is_deleted,
	created_at,
	updated_at`
	err = s.DB.QueryRowContext(
		ctx,
		stmt,
		password,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Pseudonym,
		user.Telephone,
		user.ID,
	).Scan(
		&retUser.ID,
		&retUser.Username,
		&retUser.Email,
		&retUser.FirstName,
		&retUser.LastName,
		&retUser.MiddleName,
		&retUser.Pseudonym,
		&retUser.Telephone,
		&retUser.PasswordHash,
		&retUser.IsDeleted,
		&retUser.CreatedAt,
		&retUser.UpdatedAt,
	)
	if err != nil {
		logger.Log().Debug(ctx, err.Error())
		var pg *pgconn.PgError
		if ok := errors.As(err, &pg); ok && pg.Code == pgerrcode.UniqueViolation {
			switch pg.ConstraintName {
			case constraints.UniqueUsernameConstraint:
				return nil, core.ErrUsernameAlreadyExists
			case constraints.UniqueEmailConstraint:
				return nil, core.ErrEmailAlreadyExists
			case constraints.UniqueTelephoneConstraint:
				return nil, core.ErrTelephoneAlreadyExists
			default:
				return nil, core.ErrAlreadyExists
			}
		}
		return nil, err
	}

	return retUser, nil
}

func (s *store) DeleteUser(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `UPDATE users SET
	email = null,
	telephone = null,
	is_deleted = true,
	updated_at = DEFAULT
	WHERE id = $1`

	err := s.DB.QueryRowContext(ctx, stmt, userID).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *store) AddExternalUser(ctx context.Context, user core.User, externalUser core.ExternalUser) (userID int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := s.DB.Begin()
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback() // nolint
		} else {
			tx.Commit() // nolint
		}
	}()

	stmt := `INSERT INTO users
	(username,
	first_name,
	last_name,
	middle_name,
	pseudonym,
	password_hash)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err = tx.QueryRowContext(
		ctx,
		stmt,
		fmt.Sprintf("%s%d", user.Username, rand.Int()),
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Pseudonym,
		user.PasswordHash,
	).Scan(&userID)

	if err != nil {
		logger.Log().Error(ctx, err.Error())
		var pg *pgconn.PgError
		if ok := errors.As(err, &pg); ok && pg.ConstraintName == constraints.UniqueUsernameConstraint {
			return 0, core.ErrUsernameAlreadyExists
		}
		return 0, err
	}

	stmt = `INSERT INTO external_users
	(external_id, user_id, auth_provider)
	VALUES ($1, $2, $3)`

	_, err = tx.ExecContext(ctx, stmt, &externalUser.ExternalID, &userID, &externalUser.AuthProvider)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		var pg *pgconn.PgError
		if ok := errors.As(err, &pg); ok && pg.ConstraintName == constraints.UniqueUserIDAuthProviderConstraint {
			return 0, core.ErrUserIDAuthProviderAlreadyExists
		}
		return 0, err
	}

	return userID, err
}

func (s *store) GetUserByExternalID(ctx context.Context, externalID int) (user *core.ExternalUser, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `SELECT id, external_id, user_id, auth_provider FROM external_users WHERE external_id = $1`

	externalUser := new(core.ExternalUser)
	err = s.DB.QueryRowContext(ctx, stmt, externalID).Scan(
		&externalUser.ID,
		&externalUser.ExternalID,
		&externalUser.UserID,
		&externalUser.AuthProvider,
	)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrUserNotFound
		}
		return nil, err
	}

	return externalUser, nil
}
