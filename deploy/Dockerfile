# syntax=docker/dockerfile:1.7

FROM node:24.18.1-bookworm AS web-build
WORKDIR /src/web
RUN corepack enable && corepack prepare pnpm@11.18.0 --activate
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM golang:1.26.5-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY db/ ./db/
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/server-templates/ ./web/server-templates/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
  && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=go-build /out/server /app/server
COPY --from=go-build /out/worker /app/worker
COPY --from=web-build /src/web/dist /app/web/dist
ENV HTTP_ADDR=:8080
ENV STATIC_DIR=/app/web/dist
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
