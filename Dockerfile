FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /out/todo-ist .

FROM debian:bookworm-slim
WORKDIR /srv
COPY --from=build /out/todo-ist ./todo-ist
COPY app/ ./app/
ENV PORT=8080 \
    DB_PATH=/data/todo-ist.db \
    STATIC_DIR=/srv/app
EXPOSE 8080
CMD ["./todo-ist"]
