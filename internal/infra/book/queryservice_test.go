//go:build integration

package book_test

import (
	"context"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	dto "go-ddd-book-manager/internal/application/book/dto"
	q "go-ddd-book-manager/internal/application/book/query"
	ac "go-ddd-book-manager/internal/application/common"
	"go-ddd-book-manager/internal/common/option"
	"go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/domain/user"
	"go-ddd-book-manager/internal/infra/testutil"
	"testing"
	"time"
	"uuid"
)

func TestQueryService(t *testing.T) {
	ctx := context.Background()
	i, _, _ := testutil.NewPostgres(t)
	repo := do.MustInvoke[book.Repository](i)
	service := do.MustInvoke[q.QueryService](i)
	owner := uuid.NewV7()
	actor := dc.NewActor(owner, "Owner", dc.UserRole)
	ref := user.RestoreReference(owner, "Owner")
	stamp := time.Now().UTC().Truncate(time.Microsecond)
	create := func(title, author string, checked bool) uuid.UUID {
		id := uuid.NewV7()
		audit := dc.RestoreEntityAudit(
			id,
			1,
			dc.RestoreAuditRecord(owner, "Owner", stamp),
			option.None[dc.AuditRecord](),
		)
		var checkouts []book.Checkout
		if checked {
			checkouts = []book.Checkout{
				book.RestoreCheckout(uuid.NewV7(), ref, stamp, option.None[time.Time]()),
			}
		}
		b := book.Restore(
			audit,
			title,
			[]book.Author{book.RestoreAuthor(author, 0)},
			option.None[string](),
			option.None[string](),
			ref,
			checkouts,
		)
		require.NoError(t, repo.Create(ctx, b))
		return id
	}
	id := create("Domain Design", "Author", true)
	emptyBookID := create("Domain Other", "Different", false)
	create("Other", "Author", true)
	title, author, checked := "domain", "author", true
	page, err := service.List(
		ctx,
		actor,
		dto.BookQuery{
			Title:      &title,
			Author:     &author,
			CheckedOut: &checked,
			Pagination: ac.Pagination{Page: 1, PageSize: 10},
		},
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, int64(1), page.Page)
	require.Equal(t, int64(10), page.PageSize)
	require.Len(t, page.Items, 1)
	require.Equal(t, id, page.Items[0].ID)
	require.Equal(t, []string{"Author"}, page.Items[0].Authors)
	require.True(t, page.Items[0].CheckedOut)
	require.True(t, page.Items[0].Permission.CanUpdate)
	detail, err := service.FindByID(ctx, actor, id)
	require.NoError(t, err)
	require.Equal(t, id, detail.ID)
	require.Equal(t, owner, detail.Owner.ID)
	require.True(t, detail.Permission.CanDelete)
	other := dc.NewActor(uuid.NewV7(), "Other", dc.UserRole)
	otherDetail, err := service.FindByID(ctx, other, id)
	require.NoError(t, err)
	require.False(t, otherDetail.Permission.CanDelete)
	pagination := ac.Pagination{Page: 1, PageSize: 10}
	history, err := service.Checkouts(ctx, id, pagination)
	require.NoError(t, err)
	require.Equal(t, int64(1), history.Total)
	require.Equal(t, pagination.Page, history.Page)
	require.Equal(t, pagination.PageSize, history.PageSize)
	require.Len(t, history.Items, 1)
	empty, err := service.Checkouts(ctx, uuid.NewV7(), pagination)
	require.NoError(t, err)
	require.Nil(t, empty)
	noHistory, err := service.Checkouts(ctx, emptyBookID, pagination)
	require.NoError(t, err)
	require.NotNil(t, noHistory)
	require.Equal(t, int64(0), noHistory.Total)
	require.Equal(t, pagination.Page, noHistory.Page)
	require.Equal(t, pagination.PageSize, noHistory.PageSize)
	require.NotNil(t, noHistory.Items)
	require.Empty(t, noHistory.Items)
	laterHistory, err := service.Checkouts(ctx, id, ac.Pagination{Page: 2, PageSize: 1})
	require.NoError(t, err)
	require.Equal(t, int64(1), laterHistory.Total)
	require.Equal(t, int64(2), laterHistory.Page)
	require.Empty(t, laterHistory.Items)
	firstPage, err := service.List(
		ctx,
		actor,
		dto.BookQuery{Pagination: ac.Pagination{Page: 1, PageSize: 1}},
	)
	require.NoError(t, err)
	require.Len(t, firstPage.Items, 1)
	page, err = service.List(
		ctx,
		actor,
		dto.BookQuery{Pagination: ac.Pagination{Page: 2, PageSize: 1}},
	)
	require.NoError(t, err)
	require.Equal(t, int64(3), page.Total)
	require.Equal(t, int64(2), page.Page)
	require.Equal(t, int64(1), page.PageSize)
	require.Len(t, page.Items, 1)
	require.NotEqual(t, firstPage.Items[0].ID, page.Items[0].ID)
}
