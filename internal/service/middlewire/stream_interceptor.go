package middlewire

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

func OrderStreamServerInterceptor(srv interface{},
	ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	// Pre-processing logic
	s := time.Now()

	// Invoking the StreamHandler to complete the execution of RPC invocation
	err := handler(srv, ss)

	// Post processing logic
	log.Printf("Method: %s, latency: %s\n", info.FullMethod, time.Now().Sub(s))

	return err
}

// SendMsg method call.
type wrappedStream struct {
	Recv []interface{}
	Send []interface{}
	grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	w.Recv = append(w.Recv, m)

	return err
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)

	w.Send = append(w.Send, m)

	return err
}

func newWrappedStream(s grpc.ServerStream) *wrappedStream {
	return &wrappedStream{
		make([]interface{}, 0),
		make([]interface{}, 0),
		s,
	}
}

func OrderDoubleStreamServerInterceptor(srv interface{},
	ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Pre-processing logic
	time.Sleep(4 * time.Second)
	s := time.Now()
	var timestamp string
	//读取客户端的metadata
	md, ok := metadata.FromIncomingContext(ss.Context())
	if ok && len(md.Get("time")) > 0 {
		timestamp = md.Get("time")[0]
	}
	fmt.Println("server recv client request timestamp", timestamp)
	// Invoking the StreamHandler to complete the execution of RPC invocation
	ss.SendHeader(metadata.Pairs("server1", "server1"))
	ss.SetTrailer(metadata.Pairs("trailer", "trailer"))
	nss := newWrappedStream(ss)
	fmt.Println("111111111")
	err := handler(srv, nss)
	if err != nil {
		fmt.Println("nss SetHeader err", err)
	}
	fmt.Println("2222222222")
	// Post processing logic
	log.Printf("Method: %s, req: %+v, resp: %+v, latency: %s\n",
		info.FullMethod, nss.Recv, nss.Send, time.Now().Sub(s))

	return err
}
