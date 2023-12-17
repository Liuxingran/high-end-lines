package service

import (
	"context"
	"fmt"
	pb "high-end-lines/internal/grpc/proto/proto_go"
)

var _ pb.SearchServiceServer = &SearchServiceImpl{}

type SearchServiceImpl struct {
	pb.UnimplementedSearchServiceServer
}

func (s *SearchServiceImpl) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	fmt.Println("Search exec...", "request", req)
	return &pb.SearchResponse{
		Results: []*pb.Result{
			{
				Url:      "www.baidu.com",
				Title:    "baidu",
				Snippets: []string{"牛", "羊"},
			},
		},
	}, nil
}
