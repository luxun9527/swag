package router

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag/example/gin/api/account"
)

func InitRouter(e *gin.Engine) {
	g := e.Group("/account")
	{
		g.GET("/getUserInfo", account.AccountApi.GetAccountInfo)
	}
	g1 := e.Group("/base")
	{
		g1.POST("/login", account.BaseApi.Login)
		g1.POST("/register", account.BaseApi.Register)
	}

}
