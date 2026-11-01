package webapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	bc "go-ddd-book-manager/internal/application/book/command"
	bd "go-ddd-book-manager/internal/application/book/dto"
	bq "go-ddd-book-manager/internal/application/book/query"
	ac "go-ddd-book-manager/internal/application/common"
	uc "go-ddd-book-manager/internal/application/user/command"
	ud "go-ddd-book-manager/internal/application/user/dto"
	uq "go-ddd-book-manager/internal/application/user/query"
	us "go-ddd-book-manager/internal/application/user/service"
	"go-ddd-book-manager/internal/config"
	dc "go-ddd-book-manager/internal/domain/common"
	"go-ddd-book-manager/internal/webapi"
	"go-ddd-book-manager/internal/webapi/auth"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
)

type testDependencies struct {
	AllowedOrigins  []string
	Verifier        auth.TokenVerifier
	Resolver        us.ResolveActor
	SyncUser        uc.SyncUser
	UserDetails     uq.GetUserDetails
	BookList        bq.GetBookList
	BookDetails     bq.GetBookDetails
	CreateBook      bc.CreateBook
	UpdateBook      bc.UpdateBook
	DeleteBook      bc.DeleteBook
	ChangeOwner     bc.ChangeOwner
	CheckoutBook    bc.CheckoutBook
	ReturnBook      bc.ReturnBook
	CheckoutHistory bq.GetCheckoutHistory
}

func newTestServer(t *testing.T, d testDependencies) *echo.Echo {
	t.Helper()
	if d.SyncUser == nil {
		d.SyncUser = syncer{}
	}

	if d.UserDetails == nil {
		d.UserDetails = userDetails{}
	}

	if d.BookList == nil {
		d.BookList = list{}
	}

	if d.BookDetails == nil {
		d.BookDetails = details{}
	}

	if d.CreateBook == nil {
		d.CreateBook = creator{}
	}

	if d.UpdateBook == nil {
		d.UpdateBook = updater{}
	}

	if d.DeleteBook == nil {
		d.DeleteBook = remover{}
	}

	if d.ChangeOwner == nil {
		d.ChangeOwner = ownerChanger{}
	}

	if d.CheckoutBook == nil {
		d.CheckoutBook = remover{}
	}

	if d.ReturnBook == nil {
		d.ReturnBook = remover{}
	}

	if d.CheckoutHistory == nil {
		d.CheckoutHistory = history{}
	}

	i := do.New()
	do.ProvideValue(i, config.Config{AllowedOrigins: d.AllowedOrigins})
	webapi.Inject(i)
	t.Cleanup(func() { i.Shutdown() })
	do.OverrideValue[auth.TokenVerifier](i, d.Verifier)
	do.ProvideValue[us.ResolveActor](i, d.Resolver)
	do.ProvideValue[uc.SyncUser](i, d.SyncUser)
	do.ProvideValue[uq.GetUserDetails](i, d.UserDetails)
	do.ProvideValue[bq.GetBookList](i, d.BookList)
	do.ProvideValue[bq.GetBookDetails](i, d.BookDetails)
	do.ProvideValue[bc.CreateBook](i, d.CreateBook)
	do.ProvideValue[bc.UpdateBook](i, d.UpdateBook)
	do.ProvideValue[bc.DeleteBook](i, d.DeleteBook)
	do.ProvideValue[bc.ChangeOwner](i, d.ChangeOwner)
	do.ProvideValue[bc.CheckoutBook](i, d.CheckoutBook)
	do.ProvideValue[bc.ReturnBook](i, d.ReturnBook)
	do.ProvideValue[bq.GetCheckoutHistory](i, d.CheckoutHistory)
	e, err := do.Invoke[*echo.Echo](i)
	require.NoError(t, err)
	return e
}

var userID = uuid.MustParse("00000000-0000-4000-8000-000000000001")

type verifier struct{}

func (verifier) Verify(_ context.Context, s string) (auth.Identity, error) {
	if s != "valid" {
		return auth.Identity{}, auth.ErrUnauthorized
	}

	return auth.Identity{
		ID:    userID,
		Name:  "Test",
		Email: "test@example.com",
		Role:  dc.UserRole,
	}, nil
}

type resolver struct{ err error }

func (r resolver) Execute(context.Context, uuid.UUID) (dc.Actor, error) {
	return dc.NewActor(userID, "Test", dc.UserRole), r.err
}

type syncer struct{ called *bool }

func (s syncer) Execute(_ context.Context, r ud.SyncUserRequest) (ud.UserResponse, error) {
	*s.called = true
	return ud.UserResponse{ID: r.ID, Name: r.Name}, nil
}

type list struct {
	query *bd.BookQuery
	err   error
}

func (s list) Execute(_ context.Context, _ dc.Actor, q bd.BookQuery) (bd.BookPage, error) {
	*s.query = q
	return bd.BookPage{}, s.err
}

type remover struct{ id *bd.BookIdentity }

func (s remover) Execute(_ context.Context, _ dc.Actor, id bd.BookIdentity) error {
	*s.id = id
	return nil
}

