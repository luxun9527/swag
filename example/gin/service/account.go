package service

import (
	"github.com/swaggo/swag/example/gin/model/request"
	"github.com/swaggo/swag/example/gin/model/response"
)

var AccountService accountService

type accountService struct{}

func (accountService) GetAccountInfo(req request.GetAccountInfoReq) (*response.AccountInfo, error) {

	return nil, nil

}
