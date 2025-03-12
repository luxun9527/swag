package gen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

// // @Param     data  query     request.SysOperationRecordSearch                        true  "页码, 每页大小, 搜索条件"
// @Param     data   query    request.GetLatestTsReq  true  "请求参数"
// @Success   200   {object}  response.GetLatestTsResp  "成功"

const (
	reqPrefix  = "@Param"
	respPrefix = "@Success"
)

func (req ReqParam) GetSwagComment() string {
	return fmt.Sprintf("%s %s %s %s %v %s", reqPrefix, "data", req.ReqType, req.ReqVarType, true, "请求参数")
}

type ReqParam struct {
	ReqVarType string
	ReqType    string //POST json GET query
}

type Resp struct {
	RespVarType string
	RespType    string
}

type FuncDetail struct {
	FuncName string
	ReqParam *ReqParam
	Resp     *Resp
	Comment  string
}

type FileDetail struct {
	Filename       string
	PkgPath        string
	FuncDetailList []*FuncDetail
}

func (resp Resp) GetSwagComment() string {
	return fmt.Sprintf("%s %d %s %s %s ", respPrefix, 200, resp.RespType, resp.RespVarType, "成功")
}

func parseReqResp(dir string) {

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports,
		Context:    nil,
		Logf:       nil,
		Dir:        dir,
		Env:        nil,
		BuildFlags: nil,
		Fset:       nil,
		ParseFile:  nil,
		Tests:      false,
		Overlay:    nil,
	}
	//E:\openproject\apigen\ginPkg\...
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal("包加载失败:", err)
	}

	var fdList []*FileDetail
	for _, v := range pkgs {
		fileDetail := processPackage(v)
		fdList = append(fdList, fileDetail...)
	}
	log.Printf("fdList %v", fdList)
}

func processPackage(pkg *packages.Package) []*FileDetail {
	var fdList []*FileDetail
	for _, file := range pkg.Syntax {
		fd := &FileDetail{}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				funcDetail := processFunction(pkg, node)
				if funcDetail == nil {
					return true
				}
				fd.PkgPath = pkg.PkgPath
				fd.FuncDetailList = append(fd.FuncDetailList, funcDetail)
				fdList = append(fdList, fd)
			case *ast.CallExpr:
			}
			return true
		})

	}
	return fdList
}

func processFunction(pkg *packages.Package, fn *ast.FuncDecl) *FuncDetail {
	if isGinHandler(fn) {
		req, resp, bind := parseHandlerDetails(pkg, fn)
		comment := extractSummary(fn.Doc)
		return &FuncDetail{
			ReqParam: &ReqParam{
				ReqVarType: req,
				ReqType:    bind,
			},
			Resp: &Resp{
				RespVarType: resp,
				RespType:    "",
			},
			Comment: comment,
		}
	}
	return nil
}

func isGinHandler(fn *ast.FuncDecl) bool {
	params := fn.Type.Params
	if params == nil || len(params.List) != 1 {
		return false
	}

	return isGinContext(params.List[0].Type)
}

func isGinContext(expr ast.Expr) bool {
	if star, ok := expr.(*ast.StarExpr); ok {
		if sel, ok := star.X.(*ast.SelectorExpr); ok {
			return sel.Sel.Name == "Context" && sel.X.(*ast.Ident).Name == "gin"
		}
	}
	return false
}

func getHandlerKey(fn *ast.FuncDecl) string {
	recv := ""
	if fn.Recv != nil {
		if t, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
			recv = t.X.(*ast.Ident).Name + "."
		}
	}
	return recv + fn.Name.Name
}

func extractSummary(doc *ast.CommentGroup) string {
	if doc == nil || len(doc.List) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(doc.List[0].Text, "//"))
}

func parseHandlerDetails(pkg *packages.Package, fn *ast.FuncDecl) (req, res, bind string) {
	// 解析请求参数
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if bind == "" && isBindMethod(call) {
				bind = getBindMethod(call)
				req = getBindType(pkg, call)
			}
		}
		return true
	})

	// 解析响应参数（最后一行）
	if lastStmt := fn.Body.List[len(fn.Body.List)-1]; lastStmt != nil {
		if expr, ok := lastStmt.(*ast.ExprStmt); ok {
			if call, ok := expr.X.(*ast.CallExpr); ok {
				if isResponseMethod(call) && len(call.Args) > 1 {
					res = getTypeName(pkg, call.Args[1])
				}
			}
		}
	}
	log.Printf("%s %s %s", req, res, bind)

	return req, res, bind
}

func isBindMethod(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return strings.HasPrefix(sel.Sel.Name, "ShouldBind")
	}
	return false
}

func getBindMethod(call *ast.CallExpr) string {
	sel := call.Fun.(*ast.SelectorExpr)
	return strings.TrimPrefix(sel.Sel.Name, "ShouldBind")
}

func getBindType(pkg *packages.Package, call *ast.CallExpr) string {
	if len(call.Args) == 0 {
		return ""
	}

	switch arg := call.Args[0].(type) {
	case *ast.UnaryExpr:
		return getTypeName(pkg, arg.X)
	case *ast.Ident:
		return getTypeName(pkg, arg)
	}
	return ""
}

func isResponseMethod(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		switch sel.Sel.Name {
		case "JSON", "XML", "String":
			return true
		}
	}
	return false
}

func getTypeName(pkg *packages.Package, expr ast.Expr) string {
	tv := pkg.TypesInfo.Types[expr]

	named, ok := tv.Type.(*types.Named)
	if ok {
		return fmt.Sprintf("%s.%s", named.Obj().Pkg().Name(), named.Obj().Name())

	}

	return fmt.Sprint(tv.Type)
}

func getParamLocation(bindMethod string) string {
	switch bindMethod {
	case "Query":
		return "query"
	case "JSON":
		return "body"
	default:
		return "formData"
	}
}

// 辅助函数
func isHTTPMethod(name string) bool {
	switch name {
	case "GET", "POST", "PUT", "DELETE", "OPTIONS":
		return true
	}
	return false
}

func getStringLiteral(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return strings.Trim(lit.Value, `"`)
	}
	return ""
}

func parseHandler(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		return sel.Sel.Name
	}
	return ""
}
