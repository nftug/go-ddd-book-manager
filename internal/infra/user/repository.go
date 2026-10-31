package user

import (
	"context"
	"database/sql"
	"errors"
	domain "go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/database"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
)

type repository struct{ db *bun.DB }

func NewRepository(i do.Injector) (domain.Repository, error) {
	return &repository{do.MustInvoke[*bun.DB](i)}, nil
}

var _ domain.Repository = (*repository)(nil)

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m model
	err := r.db.NewSelect().Model(&m).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u, err := m.toDomain()
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) Create(ctx context.Context, u domain.User) error {
	var m model
	m.fromDomain(u)
	_, err := r.db.NewInsert().Model(&m).Exec(ctx)
	return database.ConflictOnDuplicate(err)
}

func (r *repository) Update(ctx context.Context, u domain.User) error {
	var m model
	m.fromDomain(u)
	res, err := r.db.NewUpdate().
		Model(&m).
		Set("name = ?", m.Name).
		Set("email = ?", m.Email).
		Set("role = ?", m.Role).
		Set("updated_by_id = ?", m.UpdatedByID).
		Set("updated_by_name = ?", m.UpdatedByName).
		Set("updated_at = ?", m.UpdatedAt).
		Set("version = version + 1").
		Where("id = ? AND version = ?", m.ID, u.Audit().Version()).
		Exec(ctx)
	if err != nil {
		return err
	}
	if err := database.ValidateNoConflict(res); err != nil {
		return err
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, u domain.User) error {
	res, err := r.db.NewDelete().
		Model((*model)(nil)).
		Where("id = ? AND version = ?", u.ID(), u.Audit().Version()).
		Exec(ctx)
	if err != nil {
		return err
	}
	return database.ValidateNoConflict(res)
}
