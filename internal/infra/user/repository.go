package user

import (
	"context"
	"database/sql"
	"errors"
	"uuid"

	domain "go-ddd-book-manager/internal/domain/user"
	ic "go-ddd-book-manager/internal/infra/common"

	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
)

type repository struct{ db *bun.DB }

func NewRepository(i do.Injector) (domain.Repository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return &repository{db: db}, nil
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
	u, err := restoreUser(m)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *repository) Save(ctx context.Context, u domain.User) error {
	m := userModelFromDomain(u)
	_, err := ic.UpsertWithAudit(r.db.NewInsert().Model(&m)).
		Set("name = EXCLUDED.name").
		Set("email = EXCLUDED.email").
		Set("role = EXCLUDED.role").
		Exec(ctx)
	return err
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*model)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
