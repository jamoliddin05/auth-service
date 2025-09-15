FROM golang:1.25

WORKDIR /app

RUN go install github.com/air-verse/air@latest
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd/

CMD ["/bin/app"]