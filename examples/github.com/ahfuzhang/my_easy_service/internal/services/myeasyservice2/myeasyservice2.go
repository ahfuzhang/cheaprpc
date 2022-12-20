package myeasyservice2

import (
	"context"
	"fmt"

	"github.com/ahfuzhang/my_easy_service/pb"
)

type MyEasyService2Imp struct {
}

var Instance = &MyEasyService2Imp{
	// todo: add init field if you need
}

func init() {
	// todo: init Instance if you need
}

func (s *MyEasyService2Imp) GetEchoInfo(ctx context.Context, req *pb.GetReq) (*pb.GetRsp, error) {
	// TODO: add biz code here
	return nil, fmt.Errorf("not implement")
}
func (s *MyEasyService2Imp) Save(ctx context.Context, req *pb.SaveReq) (*pb.SaveRsp, error) {
	// TODO: add biz code here
	return nil, fmt.Errorf("not implement")
}
