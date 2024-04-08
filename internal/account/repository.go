package account

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var ErrDoesNotExist = errors.New("account not found")
var ErrInsufficientBalance = errors.New("insufficient balance")

func NewRepository(db *sql.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("db cannot be nil")
	}
	return &Repository{
		db: db,
	}, nil
}

type Repository struct {
	db *sql.DB
}

func (r *Repository) Create(ctx context.Context) (int64, error) {
	res := r.db.QueryRowContext(ctx, "INSERT INTO accounts DEFAULT VALUES RETURNING id")
	if res.Err() != nil {
		return 0, errors.Wrap(res.Err(), "database request failed")
	}

	var id int64
	if err := res.Scan(&id); err != nil {
		return 0, errors.Wrap(err, "could not read last inserted id")
	}

	return id, nil
}

func (r *Repository) FindById(ctx context.Context, id int64) (*Record, error) {
	res := r.db.QueryRowContext(ctx, "SELECT id, balance FROM accounts WHERE id = $1", id)
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "database query failed")
	}

	record := &Record{}
	if err := res.Scan(&record.Id, &record.Balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrapf(ErrDoesNotExist, "account id=%d", id)
		}
		return nil, errors.Wrap(err, "could not scan database query result into struct")
	}

	return record, nil
}

func (r *Repository) TopUp(ctx context.Context, target *Record, amount int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE accounts SET balance = balance+$1 WHERE id = $2", amount, target.Id)
	if err != nil {
		return errors.Wrapf(err, "could not top-up account with id=%d", target.Id)
	}

	return nil
}

func (r *Repository) Transfer(ctx context.Context, source *Record, target *Record, amount int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "could not start transaction")
	}
	defer tx.Rollback()

	res := tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1", source.Id)
	if err := res.Err(); res.Err() != nil {
		return errors.Wrapf(err, "could not query source account with id=%d", source.Id)
	}
	var sourceBalance int
	if err := res.Scan(&sourceBalance); err != nil {
		return errors.Wrapf(err, "could not balance of source account with id=%d", source.Id)
	}

	if sourceBalance-amount < 0 {
		return errors.Wrapf(ErrInsufficientBalance, "available balance=%d, required amount=%d", sourceBalance, amount)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance=balance+$1 WHERE id=$2", amount, target.Id); err != nil {
		return errors.Errorf("could not update balance of target account with id=%d", target.Id)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance=balance-$1 WHERE id=$2", amount, source.Id); err != nil {
		return errors.Errorf("could not update balance of source account with id=%d", source.Id)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "could not commit transaction of amount=%d, from account=%d, to account=%d", amount, source.Id, target.Id)
	}

	return nil
}
