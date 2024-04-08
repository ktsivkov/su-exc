package account

import "context"

type Creator interface {
	Create(ctx context.Context) (int64, error)
}

type Finder interface {
	FindById(ctx context.Context, id int64) (*Record, error)
}

type TopUpper interface {
	TopUp(ctx context.Context, target *Record, amount int) error
}

type Transferrer interface {
	Transfer(ctx context.Context, source *Record, target *Record, amount int) error
}
