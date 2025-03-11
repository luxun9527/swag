package gen

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
)

// // @Param     data  query     request.SysOperationRecordSearch                        true  "页码, 每页大小, 搜索条件"
// @Param     data   query    request.GetLatestTsReq  true  "请求参数"
// @Success   200   {object}  response.GetLatestTsResp  "成功"

const (
	reqPrefix  = "@Param"
	respPrefix = "@Success"
)

type ReqParam struct {
	Filename   string
	MethodName string
	ReqVarType string
}

func (req ReqParam) GetSwagComment() string {
	return fmt.Sprintf("%s %s %s %s %v %s", reqPrefix, "data", "query", req.ReqVarType, true, "请求参数")
}

type Resp struct {
	Filename    string
	methodName  string
	RespVarType string
	RespType    string
}

func (resp Resp) GetSwagComment() string {
	return fmt.Sprintf("%s %d %s %s %s ", respPrefix, 200, resp.RespType, resp.RespVarType, "成功")
}

func parseReqResp(dir string) {

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, dir+"/...")
	if err != nil {
		log.Fatal("包加载失败:", err)
	}
	for _, v := range pkgs {

	}
}
func processPackage(pkg *packages.Package, comments map[string]string, routes *[]RouteInfo) {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				processFunction(pkg, node, comments, routes)
			case *ast.CallExpr:
				processRouteRegistration(pkg, node, comments, routes)
			}
			return true
		})
	}
}
