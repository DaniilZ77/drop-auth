package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	sl "github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/postgres"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserStore struct {
	*postgres.Postgres
	*generated.Queries
	log *slog.Logger
}

func NewUserStore(pg *postgres.Postgres, log *slog.Logger) *UserStore {
	return &UserStore{pg, generated.New(pg.DB), log}
}

func (s *UserStore) UpdateUser(ctx context.Context, updateUser generated.UpdateUserParams) (*generated.User, error) {
	user, err := s.Queries.UpdateUser(ctx, updateUser)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetUsers(ctx context.Context, params model.GetUsersParams) (users []generated.User, total *int, err error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := builder.Select("*").From("users")

	if params.UserID != nil {
		query = query.Where(sq.Eq{"id": *params.UserID})
	}
	if params.Username != nil {
		query = query.Where(sq.Eq{"username": *params.Username})
	}
	if params.Pseudonym != nil {
		query = query.Where(sq.Eq{"pseudonym": *params.Pseudonym})
	}
	if params.FirstName != nil {
		query = query.Where(sq.Eq{"first_name": *params.FirstName})
	}
	if params.LastName != nil {
		query = query.Where(sq.Eq{"last_name": *params.LastName})
	}

	count := builder.Select("count(*)").FromSelect(query, "u")
	stmt, args, err := count.ToSql()
	if err != nil {
		s.log.Error("failed to convert to sql", sl.Err(err))
		return nil, nil, err
	}

	err = s.DB.QueryRow(ctx, stmt, args...).Scan(&total)
	if err != nil {
		s.log.Error("failed to count users", sl.Err(err))
		return nil, nil, err
	}

	if params.OrderBy != nil {
		query = query.OrderBy(fmt.Sprintf("%q %s", params.OrderBy.Field, params.OrderBy.Order))
	}

	query = query.Limit(uint64(params.Limit)).Offset(uint64(params.Offset)) // nolint

	stmt, args, err = query.ToSql()
	if err != nil {
		s.log.Error("failed to convert to sql", sl.Err(err))
		return nil, nil, err
	}

	rows, err := s.DB.Query(ctx, stmt, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, nil, model.ErrOrderByInvalidField
		}
		s.log.Error("failed to query users", sl.Err(err))
		return nil, nil, err
	}
	defer rows.Close()

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[generated.User])
	if err != nil {
		s.log.Error("failed to scan users", sl.Err(err))
		return nil, nil, err
	}

	return users, total, nil
}

func (s *UserStore) SaveUser(ctx context.Context, user generated.SaveUserParams) (*uuid.UUID, error) {
	id, err := s.Queries.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id uuid.UUID) (*generated.User, error) {
	user, err := s.Queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetUserByUsername(ctx context.Context, username string) (*generated.User, error) {
	user, err := s.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) SaveAdmin(ctx context.Context, params generated.SaveAdminParams) error {
	if err := s.Queries.SaveAdmin(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrAdminAlreadyExists
		}
		return err
	}

	return nil
}

func (s *UserStore) DeleteAdmin(ctx context.Context, userID uuid.UUID) error {
	if err := s.Queries.DeleteAdmin(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetAdminByID(ctx context.Context, id uuid.UUID) (*generated.GetAdminByIDRow, error) {
	admin, err := s.Queries.GetAdminByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return &admin, nil
}
