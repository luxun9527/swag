package account

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/gin/model/request"
	"github.com/swaggo/swag/example/gin/service"
)

var AccountApi = &accountApi{}

type accountApi struct{}

// GetAccountInfo 注册
func (*accountApi) GetAccountInfo(c *gin.Context) {
	// 获取用户信息
	var req request.GetAccountInfoReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": err.Error()})
	}

	accountInfo, err := service.AccountService.GetAccountInfo(req)
	if err != nil {
		return
	}

	c.JSON(200, accountInfo)
}
