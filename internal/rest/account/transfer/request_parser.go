package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var ErrRequestBodyNotSet = errors.New("request body is mandatory")

type RequestParser func(ctx context.Context, r *http.Request) (*Request, error)

func GetRequestParser() RequestParser {
	return func(ctx context.Context, r *http.Request) (*Request, error) {
		if r.Body == http.NoBody {
			return nil, ErrRequestBodyNotSet
		}

		variables := mux.Vars(r)
		id, err := strconv.ParseInt(variables["id"], 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "cannot parse account id")
		}

		req := &Request{
			Source: id,
			Data:   &RequestData{},
		}

		if err := json.NewDecoder(r.Body).Decode(req.Data); err != nil {
			return nil, errors.Wrap(err, "request body not a valid json")
		}

		return req, nil
	}
}
