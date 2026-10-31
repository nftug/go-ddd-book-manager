package common

import (
	domain "go-ddd-book-manager/internal/domain/common"
	"time"

	"github.com/samber/do/v2"
)

type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }

func NewClock(do.Injector) (domain.Clock, error) { return Clock{}, nil }
