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

	"github.com/labstack/echo/v5"
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

// list godoc
// @ID ListBooks
// @Summary List books
// @Tags Books
// @Produce json
// @Param title query string false "Title"
// @Param author query string false "Author"
// @Param checked_out query boolean false "Checked-out status"
// @Param page query integer false "Page number"
// @Param page_size query integer false "Page size"
// @Success 200 {object} dto.BookPage
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books [get]
func (h *Handler) list(c *echo.Context) error {
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

// create godoc
// @ID CreateBook
// @Summary Create a book
// @Tags Books
// @Produce json
// @Accept json
// @Param request body dto.BookRequest true "Book"
// @Success 201 {object} common.CreatedResponse
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Failure 415 {object} handler.ErrorResponse "Unsupported Media Type"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books [post]
func (h *Handler) create(c *echo.Context) error {
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

// details godoc
// @ID GetBook
// @Summary Get book details
// @Tags Books
// @Produce json
// @Param book_id path string true "Book ID" format(uuid)
// @Success 200 {object} dto.BookResponse
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id} [get]
func (h *Handler) details(c *echo.Context) error {
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

// update godoc
// @ID UpdateBook
// @Summary Update a book
// @Tags Books
// @Produce json
// @Accept json
// @Param book_id path string true "Book ID" format(uuid)
// @Param request body dto.BookRequest true "Book"
// @Success 204 "No Content"
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} handler.ErrorResponse "Forbidden"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Failure 415 {object} handler.ErrorResponse "Unsupported Media Type"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id} [put]
func (h *Handler) update(c *echo.Context) error {
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

// delete godoc
// @ID DeleteBook
// @Summary Delete a book
// @Tags Books
// @Produce json
// @Param book_id path string true "Book ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} handler.ErrorResponse "Forbidden"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id} [delete]
func (h *Handler) delete(c *echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		return h.deleteBook.Execute(ctx, actor, id)
	})
}

// changeOwner godoc
// @ID ChangeBookOwner
// @Summary Change the owner of a book
// @Tags Books
// @Produce json
// @Accept json
// @Param book_id path string true "Book ID" format(uuid)
// @Param request body dto.OwnerChangeRequest true "New owner"
// @Success 204 "No Content"
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} handler.ErrorResponse "Forbidden"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Failure 415 {object} handler.ErrorResponse "Unsupported Media Type"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id}/owner [put]
func (h *Handler) changeOwner(c *echo.Context) error {
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

// checkout godoc
// @ID CheckoutBook
// @Summary Check out a book
// @Tags Books
// @Produce json
// @Param book_id path string true "Book ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id}/checkouts [post]
func (h *Handler) checkout(c *echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		return h.checkoutBook.Execute(ctx, actor, id)
	})
}

// returnBook godoc
// @ID ReturnBook
// @Summary Return a book
// @Tags Books
// @Produce json
// @Param book_id path string true "Book ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} handler.ErrorResponse "Forbidden"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 409 {object} handler.ErrorResponse "Conflict"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id}/return [post]
func (h *Handler) returnBook(c *echo.Context) error {
	return handler.NoContent(c, func(ctx context.Context, actor dc.Actor) error {
		id, err := bind.PathParams[bd.BookIdentity](c)
		if err != nil {
			return err
		}
		return h.returnBookUseCase.Execute(ctx, actor, id)
	})
}

// history godoc
// @ID GetBookCheckoutHistory
// @Summary Get checkout history
// @Tags Books
// @Produce json
// @Param book_id path string true "Book ID" format(uuid)
// @Param page query integer false "Page number"
// @Param page_size query integer false "Page size"
// @Success 200 {object} dto.CheckoutPage
// @Failure 400 {object} handler.ErrorResponse "Bad Request"
// @Failure 401 {object} handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} handler.ErrorResponse "Forbidden"
// @Failure 404 {object} handler.ErrorResponse "Not Found"
// @Failure 500 {object} handler.ErrorResponse "Internal Server Error"
// @Header 401 {string} WWW-Authenticate "Bearer"
// @Security BearerAuth
// @Router /books/{book_id}/checkouts [get]
func (h *Handler) history(c *echo.Context) error {
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
