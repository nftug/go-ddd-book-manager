package book

import (
	"context"
	"database/sql"
	dto "go-ddd-book-manager/internal/application/book/dto"
	aq "go-ddd-book-manager/internal/application/book/query"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/infra/database"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

type queryService struct{ db *bun.DB }

func NewQueryService(i do.Injector) (aq.QueryService, error) {
	return &queryService{do.MustInvoke[*bun.DB](i)}, nil
}

func (s *queryService) FindByID(
	ctx context.Context,
	actor dc.Actor,
	id uuid.UUID,
) (*dto.BookResponse, error) {
	var out *dto.BookResponse
	err := s.db.RunInTx(
		ctx,
		&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead},
		func(ctx context.Context, tx bun.Tx) error {
			var m model
			err := relations(tx.NewSelect().Model(&m)).Where("id = ?", id).Scan(ctx)
			if err != nil {
				return database.IgnoreNoRows(err)
			}
			out = new(m.toResponse(actor))
			return nil
		},
	)
	return out, err
}

func (s *queryService) List(
	ctx context.Context,
	actor dc.Actor,
	f dto.BookQuery,
) (dto.BookPage, error) {
	page := dto.BookPage{Page: f.Pagination.Page, PageSize: f.Pagination.PageSize}
	if err := s.db.RunInTx(
		ctx,
		&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead},
		func(ctx context.Context, tx bun.Tx) error {
			var err error
			page.Total, err = filterBooks(tx.NewSelect().Model((*model)(nil)), f).Count(ctx)
			if err != nil {
				return err
			}
			var rows []model
			if err := relations(
				filterBooks(tx.NewSelect().Model(&rows), f),
			).OrderExpr("model.created_at DESC, model.id ASC").
				Limit(f.Pagination.PageSize).
				Offset(f.Pagination.Offset()).
				Scan(ctx); err != nil {
				return err
			}
			page.Items = lo.Map(rows, func(m model, _ int) dto.BookResponse {
				return m.toResponse(actor)
			})
			return nil
		},
	); err != nil {
		return dto.BookPage{}, err
	}
	return page, nil
}

func filterBooks(q *bun.SelectQuery, f dto.BookQuery) *bun.SelectQuery {
	if f.Title != nil {
		q = q.Where("model.title ILIKE ?", database.LikeContainsPattern(*f.Title))
	}
	if f.Author != nil {
		q = q.Where(
			`EXISTS (
				SELECT 1
				FROM book_authors a
				WHERE a.book_id = model.id AND a.name ILIKE ?
			)`,
			database.LikeContainsPattern(*f.Author),
		)
	}
	if f.CheckedOut != nil {
		q = q.Where(
			`COALESCE((
				SELECT c.returned_at IS NULL
				FROM book_checkouts c
				WHERE c.book_id = model.id
				ORDER BY c.checked_out_at DESC, c.position DESC
				LIMIT 1
			), FALSE) = ?`,
			*f.CheckedOut,
		)
	}
	return q
}

func (s *queryService) Checkouts(
	ctx context.Context,
	id uuid.UUID,
	p ac.Pagination,
) (*dto.CheckoutPage, error) {
	var page *dto.CheckoutPage
	if err := s.db.RunInTx(
		ctx,
		&sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead},
		func(ctx context.Context, tx bun.Tx) error {
			exists, err := tx.NewSelect().Model((*model)(nil)).Where("id = ?", id).Exists(ctx)
			if err != nil || !exists {
				return err
			}
			page = &dto.CheckoutPage{Page: p.Page, PageSize: p.PageSize}
			page.Total, err = tx.NewSelect().
				Model((*checkoutModel)(nil)).
				Where("book_id = ?", id).
				Count(ctx)
			if err != nil {
				return err
			}
			var rows []checkoutModel
			if err := tx.NewSelect().
				Model(&rows).
				Where("book_id = ?", id).
				OrderExpr("checked_out_at DESC, id ASC").
				Offset(p.Offset()).
				Limit(p.PageSize).
				Scan(ctx); err != nil {
				return err
			}
			page.Items = lo.Map(rows, func(m checkoutModel, _ int) dto.CheckoutResponse {
				return m.toResponse()
			})
			return nil
		},
	); err != nil {
		return nil, err
	}
	return page, nil
}
