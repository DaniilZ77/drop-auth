package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/db/generated"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/logger"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/lib/postgres"
	"github.com/MAXXXIMUS-tropical-milkshake/beatflow-auth/internal/model"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserStore struct {
	*postgres.Postgres
	*generated.Queries
}

func NewUserStore(pg *postgres.Postgres) *UserStore {
	return &UserStore{pg, generated.New(pg.DB)}
}

func (s *UserStore) UpdateUser(ctx context.Context, updateUser generated.UpdateUserParams) (*generated.User, error) {
	user, err := s.Queries.UpdateUser(ctx, updateUser)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
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
	if params.OrderBy != nil {
		query = query.OrderBy(fmt.Sprintf("%q %s", params.OrderBy.Field, params.OrderBy.Order))
	}

	count := builder.Select("count(*)").FromSelect(query, "u")
	sql, args, err := count.ToSql()
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	err = s.DB.QueryRow(ctx, sql, args...).Scan(&total)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	query = query.Limit(uint64(params.Limit)).Offset(uint64(params.Offset))

	sql, args, err = query.ToSql()
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	rows, err := s.DB.Query(ctx, sql, args...)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}
	defer rows.Close()

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[generated.User])
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, nil, err
	}

	return users, total, nil
}

func (s *UserStore) SaveUser(ctx context.Context, user generated.SaveUserParams) (*uuid.UUID, error) {
	id, err := s.Queries.SaveUser(ctx, user)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		return nil, err
	}

	return &id, nil
}

func (s *UserStore) GetUserByExternalID(ctx context.Context, id int32) (*generated.User, error) {
	user, err := s.Queries.GetUserByExternalID(ctx, id)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id uuid.UUID) (*generated.User, error) {
	user, err := s.Queries.GetUserByID(ctx, id)
	if err != nil {
		logger.Log().Error(ctx, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
