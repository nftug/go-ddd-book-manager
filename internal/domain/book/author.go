package book

import (
	"unicode/utf8"

	"github.com/samber/lo"
)

type Author struct {
	name  string
	order int
}

func newOrderedAuthor(name string, index int) (Author, error) {
	if name == "" {
		return Author{}, AuthorEmptyErr(index)
	}
	if utf8.RuneCountInString(name) > 255 {
		return Author{}, AuthorTooLongErr(index)
	}
	return Author{name, index}, nil
}
func RestoreAuthor(name string, orderIndex int) Author {
	return Author{name, orderIndex}
}
func (a Author) Name() string    { return a.name }
func (a Author) OrderIndex() int { return a.order }

type AuthorList struct{ values []Author }

func NewAuthorList(names []string) (AuthorList, error) {
	if len(names) == 0 {
		return AuthorList{}, AuthorsEmptyErr
	}
	values, err := lo.MapErr(names, func(name string, index int) (Author, error) {
		return newOrderedAuthor(name, index)
	})
	if err != nil {
		return AuthorList{}, err
	}
	seenNames := make(map[string]struct{}, len(values))
	for _, author := range values {
		if _, exists := seenNames[author.Name()]; exists {
			return AuthorList{}, AuthorsDuplicateErr
		}
		seenNames[author.Name()] = struct{}{}
	}
	return AuthorList{values}, nil
}
func RestoreAuthorList(values []Author) AuthorList {
	return AuthorList{lo.Clone(values)}
}
func (a AuthorList) Values() []Author { return lo.Clone(a.values) }
