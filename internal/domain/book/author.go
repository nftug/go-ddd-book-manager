package book

import (
	"unicode/utf8"

	"go-ddd-book-manager/internal/domain/common"

	"github.com/samber/lo"
)

type AuthorName struct{ value string }

func NewAuthorName(value string) (AuthorName, error) {
	if value == "" {
		return AuthorName{}, common.NewValidationError("book.author.empty")
	}
	if utf8.RuneCountInString(value) > 255 {
		return AuthorName{}, common.NewValidationError("book.author.too_long")
	}
	return AuthorName{value}, nil
}
func RestoreAuthorName(value string) AuthorName { return AuthorName{value} }
func (v AuthorName) Value() string              { return v.value }

type OrderedAuthor struct {
	name  AuthorName
	order int
}

func (a OrderedAuthor) Name() string    { return a.name.Value() }
func (a OrderedAuthor) OrderIndex() int { return a.order }
func RestoreOrderedAuthor(name AuthorName, orderIndex int) OrderedAuthor {
	return OrderedAuthor{name, orderIndex}
}

type AuthorList struct{ values []OrderedAuthor }

func NewAuthorList(names []string) (AuthorList, error) {
	if len(names) == 0 {
		return AuthorList{}, common.NewValidationError("book.authors.empty")
	}
	authorNames, err := lo.MapErr(names, func(name string, _ int) (AuthorName, error) {
		return NewAuthorName(name)
	})
	if err != nil {
		return AuthorList{}, err
	}
	if len(lo.Uniq(authorNames)) != len(authorNames) {
		return AuthorList{}, common.NewValidationError("book.authors.duplicate")
	}
	values := lo.Map(authorNames, func(name AuthorName, index int) OrderedAuthor {
		return OrderedAuthor{name, index}
	})
	return AuthorList{values}, nil
}
func RestoreAuthorList(values []OrderedAuthor) AuthorList {
	return AuthorList{lo.Clone(values)}
}
func (a AuthorList) Values() []OrderedAuthor { return lo.Clone(a.values) }
