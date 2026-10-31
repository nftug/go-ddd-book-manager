package application_test

import (
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"go-ddd-book-manager/internal/application"
	bc "go-ddd-book-manager/internal/application/book/command"
	bq "go-ddd-book-manager/internal/application/book/query"
	uc "go-ddd-book-manager/internal/application/user/command"
	uq "go-ddd-book-manager/internal/application/user/query"
	"go-ddd-book-manager/internal/infra"
	"testing"
)

func TestInjection(t *testing.T) {
	i := application.Inject(
		infra.Inject(do.New(), "postgres://unused:unused@localhost/unused?sslmode=disable"),
	)
	t.Cleanup(func() { i.Shutdown() })
	require.NotNil(t, do.MustInvoke[*bc.CreateBook](i))
	require.NotNil(t, do.MustInvoke[*bc.ChangeOwner](i))
	require.NotNil(t, do.MustInvoke[*bq.GetBookList](i))
	require.NotNil(t, do.MustInvoke[*uc.SyncUser](i))
	require.NotNil(t, do.MustInvoke[*uq.GetUserDetails](i))
}
