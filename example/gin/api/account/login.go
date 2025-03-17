package account

import (
	"github.com/gin-gonic/gin"
	commonResp "github.com/swaggo/swag/example/gin/model/common/response"
	"github.com/swaggo/swag/example/gin/model/request"
	"github.com/swaggo/swag/example/gin/model/response"
	"log"
)

var BaseApi = &baseApi{}

type baseApi struct{}

// Login 注册
func (*baseApi) Login(c *gin.Context) {
	// 获取用户信息
	var req request.LoginReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": err.Error()})
	}
	name := c.Query("name")
	log.Println(name)
	var resp response.LoginResp
	commonResp.OkWithData(c, resp)
}

// Register 注册
func (*baseApi) Register(c *gin.Context) {
	// 获取用户信息
	var req request.RegisterReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": err.Error()})
	}
	commonResp.Ok(c)
}
