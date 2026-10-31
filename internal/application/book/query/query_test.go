package query_test

import (
	"context"
	"errors"
	dto "go-ddd-book-manager/internal/application/book/dto"
	q "go-ddd-book-manager/internal/application/book/query"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
	"math"
	"testing"
	"uuid"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type bookQueryMock struct{ mock.Mock }

func (m *bookQueryMock) FindByID(
	ctx context.Context,
	actor dc.Actor,
	id uuid.UUID,
) (*dto.BookResponse, error) {
	args := m.Called(ctx, actor, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.BookResponse), args.Error(1)
}

func (m *bookQueryMock) List(
	ctx context.Context,
	actor dc.Actor,
	query dto.BookQuery,
) (dto.BookPage, error) {
	args := m.Called(ctx, actor, query)
	return args.Get(0).(dto.BookPage), args.Error(1)
}

func (m *bookQueryMock) Checkouts(
	ctx context.Context,
	id uuid.UUID,
	p ac.Pagination,
) (*dto.CheckoutPage, error) {
	args := m.Called(ctx, id, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CheckoutPage), args.Error(1)
}

func TestGetBookList(t *testing.T) {
	ownerID := uuid.NewV7()
	actor := dc.NewActor(ownerID, "owner", dc.UserRole)
	t.Run("default pagination and permissions from query", func(t *testing.T) {
		service := &bookQueryMock{}
		service.On("List", mock.Anything, actor, dto.BookQuery{Pagination: ac.Pagination{Page: 1, PageSize: 10}}).
			Return(dto.BookPage{Items: []dto.BookResponse{{Owner: ac.UserReference{ID: ownerID}, Permission: ac.PermissionResponse{CanUpdate: true}}}, Total: 1, Page: 1, PageSize: 10}, nil).
			Once()
		page, err := newBookQueryService(
			t,
			service,
			q.NewGetBookList,
		).Execute(context.Background(), actor, dto.BookQuery{})
		require.NoError(t, err)
		require.Equal(t, int64(1), page.Page)
		require.Equal(t, int64(10), page.PageSize)
		require.True(t, page.Items[0].Permission.CanUpdate)
		service.AssertExpectations(t)
	})
	t.Run("overflow is rejected before querying", func(t *testing.T) {
		service := &bookQueryMock{}
		_, err := newBookQueryService(t, service, q.NewGetBookList).
			Execute(context.Background(), actor, dto.BookQuery{Pagination: ac.Pagination{Page: math.MaxInt64, PageSize: 100}})
		var validation *dc.ValidationError
		require.ErrorAs(t, err, &validation)
		service.AssertNotCalled(
			t,
			"List",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
	})
	t.Run("query errors are preserved", func(t *testing.T) {
		failure := errors.New("query failed")
		service := &bookQueryMock{}
		service.On("List", mock.Anything, actor, mock.Anything).
			Return(dto.BookPage{}, failure).
			Once()
		_, err := newBookQueryService(
			t,
			service,
			q.NewGetBookList,
		).Execute(context.Background(), actor, dto.BookQuery{})
		require.ErrorIs(t, err, failure)
		service.AssertExpectations(t)
	})
}

func TestGetBookDetailsPassesActorAndReturnsQueryPermission(t *testing.T) {
	service := &bookQueryMock{}
	actor := dc.NewActor(uuid.NewV7(), "other", dc.UserRole)
	service.On("FindByID", mock.Anything, actor, mock.Anything).
		Return(&dto.BookResponse{Owner: ac.UserReference{ID: uuid.NewV7()}, Permission: ac.PermissionResponse{CanDelete: true}}, nil).
		Once()
	detail, err := newBookQueryService(t, service, q.NewGetBookDetails).
		Execute(context.Background(), actor, dto.BookIdentity{})
	require.NoError(t, err)
	require.True(t, detail.Permission.CanDelete)
	service.AssertExpectations(t)
}

func TestGetCheckoutHistory(t *testing.T) {
	t.Run("non-admin is rejected before querying", func(t *testing.T) {
		service := &bookQueryMock{}
		actor := dc.NewActor(uuid.NewV7(), "user", dc.UserRole)
		_, err := newBookQueryService(t, service, q.NewGetCheckoutHistory).
			Execute(context.Background(), actor, dto.BookIdentity{}, ac.Pagination{})
		require.ErrorIs(t, err, dc.ErrForbidden)
		service.AssertNotCalled(
			t,
			"Checkouts",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
	})
	t.Run("missing book returns not found", func(t *testing.T) {
		service := &bookQueryMock{}
		actor := dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole)
		id := uuid.NewV7()
		service.On("Checkouts", mock.Anything, id, ac.Pagination{Page: 1, PageSize: 10}).
			Return(nil, nil).
			Once()
		_, err := newBookQueryService(t, service, q.NewGetCheckoutHistory).
			Execute(context.Background(), actor, dto.BookIdentity{BookID: id}, ac.Pagination{})
		require.ErrorIs(t, err, dc.ErrNotFound)
		service.AssertExpectations(t)
	})
	t.Run("admin receives an empty page for a book without history", func(t *testing.T) {
		service := &bookQueryMock{}
		service.On("Checkouts", mock.Anything, mock.Anything, ac.Pagination{Page: 1, PageSize: 10}).
			Return(&dto.CheckoutPage{Items: []dto.CheckoutResponse{}, Page: 1, PageSize: 10}, nil).
			Once()
		actor := dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole)
		page, err := newBookQueryService(t, service, q.NewGetCheckoutHistory).
			Execute(context.Background(), actor, dto.BookIdentity{}, ac.Pagination{})
		require.NoError(t, err)
		require.Equal(
			t,
			dto.CheckoutPage{Items: []dto.CheckoutResponse{}, Page: 1, PageSize: 10},
			page,
		)
		service.AssertExpectations(t)
	})
	t.Run("admin receives the requested page", func(t *testing.T) {
		service := &bookQueryMock{}
		pagination := ac.Pagination{Page: 2, PageSize: 3}
		want := dto.CheckoutPage{
			Items:    []dto.CheckoutResponse{{ID: uuid.NewV7()}},
			Total:    4,
			Page:     2,
			PageSize: 3,
		}
		service.On("Checkouts", mock.Anything, mock.Anything, pagination).
			Return(&want, nil).
			Once()
		actor := dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole)
		page, err := newBookQueryService(t, service, q.NewGetCheckoutHistory).
			Execute(context.Background(), actor, dto.BookIdentity{}, pagination)
		require.NoError(t, err)
		require.Equal(t, want, page)
		service.AssertExpectations(t)
	})
	t.Run("checkout query error is preserved", func(t *testing.T) {
		service := &bookQueryMock{}
		failure := errors.New("checkout query failed")
		service.On("Checkouts", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, failure).
			Once()
		actor := dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole)
		_, err := newBookQueryService(t, service, q.NewGetCheckoutHistory).
			Execute(context.Background(), actor, dto.BookIdentity{}, ac.Pagination{})
		require.ErrorIs(t, err, failure)
		service.AssertExpectations(t)
	})
	t.Run("overflow is rejected before querying", func(t *testing.T) {
		service := &bookQueryMock{}
		actor := dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole)
		_, err := newBookQueryService(t, service, q.NewGetCheckoutHistory).
			Execute(context.Background(), actor, dto.BookIdentity{}, ac.Pagination{Page: math.MaxInt64, PageSize: 100})
		var validation *dc.ValidationError
		require.ErrorAs(t, err, &validation)
		service.AssertNotCalled(t, "Checkouts", mock.Anything, mock.Anything, mock.Anything)
	})
}

func newBookQueryService[T any](
	t *testing.T,
	query q.QueryService,
	constructor func(do.Injector) (T, error),
) T {
	t.Helper()
	injector := do.New()
	do.ProvideValue[q.QueryService](injector, query)
	t.Cleanup(func() { injector.Shutdown() })
	service, err := constructor(injector)
	require.NoError(t, err)
	return service
}
