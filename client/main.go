package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	pb "high-end-lines/internal/grpc/proto/proto_go"
	"io"
	"log"
	"strconv"
	"time"
)

func orderUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Pre-processor phase
	s := time.Now()

	// Invoking the remote method
	err := invoker(ctx, method, req, reply, cc, opts...)

	// Post-processor phase
	log.Printf("method: %s, req: %s, resp: %s, latency: %s\n",
		method, req, reply, time.Now().Sub(s))

	return err
}

func OrderStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	// Pre-processing logic
	s := time.Now()

	cs, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		fmt.Println("streamer err", err)
	}
	// Post processing logic
	log.Printf("method: %s, latency: %s\n", method, time.Now().Sub(s))

	return cs, err
}

// SendMsg method call.
type wrappedStream struct {
	method string
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	fmt.Println("client get trailer, ", w.ClientStream.Trailer())
	log.Printf("method: %s, res: %s\n", w.method, m)

	return err
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)

	log.Printf("method: %s, req: %s\n", w.method, m)

	return err
}

func newWrappedStream(method string, s grpc.ClientStream) *wrappedStream {
	return &wrappedStream{
		method,
		s,
	}
}

func orderDoubleStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc,
	cc *grpc.ClientConn, method string, streamer grpc.Streamer,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// Pre-processing logic
	s := time.Now()

	var timestamp string
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok && len(md.Get("time")) > 0 {
		timestamp = md.Get("time")[0]
	} else {
		timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
		ctx = metadata.AppendToOutgoingContext(ctx, "time", timestamp)
	}
	fmt.Println("client request timestamp", timestamp)
	fmt.Println("aaaaaaaaaa")
	cs, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		fmt.Println("streamer err", err)
	}

	fmt.Println("bbbbbbbbbb")
	header, err := cs.Header()
	if err == nil {
		for k, v := range header {
			fmt.Println("k", k, "v", v)
		}
	} else {
		fmt.Println("cs Header err", err)
	}
	//fmt.Println("client get trailer, ", cs.Trailer())
	fmt.Println("cccccccccc")

	// Post processing logic
	log.Printf("method: %s, latency: %s\n", method, time.Now().Sub(s))

	return newWrappedStream(method, cs), err
}

func main() {
	//conn, err := grpc.Dial("127.0.0.1:8009",
	//	grpc.WithInsecure(),
	//	grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
	//	grpc.WithChainStreamInterceptor(OrderStreamClientInterceptor, orderDoubleStreamClientInterceptor),
	//)

	//client 超时控制
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "127.0.0.1:8009", grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithChainStreamInterceptor(orderDoubleStreamClientInterceptor))
	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("DeadlineExceeded err", err)
			panic(err)
		} else {
			fmt.Println("not DeadlineExceeded err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//client := pb.NewSearchServiceClient(conn)
	//
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	//
	//// Get Order
	//retrievedOrder, err := client.Search(ctx, &pb.SearchRequest{
	//	PageNum:  1,
	//	Query:    "1",
	//	PageSize: 10,
	//})
	//if err != nil {
	//	fmt.Println("client.Search err", err)
	//	return
	//}
	//status 错误处理
	//st, ok := status.FromError(err)
	//if ok && st.Code() == 0 {
	//	fmt.Println("success st", st.Code(), st.Message())
	//} else {
	//	fmt.Println("failed st", st.Code(), st.Message())
	//}

	//log.Print("GetOrder Response -> : ", retrievedOrder)
	c := pb.NewOrderManagementClient(conn)
	ctx, cancelFn := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFn()

	stream, err := c.ProcessOrders(ctx)
	//status WithDetails错误处理
	st, ok := status.FromError(err)
	if !ok {
		fmt.Println("status format err", err)
		return
	}
	switch st.Code() {
	case codes.PermissionDenied:
		for _, d := range st.Details() {
			switch d.(type) {
			case *pb.HttpErrorCode:
				fmt.Println("pb HttpErrorCode Details", st.Details())
			}
		}
	case codes.DeadlineExceeded:
		fmt.Println("response timeout", st)
		panic(err)
	}

	go func() {
		if err := stream.Send(&wrapperspb.StringValue{Value: "101"}); err != nil {
			panic(err)
		}

		if err := stream.Send(&wrapperspb.StringValue{Value: "102"}); err != nil {
			panic(err)
		}

		if err := stream.CloseSend(); err != nil {
			panic(err)
		}
	}()

	for {
		combinedShipment, err := stream.Recv()
		if err == io.EOF {
			break
		}
		log.Println("Combined shipment : ", combinedShipment.OrderList)
	}
}
