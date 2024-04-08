package transfer

import (
	"github.com/pkg/errors"
)

var ErrRequestDataNotSet = errors.New("request data is mandatory")
var ErrRequestInvalidAmountLt0 = errors.New("amount cannot be less than 1")

type Request struct {
	Source int64
	Data   *RequestData
}

func (r *Request) Validate() error {
	if r.Data == nil {
		return ErrRequestDataNotSet
	}

	if err := r.Data.Validate(); err != nil {
		return errors.Wrap(err, "invalid request data")
	}

	return nil
}

type RequestData struct {
	Target int64 `json:"target"`
	Amount int   `json:"amount"`
}

func (d *RequestData) Validate() error {
	if d.Amount < 1 {
		return ErrRequestInvalidAmountLt0
	}

	return nil
}
