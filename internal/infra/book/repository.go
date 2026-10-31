package book

import (
	"context"
	"database/sql"
	"errors"
	"uuid"

	domain "go-ddd-book-manager/internal/domain/book"
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

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	var result *domain.Book
	err := r.db.RunInTx(
		ctx,
		&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead},
		func(ctx context.Context, tx bun.Tx) error {
			var m model
			err := tx.NewSelect().
				Model(&m).
				Relation("Authors", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Order("order_index ASC")
				}).
				Relation("Checkouts", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Order("position ASC")
				}).
				Where("id = ?", id).
				Scan(ctx)
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			b, err := restoreBook(m)
			if err != nil {
				return err
			}
			result = &b
			return nil
		},
	)
	return result, err
}

func (r *repository) Save(ctx context.Context, b domain.Book) error {
	m := bookModelFromDomain(b)
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := ic.UpsertWithAudit(tx.NewInsert().Model(&m)).
			Set("title = EXCLUDED.title").
			Set("isbn = EXCLUDED.isbn").
			Set("description = EXCLUDED.description").
			Set("owner_id = EXCLUDED.owner_id").
			Set("owner_name = EXCLUDED.owner_name").
			Exec(ctx)
		if err != nil {
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
		for _, v := range m.Authors {
			if _, err = tx.NewInsert().Model(&v).Exec(ctx); err != nil {
				return err
			}
		}
		for _, v := range m.Checkouts {
			if _, err = tx.NewInsert().Model(&v).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().Model((*model)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}
