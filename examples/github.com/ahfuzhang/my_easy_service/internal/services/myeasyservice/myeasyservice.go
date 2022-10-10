package myeasyservice

import (
	"context"
	"fmt"

	"github.com/ahfuzhang/my_easy_service/pb"
)

type MyEasyServiceImp struct {
}

var Instance = &MyEasyServiceImp{
    // todo: add init field if you need
}

func init(){
   // todo: init Instance if you need
}


func (s *MyEasyServiceImp) GetEchoInfo(ctx context.Context, req *pb.GetReq) (*pb.GetRsp, error) {
    // TODO: add biz code here
	return nil, fmt.Errorf("not implement")
}
func (s *MyEasyServiceImp) Save(ctx context.Context, req *pb.SaveReq) (*pb.SaveRsp, error) {
    // TODO: add biz code here
	return nil, fmt.Errorf("not implement")
}

