# Setup
- Copy `deployment/example.env` as `deployment/.env` and replace values if needed
- Copy `configs/example.app.yaml` as `configs/app.yaml` and replace values if needed

# How to run
- Run `make up` to start postgres instances
- Run `go run cmd/server/main.go ./...` to start the application

# How to run tests
```bash
export POSTGRES_URI_TEST="postgres://su-exc:qwerty123@127.0.0.1:5544/su?sslmode=disable"
go test ./...
```
Note that the `POSTGRES_URI_TEST` values might change depending on the changes made in the `.env` file

# How to build docker image
- `docker build . -f deployment/docker/Dockerfile`

# API Docs
Navigate to `SU-exc.postman_collection.json` which contains Postman Collection with examples
