package book

import (
	"context"
	"database/sql"
	domain "go-ddd-book-manager/internal/domain/book"
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

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	var result *domain.Book
	err := r.db.RunInTx(
		ctx,
		&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead},
		func(ctx context.Context, tx bun.Tx) error {
			var m model
			err := relations(tx.NewSelect().Model(&m)).Where("id = ?", id).Scan(ctx)
			if err != nil {
				return database.IgnoreNoRows(err)
			}
			b, err := m.toDomain()
			if err != nil {
				return err
			}
			result = &b
			return nil
		},
	)
	return result, err
}

func (r *repository) Create(ctx context.Context, b domain.Book) error {
	var m model
	m.fromDomain(b)
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(&m).Exec(ctx); err != nil {
			return database.ConflictOnDuplicate(err)
		}
		return insertRelations(ctx, tx, &m)
	})
}

func (r *repository) Update(ctx context.Context, b domain.Book) error {
	var m model
	m.fromDomain(b)
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		res, err := tx.NewUpdate().
			Model(&m).
			Set("title = ?", m.Title).
			Set("isbn = ?", m.ISBN).
			Set("description = ?", m.Description).
			Set("owner_id = ?", m.OwnerID).
			Set("owner_name = ?", m.OwnerName).
			Set("updated_by_id = ?", m.UpdatedByID).
			Set("updated_by_name = ?", m.UpdatedByName).
			Set("updated_at = ?", m.UpdatedAt).
			Set("version = version + 1").
			Where("id = ? AND version = ?", m.ID, b.Audit().Version()).
			Exec(ctx)
		if err != nil {
			return err
		}
		if err := database.ValidateNoConflict(res); err != nil {
			return err
		}
		if _, err = tx.NewDelete().
			Model((*authorModel)(nil)).
			Where("book_id = ?", m.ID).
			Exec(ctx); err != nil {
			return err
		}
		if _, err = tx.NewDelete().
			Model((*checkoutModel)(nil)).
			Where("book_id = ?", m.ID).
			Exec(ctx); err != nil {
			return err
		}
		return insertRelations(ctx, tx, &m)
	})
}

func insertRelations(ctx context.Context, tx bun.Tx, m *model) error {
	if len(m.Authors) > 0 {
		if _, err := tx.NewInsert().Model(&m.Authors).Exec(ctx); err != nil {
			return err
		}
	}
	if len(m.Checkouts) > 0 {
		if _, err := tx.NewInsert().Model(&m.Checkouts).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, b domain.Book) error {
	res, err := r.db.NewDelete().
		Model((*model)(nil)).
		Where("id = ? AND version = ?", b.ID(), b.Audit().Version()).
		Exec(ctx)
	if err != nil {
		return err
	}
	return database.ValidateNoConflict(res)
}
