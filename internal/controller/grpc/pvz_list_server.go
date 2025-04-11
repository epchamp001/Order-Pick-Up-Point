package grpc

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"order-pick-up-point/api/pb"
	"order-pick-up-point/internal/errs"
	grpcService "order-pick-up-point/internal/service/grpc"
)

type PvzServer struct {
	pb.UnimplementedPVZServiceServer
	svc grpcService.PvzService
}

func NewPvzServer(svc grpcService.PvzService) *PvzServer {
	return &PvzServer{svc: svc}
}

func (s *PvzServer) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	pvzs, err := s.svc.GetPVZList(ctx)
	if err != nil {
		return nil, errs.Wrap(err, errs.ErrInternalCode, "failed to get PVZ list")
	}

	var pbPvzs []*pb.PVZ
	for _, pvz := range pvzs {
		ts := timestamppb.New(pvz.RegistrationDate)
		pbPvzs = append(pbPvzs, &pb.PVZ{
			Id:               pvz.ID,
			RegistrationDate: ts,
			City:             pvz.City,
		})
	}

	return &pb.GetPVZListResponse{
		Pvzs: pbPvzs,
	}, nil
}
