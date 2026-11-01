package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-ddd-book-manager/internal/application"
	"go-ddd-book-manager/internal/config"
	"go-ddd-book-manager/internal/infra"
	"go-ddd-book-manager/internal/webapi"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	config, err := config.Load()
	if err != nil {
		return err
	}

	i := do.New()
	do.ProvideValue(i, config)
	infra.Inject(i)
	application.Inject(i)
	webapi.Inject(i)
	defer i.Shutdown()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	e := do.MustInvoke[*echo.Echo](i)
	server := &http.Server{Addr: config.HTTPAddr, Handler: e}
	return serve(ctx, server, server.ListenAndServe)
}

func serve(ctx context.Context, server *http.Server, listenAndServe func() error) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := listenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
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
		return server.Shutdown(shutdownCtx)
	})

	return g.Wait()
}