func request(e http.Handler, method, path, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	return w
}

func TestAuthenticationAndBinding(t *testing.T) {
	called := false
	var q bd.BookQuery
	var id bd.BookIdentity
	e := newTestServer(t,
		testDependencies{
			Verifier:   verifier{},
			Resolver:   resolver{},
			SyncUser:   syncer{&called},
			BookList:   list{query: &q},
			DeleteBook: remover{&id},
		},
	)
	w := request(e, "GET", "/books", " ")
	require.Equal(t, 401, w.Code)
	require.Equal(t, "Bearer", w.Header().Get("WWW-Authenticate"))
	w = request(e, "POST", "/users/me", "valid")
	require.Equal(t, 200, w.Code)
	require.True(t, called)
	w = request(e, "GET", "/books?title=go&checked_out=true&page=2&page_size=15", "valid")
	require.Equal(t, 200, w.Code)
	require.Equal(t, "go", *q.Title)
	require.True(t, *q.CheckedOut)
	require.EqualValues(t, 2, q.Pagination.Page)
	require.EqualValues(t, 15, q.Pagination.PageSize)
	w = request(e, "GET", "/books?page=abc", "valid")
	require.Equal(t, 400, w.Code)
	w = request(e, "DELETE", "/books/invalid", "valid")
	require.Equal(t, 400, w.Code)
	w = request(e, "DELETE", "/books/"+userID.String(), "valid")
	require.Equal(t, 204, w.Code)
	require.Empty(t, w.Body.String())
	require.Equal(t, userID, id.BookID)
}

func TestAuthErrorsAndInternalDetails(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{dc.ErrNotFound, 401}, {errors.New("sensitive DB failure"), 500}} {
		called := false
		e := newTestServer(t,
			testDependencies{
				Verifier: verifier{},
				Resolver: resolver{tc.err},
				SyncUser: syncer{&called},
			},
		)
		w := request(e, "GET", "/users/me", "valid")
		require.Equal(t, tc.status, w.Code)
		require.False(t, called)
		require.NotContains(t, w.Body.String(), "sensitive")
		if tc.status == 401 {
			require.Equal(t, "Bearer", w.Header().Get("WWW-Authenticate"))
		}

		w = request(e, "POST", "/users/me", "valid")
		require.Equal(t, 200, w.Code)
		require.True(t, called)
	}
}

func TestRoutesRequireRegisteredUser(t *testing.T) {
	e := newTestServer(
		t,
		testDependencies{Verifier: verifier{}, Resolver: resolver{dc.ErrNotFound}},
	)
	base := "/books/" + userID.String()
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/users/me"},
		{http.MethodGet, "/books"},
		{http.MethodPost, "/books"},
		{http.MethodGet, base},
		{http.MethodPut, base},
		{http.MethodDelete, base},
		{http.MethodPut, base + "/owner"},
		{http.MethodPost, base + "/checkouts"},
		{http.MethodPost, base + "/return"},
		{http.MethodGet, base + "/checkouts"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			for _, token := range []string{"", "invalid", "valid"} {
				w := request(e, route.method, route.path, token)
				require.Equal(t, http.StatusUnauthorized, w.Code, "token: %q", token)
				require.Equal(t, "Bearer", w.Header().Get(echo.HeaderWWWAuthenticate))
			}
		})
	}
}

func TestErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{dc.ErrForbidden, 403}, {dc.ErrNotFound, 404}, {dc.ErrConflict, 409}, {dc.NewValidationError("title", "invalid"), 400}, {errors.New("secret"), 500}} {
		var q bd.BookQuery
		e := newTestServer(t,
			testDependencies{
				Verifier: verifier{},
				Resolver: resolver{},
				BookList: list{&q, tc.err},
			},
		)
		w := request(e, "GET", "/books", "valid")
		require.Equal(t, tc.status, w.Code)
		require.False(t, strings.Contains(w.Body.String(), "secret"))
	}
}

type userDetails struct{}

func (userDetails) Execute(_ context.Context, a dc.Actor) (ud.UserResponse, error) {
	return ud.UserResponse{ID: a.ID()}, nil
}

type details struct{}

func (details) Execute(_ context.Context, _ dc.Actor, id bd.BookIdentity) (bd.BookResponse, error) {
	return bd.BookResponse{ID: id.BookID}, nil
}

type creator struct{ request *bd.BookRequest }

func (s creator) Execute(
	_ context.Context,
	_ dc.Actor,
	r bd.BookRequest,
) (ac.CreatedResponse, error) {
	*s.request = r
	return ac.CreatedResponse{ID: userID}, nil
}

type updater struct {
	request *bd.BookRequest
	id      *bd.BookIdentity
}

func (s updater) Execute(
	_ context.Context,
	_ dc.Actor,
	id bd.BookIdentity,
	r bd.BookRequest,
) error {
	*s.request = r
	*s.id = id
	return nil
}

type ownerChanger struct{ request *bd.OwnerChangeRequest }

func (s ownerChanger) Execute(
	_ context.Context,
	_ dc.Actor,
	_ bd.BookIdentity,
	r bd.OwnerChangeRequest,
) error {
	*s.request = r
	return nil
}

