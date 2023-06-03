package rpc

import "context"

type UserService struct {
	GetByID func(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error)
}

type GetByIDReq struct {
	ID int
}

type GetByIDResp struct {
}

func (s UserService) Name() string {
	return "user-service"
}
