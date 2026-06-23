# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Builder: regenerate protobuf/gRPC stubs with buf, then build a static binary
# ---------------------------------------------------------------------------
FROM golang:1.25.3-alpine AS builder

# git is required by `go install` (module fetch) and by buf.
RUN apk add --no-cache git

# Install codegen tooling. Versions are pinned to match the dev environment
# and go.mod:
#   - buf                : v2 config (buf.yaml / buf.gen.yaml) requires buf >= 1.32
#   - protoc-gen-go      : matches google.golang.org/protobuf in go.mod (v1.36.11)
#   - protoc-gen-go-grpc : matches the locally used generator (v1.6.1)
# buf ships its own protocol compiler, so protoc itself is not needed.
ENV GOBIN=/usr/local/bin
RUN go install github.com/bufbuild/buf/cmd/buf@v1.66.1 \
 && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 \
 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1

WORKDIR /src

# Cache dependencies before copying the rest of the sources.
COPY go.mod go.sum ./
RUN go mod download

# Copy sources (see .dockerignore for what is excluded).
COPY . .

# Regenerate gRPC/protobuf code from proto/ into internal/grpc/pb.
RUN buf generate

# Build a fully static binary (no CGO) for the scratch-like final stage.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server/

# ---------------------------------------------------------------------------
# Final: minimal runtime image. Migrations are NOT baked in — they are applied
# out-of-band against the staging database.
# ---------------------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/server /usr/local/bin/server

# Run as an unprivileged user.
RUN adduser -D -u 10001 appuser
USER appuser

# SERVER_PORT (REST) and GRPC_PORT.
EXPOSE 8080 50051

ENTRYPOINT ["/usr/local/bin/server"]
