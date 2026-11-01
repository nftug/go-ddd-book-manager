package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-ddd-book-manager/internal/application"
	"go-ddd-book-manager/internal/infra"
	"go-ddd-book-manager/internal/webapi"
	"go-ddd-book-manager/internal/webapi/auth"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	issuer := os.Getenv("OIDC_ISSUER")
	audience := os.Getenv("OIDC_AUDIENCE")
	if dsn == "" || issuer == "" || audience == "" {
		return errors.New("DATABASE_URL, OIDC_ISSUER and OIDC_AUDIENCE are required")
	}
	addr, _ := lo.Coalesce(os.Getenv("HTTP_ADDR"), ":8080")

	i := do.New()
	infra.Inject(i, dsn)
	application.Inject(i)
	webapi.Inject(i, auth.Config{Issuer: issuer, Audience: audience})
	defer i.Shutdown()

	e, err := do.Invoke[*echo.Echo](i)
	if err != nil {
		return fmt.Errorf("initialize web API: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := e.Start(addr); errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()
		return e.Shutdown(shutdownCtx)
	})

	return g.Wait()
}
