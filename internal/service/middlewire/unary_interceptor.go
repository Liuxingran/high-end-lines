package middlewire

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "high-end-lines/internal/grpc/proto/proto_go"
	"time"
)

// OrderUnaryServerInterceptor 实现 unary interceptors
func OrderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

	// Pre-processing logic
	s := time.Now()

	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Post processing logic
	fmt.Printf("Method: %s, req: %s, resp: %s, latency: %s\n",
		info.FullMethod, req, m, time.Now().Sub(s))

	//WithDetails
	//return m, err
	//return m, status.New(codes.PermissionDenied, "OK").WithDetails().Err()
	st, err := status.New(codes.PermissionDenied, "OK").WithDetails(&pb.HttpErrorCode{
		Code:    pb.HttpErrorCode_SysErrorCode,
		Message: ParamErrorCode.String(),
	})
	if err == nil {
		fmt.Println("status new err", err)
	}
	return m, st.Err()
}
