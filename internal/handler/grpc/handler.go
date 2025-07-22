package grpc

import (
	"context"
	"id-generator/internal/usecase"
	pb "id-generator/snowflake-pb"
)

type IDHandler struct {
	pb.UnimplementedIDServiceServer
	uc *usecase.IDUsecase
}

func NewIDHandler(uc *usecase.IDUsecase) *IDHandler {
	return &IDHandler{uc: uc}
}

func (h *IDHandler) Generate(ctx context.Context, req *pb.IDRequest) (*pb.IDResponse, error) {
	id, err := h.uc.GenerateID()
	return &pb.IDResponse{Id: id}, err
}

func (h *IDHandler) GenerateIDs(ctx context.Context, req *pb.IDsRequest) (*pb.IDsResponse, error) {
	ids, duration, err := h.uc.GenerateIDs(int(req.Count))
	return &pb.IDsResponse{Ids: *ids, DurationMs: duration}, err
}