type history struct{ pagination *ac.Pagination }

func (s history) Execute(
	_ context.Context,
	_ dc.Actor,
	_ bd.BookIdentity,
	p ac.Pagination,
) (bd.CheckoutPage, error) {
	*s.pagination = p
	return bd.CheckoutPage{}, nil
}

func TestFeatureRoutes(t *testing.T) {
	var req bd.BookRequest
	var id bd.BookIdentity
	var owner bd.OwnerChangeRequest
	var p ac.Pagination
	e := newTestServer(t,
		testDependencies{
			Verifier:        verifier{},
			Resolver:        resolver{},
			UserDetails:     userDetails{},
			BookDetails:     details{},
			CreateBook:      creator{&req},
			UpdateBook:      updater{&req, &id},
			DeleteBook:      remover{&id},
			ChangeOwner:     ownerChanger{&owner},
			CheckoutBook:    remover{&id},
			ReturnBook:      remover{&id},
			CheckoutHistory: history{&p},
		},
	)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer valid")
		if body != "" {
			r.Header.Set("Content-Type", "application/json")
		}

		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		return w
	}
	base := "/books/" + userID.String()
	require.Equal(t, 200, call("GET", "/users/me", "").Code)
	require.Equal(t, 200, call("GET", base, "").Code)
	w := call("POST", "/books", `{"title":"Go","authors":["A"]}`)
	require.Equal(t, 201, w.Code)
	require.Equal(t, "Go", req.Title)
	require.Contains(t, w.Body.String(), userID.String())
	w = call("PUT", base, `{"title":"Updated","authors":["B"],"book_id":"other"}`)
	require.Equal(t, 204, w.Code)
	require.Equal(t, "Updated", req.Title)
	require.Equal(t, userID, id.BookID)
	w = call("PUT", base+"/owner", `{"ownerId":"`+userID.String()+`"}`)
	require.Equal(t, 204, w.Code)
	require.Equal(t, userID, owner.OwnerID)
	for _, path := range []string{base + "/checkouts", base + "/return"} {
		w = call("POST", path, "")
		require.Equal(t, 204, w.Code)
		require.Empty(t, w.Body.String())
	}
	w = call("GET", base+"/checkouts?page=3&page_size=7", "")
	require.Equal(t, 200, w.Code)
	require.EqualValues(t, 3, p.Page)
	require.EqualValues(t, 7, p.PageSize)
	require.Equal(t, 400, call("POST", "/books", `{"title":`).Code)
	require.Equal(t, 400, call("PUT", base+"/owner", `{"ownerId":"invalid"}`).Code)
}

func TestCORS(t *testing.T) {
	var query bd.BookQuery
	e := newTestServer(t, testDependencies{
		Verifier:       verifier{},
		Resolver:       resolver{},
		AllowedOrigins: []string{"http://localhost:3000", "https://app.example.com"},
		BookList:       list{query: &query},
	})
	for _, tc := range []struct {
		name, origin string
		allowed      bool
	}{
		{"development", "http://localhost:3000", true},
		{"production", "https://app.example.com", true},
		{"untrusted", "https://untrusted.example.com", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/books", nil)
			req.Header.Set(echo.HeaderOrigin, tc.origin)
			req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
			req.Header.Set(echo.HeaderAccessControlRequestHeaders, "authorization,content-type")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
			require.Equal(t, http.StatusNoContent, w.Code)
			if tc.allowed {
				require.Equal(t, tc.origin, w.Header().Get(echo.HeaderAccessControlAllowOrigin))
				require.Contains(
					t,
					w.Header().Get(echo.HeaderAccessControlAllowMethods),
					http.MethodPost,
				)
				require.Equal(
					t,
					"Authorization,Content-Type",
					w.Header().Get(echo.HeaderAccessControlAllowHeaders),
				)
			} else {
				require.Empty(t, w.Header().Get(echo.HeaderAccessControlAllowOrigin))
			}
			require.Empty(t, w.Header().Get(echo.HeaderAccessControlAllowCredentials))
		})
	}
	for _, token := range []string{"valid", ""} {
		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		req.Header.Set(echo.HeaderOrigin, "http://localhost:3000")
		if token != "" {
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
		}
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
		if token == "" {
			require.Equal(t, http.StatusUnauthorized, w.Code)
		} else {
			require.Equal(t, http.StatusOK, w.Code)
		}
		require.Equal(
			t,
			"http://localhost:3000",
			w.Header().Get(echo.HeaderAccessControlAllowOrigin),
		)
	}
}

func TestCORSDisabledWhenNoOriginsConfigured(t *testing.T) {
	var query bd.BookQuery
	e := newTestServer(
		t,
		testDependencies{Verifier: verifier{}, Resolver: resolver{}, BookList: list{query: &query}},
	)
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	req.Header.Set(echo.HeaderOrigin, "http://localhost:3000")
	req.Header.Set(echo.HeaderAuthorization, "Bearer valid")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, w.Header().Get(echo.HeaderAccessControlAllowOrigin))
}
