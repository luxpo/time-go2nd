package rpc

import (
	"context"
	"log"
)

type UserService struct {
	GetByID func(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error)
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
}

func (s *UserServiceServer) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	log.Println(req)
	return &GetByIDResp{
		Msg: "hi",
	}, nil
}

func (s *UserServiceServer) Name() string {
	return "user-service"
}
