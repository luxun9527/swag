

# **通过代码来定制生成文档。**

**通过代码来定制生成文档。**

swag生成api文档的方式是通过方法上面的注释，在日常开发中，我们还要去维护方法上面的注释。

如果我们请求参数有变化则要手动维护，有时候我们在使用一些框架的时候，获取参数，返回数据是一些固定的格式，尝试通过分析代码ast语法树来实现。

示例，我们使用gin框架。

**获取参数**：我们规定ShouldBindQuery,ShouldBindJSON等这些方法是获取参数的方法。这些方法的第一个参数为参数，JSON Query 标识位置

// @Param     data  body      example.ExaAttachmentCategory  true  "媒体库分类数据"

**获取返回值**：我们规定方法内最后一行的第二个参数为返回值，会生成类似

// @Success   200   {object}  response.Response  "添加媒体库分类"  的注释

**获取Summary**：summary为方法上的第一行注释

//@Summary GetAccountInfo 获取用户信息

**获取路由**：调用c.POST c.GET方法第一个参数

// @Router    /api/v1/account/GetAccountInfo [get]



实现原理是通过代码来生成注释，然后在swag生成文档的时候，插入到ast语法树中，这就表示你也可以在方法上加上swag的注释，同样是可以解析的。

如果方法上是swag注释超过5行则表示这使一个swag注释，我们不会再根据代码来生成。

```go
package account

import (
	 "apigen/ginPkg/model/request"
	"apigen/ginPkg/service"
	"github.com/gin-gonic/gin"
)

var AccountApi = &accountApi{}

type accountApi struct{}
// GetAccountInfo 获取用户信息
// @Tags      sync
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

```

下面这段代码，中请求参数是类型是request.GetAccountInfoReq，使用shouldBindJSON获取则表示是从请求体body中获取对应的注释是

```
// @Param     data  body  request.GetAccountInfoReq  true  "请求参数"
// @Success   200   {object}  response.AccountInfo  "成功"
```

