FROM golang:alpine AS BUILDER

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o build .


FROM alpine AS RUNNER

WORKDIR /app

COPY --from=BUILDER /app/build .

ENTRYPOINT ["./build"]

