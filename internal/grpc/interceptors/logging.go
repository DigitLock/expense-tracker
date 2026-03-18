package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err)
		log.Printf("[gRPC] method=%s duration=%s status=%s",
			info.FullMethod,
			time.Since(start),
			code,
		)

		return resp, err
	}
}
