package command_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"go-ddd-book-manager/internal/application/book/command"
	dto "go-ddd-book-manager/internal/application/book/dto"
	"go-ddd-book-manager/internal/common/option"
	db "go-ddd-book-manager/internal/domain/book"
	dc "go-ddd-book-manager/internal/domain/common"
	du "go-ddd-book-manager/internal/domain/user"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type bookRepositoryMock struct{ mock.Mock }

func (m *bookRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*db.Book, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Book), args.Error(1)
}

func (m *bookRepositoryMock) Create(ctx context.Context, book db.Book) error {
	return m.Called(ctx, book).Error(0)
}

func (m *bookRepositoryMock) Update(ctx context.Context, book db.Book) error {
	return m.Called(ctx, book).Error(0)
}

func (m *bookRepositoryMock) Delete(ctx context.Context, book db.Book) error {
	return m.Called(ctx, book).Error(0)
}

type userRepositoryMock struct{ mock.Mock }

func (m *userRepositoryMock) FindByID(ctx context.Context, id uuid.UUID) (*du.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*du.User), args.Error(1)
}

func (m *userRepositoryMock) Create(ctx context.Context, user du.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *userRepositoryMock) Update(ctx context.Context, user du.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *userRepositoryMock) Delete(ctx context.Context, user du.User) error {
	return m.Called(ctx, user).Error(0)
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

type fixedIDGenerator struct{ id uuid.UUID }

func (g fixedIDGenerator) NewUUIDv7() uuid.UUID { return g.id }

func existingBook(id, ownerID uuid.UUID) db.Book {
	audit := dc.RestoreEntityAudit(
		id,
		1,
		dc.RestoreAuditRecord(ownerID, "owner", fixedClock{}.Now()),
		option.None[dc.AuditRecord](),
	)
	return db.Restore(
		audit,
		"Book",
		nil,
		option.None[string](),
		option.None[string](),
		du.RestoreReference(ownerID, "owner"),
		nil,
	)
}

func bookPtr(book db.Book) *db.Book { return &book }

func TestCreateBook(t *testing.T) {
	ownerID, bookID := uuid.NewV7(), uuid.NewV7()
	actor := dc.NewActor(ownerID, "owner", dc.UserRole)
	request := dto.BookRequest{Title: "Book", Authors: []string{"Author"}}

	t.Run("invalid input is returned without creating a book", func(t *testing.T) {
		repo := &bookRepositoryMock{}
		_, err := newBookCommandService(
			t,
			repo,
			&userRepositoryMock{},
			fixedIDGenerator{bookID},
			command.NewCreateBook,
		).
			Execute(context.Background(), actor, dto.BookRequest{})
		var validation *dc.ValidationError
		require.ErrorAs(t, err, &validation)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
	t.Run("the actor owns the created book and its ID is returned", func(t *testing.T) {
		repo := &bookRepositoryMock{}
		repo.On("Create", mock.Anything, mock.MatchedBy(func(b db.Book) bool { return b.Owner().ID() == ownerID })).
			Return(nil).
			Once()
		id, err := newBookCommandService(
			t,
			repo,
			&userRepositoryMock{},
			fixedIDGenerator{bookID},
			command.NewCreateBook,
		).
			Execute(context.Background(), actor, request)
		require.NoError(t, err)
		require.Equal(t, bookID, id)
		repo.AssertExpectations(t)
	})
}

func TestDeleteBookReturnsNotFoundWhenAbsent(t *testing.T) {
	repo := &bookRepositoryMock{}
	actor := dc.NewActor(uuid.NewV7(), "owner", dc.UserRole)
	repo.On("FindByID", mock.Anything, mock.Anything).Return(nil, nil).Once()
	err := newBookCommandService(
		t,
		repo,
		&userRepositoryMock{},
		fixedIDGenerator{uuid.NewV7()},
		command.NewDeleteBook,
	).
		Execute(context.Background(), actor, dto.BookIdentity{BookID: uuid.NewV7()})
	require.ErrorIs(t, err, dc.ErrNotFound)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestUpdateBookDoesNotSaveWhenActorIsNotOwner(t *testing.T) {
	id, ownerID := uuid.NewV7(), uuid.NewV7()
	repo := &bookRepositoryMock{}
	repo.On("FindByID", mock.Anything, id).
		Return(bookPtr(existingBook(id, ownerID)), nil).
		Once()
	actor := dc.NewActor(uuid.NewV7(), "other", dc.UserRole)
	request := dto.BookRequest{Title: "Book", Authors: []string{"Author"}}
	err := newBookCommandService(
		t,
		repo,
		&userRepositoryMock{},
		fixedIDGenerator{uuid.NewV7()},
		command.NewUpdateBook,
	).
		Execute(context.Background(), actor, dto.BookIdentity{BookID: id}, request)
	require.ErrorIs(t, err, dc.ErrForbidden)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestChangeOwnerRejectsMissingUserWithoutSaving(t *testing.T) {
	id, ownerID := uuid.NewV7(), uuid.NewV7()
	repo := &bookRepositoryMock{}
	repo.On("FindByID", mock.Anything, id).
		Return(bookPtr(existingBook(id, ownerID)), nil).
		Once()
	admin := dc.NewActor(uuid.NewV7(), "admin", dc.AdminRole)
	userRepo := &userRepositoryMock{}
	userRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, nil).Once()
	err := newBookCommandService(
		t,
		repo,
		userRepo,
		fixedIDGenerator{uuid.NewV7()},
		command.NewChangeOwner,
	).
		Execute(context.Background(), admin, dto.BookIdentity{BookID: id}, uuid.NewV7())
	var validation *dc.ValidationError
	require.ErrorAs(t, err, &validation)
	require.Equal(t, "book.owner_id", validation.Field())
	require.Equal(t, "not_found", validation.Code())
	userRepo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestCheckoutBookPreservesSaveError(t *testing.T) {
	id, ownerID := uuid.NewV7(), uuid.NewV7()
	failure := errors.New("save failed")
	repo := &bookRepositoryMock{}
	repo.On("FindByID", mock.Anything, id).
		Return(bookPtr(existingBook(id, ownerID)), nil).
		Once()
	repo.On("Update", mock.Anything, mock.Anything).Return(failure).Once()
	actor := dc.NewActor(ownerID, "owner", dc.UserRole)
	err := newBookCommandService(
		t,
		repo,
		&userRepositoryMock{},
		fixedIDGenerator{uuid.NewV7()},
		command.NewCheckoutBook,
	).
		Execute(context.Background(), actor, dto.BookIdentity{BookID: id})
	require.ErrorIs(t, err, failure)
	repo.AssertExpectations(t)
}

func newBookCommandService[T any](
	t *testing.T,
	books db.Repository,
	users du.Repository,
	ids dc.UUIDGenerator,
	constructor func(do.Injector) (*T, error),
) *T {
	t.Helper()
	injector := do.New()
	do.ProvideValue[db.Repository](injector, books)
	do.ProvideValue[du.Repository](injector, users)
	do.ProvideValue[dc.Clock](injector, fixedClock{})
	do.ProvideValue[dc.UUIDGenerator](injector, ids)
	t.Cleanup(func() { injector.Shutdown() })
	service, err := constructor(injector)
	require.NoError(t, err)
	return service
}
