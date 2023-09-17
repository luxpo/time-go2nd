package rpc

import (
	"context"
	"log"

	"github.com/luxpo/time-go2nd/micro/proto/gen"
)

type UserService struct {
	GetByID func(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error)

	GetByIDProto func(ctx context.Context, req *gen.GetByIDReq) (*gen.GetByIDResp, error)
}

type GetByIDReq struct {
	ID int
}

type GetByIDResp struct {
	Msg string
}

func (s UserService) Name() string {
	return "user-service"
}

type UserServiceServer struct {
	Msg string
	Err error
}

func (s *UserServiceServer) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	log.Println(req)
	return &GetByIDResp{
		Msg: s.Msg,
	}, s.Err
}

func (s *UserServiceServer) GetByIDProto(ctx context.Context, req *gen.GetByIDReq) (*gen.GetByIDResp, error) {
	log.Println(req)
	return &gen.GetByIDResp{
		User: &gen.User{
			Name: s.Msg,
		},
	}, s.Err
}

func (s *UserServiceServer) Name() string {
	return "user-service"
}
