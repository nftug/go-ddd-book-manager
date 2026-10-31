package book

import (
	"context"
	"net/http"

	bc "go-ddd-book-manager/internal/application/book/command"
	bd "go-ddd-book-manager/internal/application/book/dto"
	bq "go-ddd-book-manager/internal/application/book/query"
	ac "go-ddd-book-manager/internal/application/common"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi/bind"
	"go-ddd-book-manager/internal/webapi/handler"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
)

type Handler struct {
	listBooks         bq.GetBookList
	bookDetails       bq.GetBookDetails
	createBook        bc.CreateBook
	updateBook        bc.UpdateBook
	deleteBook        bc.DeleteBook
	changeBookOwner   bc.ChangeOwner
	checkoutBook      bc.CheckoutBook
	returnBookUseCase bc.ReturnBook
	checkoutHistory   bq.GetCheckoutHistory
}

func NewHandler(i do.Injector) (*Handler, error) {
	return &Handler{
		listBooks:         do.MustInvoke[bq.GetBookList](i),
		bookDetails:       do.MustInvoke[bq.GetBookDetails](i),
		createBook:        do.MustInvoke[bc.CreateBook](i),
		updateBook:        do.MustInvoke[bc.UpdateBook](i),
		deleteBook:        do.MustInvoke[bc.DeleteBook](i),
		changeBookOwner:   do.MustInvoke[bc.ChangeOwner](i),
		checkoutBook:      do.MustInvoke[bc.CheckoutBook](i),
		returnBookUseCase: do.MustInvoke[bc.ReturnBook](i),
		checkoutHistory:   do.MustInvoke[bq.GetCheckoutHistory](i),
	}, nil
}

func (h *Handler) list(c echo.Context) error {
	return handler.JSON(
		c,
		http.StatusOK,
		func(ctx context.Context, actor dc.Actor) (bd.BookPage, error) {
			q, err := bind.QueryParams[bd.BookQuery](c)
			if err != nil {
				return bd.BookPage{}, err
			}
			return h.listBooks.Execute(ctx, actor, q)
		},
	)
}

func (h *Handler) create(c echo.Context) error {
	return handler.JSON(
		c,
		http.StatusCreated,
		func(ctx context.Context, actor dc.Actor) (ac.CreatedResponse, error) {
			req, err := bind.Body[bd.BookRequest](c)
			if err != nil {
				return ac.CreatedResponse{}, err
			}
			return h.createBook.Execute(ctx, actor, req)
		},
	)
}

func (h *Handler) details(c echo.Context) error {
	return handler.JSON(
		c,
		http.StatusOK,
		func(ctx context.Context, actor dc.Actor) (bd.BookResponse, error) {
			id, err := bind.PathParams[bd.BookIdentity](c)
			if err != nil {
				return bd.BookResponse{}, err
			}
			return h.bookDetails.Execute(ctx, actor, id)
		},
	)
}

func (h *Handler) update(c echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		req, err := bind.Body[bd.BookRequest](c)
		if err != nil {
			return err
		}
		return h.updateBook.Execute(ctx, actor, id, req)
	})
}

func (h *Handler) delete(c echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		return h.deleteBook.Execute(ctx, actor, id)
	})
}

func (h *Handler) changeOwner(c echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		req, err := bind.Body[bd.OwnerChangeRequest](c)
		if err != nil {
			return err
		}
		return h.changeBookOwner.Execute(ctx, actor, id, req)
	})
}

func (h *Handler) checkout(c echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		return h.checkoutBook.Execute(ctx, actor, id)
	})
}

func (h *Handler) returnBook(c echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		return h.returnBookUseCase.Execute(ctx, actor, id)
	})
}

func (h *Handler) history(c echo.Context) error {
	return handler.JSON(
		c,
		http.StatusOK,
		func(ctx context.Context, actor dc.Actor) (bd.CheckoutPage, error) {
			id, err := bind.PathParams[bd.BookIdentity](c)
			if err != nil {
				return bd.CheckoutPage{}, err
			}
			p, err := bind.QueryParams[ac.Pagination](c)
			if err != nil {
				return bd.CheckoutPage{}, err
			}
			return h.checkoutHistory.Execute(ctx, actor, id, p)
		},
	)
}
