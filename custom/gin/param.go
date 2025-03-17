package gin

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

const (
	reqPrefix     = "//@Param"
	respPrefix    = "//@Success"
	summaryPrefix = "//@Summary"
	acceptPrefix  = "//@Accept application/json"
	producePrefix = "//@Produce application/json"
)

func (req ReqParam) GetSwagComment() string {
	if req.ReqVarName == "" {
		req.ReqVarName = "data"
	}
	return fmt.Sprintf("%s %s %s %s %v %s", reqPrefix, req.ReqVarName, req.Location, req.ReqVarType, true, "\"请求参数\"")
}

type ReqParam struct {
	ReqVarType string //shouldBindJson shouldBindQuery 参数名
	Location   string //query(post方法form提交也是query)  json header
	ReqVarName string // query postForm 参数名
}

type Resp struct {
	RespVarType string
	RespType    string
}

type FuncDetail struct {
	FuncName string
	ReqParam []*ReqParam
	Resp     *Resp
	Comment  string
	Router   string
}

func (f *FuncDetail) BuildComment() []*ast.Comment {
	summary := &ast.Comment{Text: fmt.Sprintf("%s %s", summaryPrefix, f.Comment)}
	accept := &ast.Comment{Text: acceptPrefix}
	produce := &ast.Comment{Text: producePrefix}

	resp := &ast.Comment{Text: f.Resp.GetSwagComment()}
	router := &ast.Comment{Text: f.Router}
	comments := []*ast.Comment{summary, accept, produce, resp, router}
	for _, v := range f.ReqParam {
		c := &ast.Comment{Text: v.GetSwagComment()}
		log.Println(v.GetSwagComment())
		comments = append(comments, c)
	}
	return comments
}

type FileDetail struct {
	Filename       string
	PkgPath        string
	FuncDetailList []*FuncDetail
}

func (resp Resp) GetSwagComment() string {
	return fmt.Sprintf("%s %d %s %s %s ", respPrefix, 200, resp.RespType, resp.RespVarType, "\"成功\"")
}

func processPackage(pkg *packages.Package) ([]*FileDetail, map[string]*RouterInfo) {
	var (
		fdList     []*FileDetail
		routerInfo = map[string]*RouterInfo{}
	)
	//pkg.Syntax 为文件整个节点
	for i, file := range pkg.Syntax {
		fd := &FileDetail{}
		ast.Inspect(file, func(n ast.Node) bool {
			fd.Filename = pkg.GoFiles[i]
			fd.PkgPath = pkg.PkgPath
			switch node := n.(type) {
			case *ast.FuncDecl:
				funcDetail := processFunction(pkg, node)
				if funcDetail == nil {
					r := GetRouterInfo(node)
					for k, v := range r {
						routerInfo[k] = v
					}
					return true
				}

				fd.FuncDetailList = append(fd.FuncDetailList, funcDetail)

			}
			return true
		})
		if len(fd.FuncDetailList) == 0 {
			continue
		}
		fdList = append(fdList, fd)
	}

	return fdList, routerInfo
}

func processFunction(pkg *packages.Package, fn *ast.FuncDecl) *FuncDetail {

	if isGinHandler(fn) && (fn.Doc == nil || len(fn.Doc.List) < 5) {
		req, resp := parseHandlerDetails(pkg, fn)
		comment := extractSummary(fn.Doc)
		respType := "{object}"

		return &FuncDetail{
			ReqParam: req,
			Resp: &Resp{
				RespVarType: resp,
				RespType:    respType,
			},
			Comment:  comment,
			FuncName: fn.Name.Name,
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

func extractSummary(doc *ast.CommentGroup) string {
	if doc == nil || len(doc.List) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(doc.List[0].Text, "//"))
}

func parseHandlerDetails(pkg *packages.Package, fn *ast.FuncDecl) (req []*ReqParam, resp string) {
	// 解析请求参数
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		typeName := getTypeName(pkg, sel.X)
		if typeName != "gin.Context" {
			return true
		}
		//调用参数是gin.Context

		reqDetail := getReqInfoList(sel.Sel, call.Args)
		if reqDetail == nil {
			return true
		}
		reqVarType := getReqVarType(pkg, call)
		if reqDetail.ReqVarName != "" {
			reqVarType = reqDetail.ReqVarType
		}
		location := getParamLocation(reqDetail.Location)
		reqParam := &ReqParam{
			ReqVarType: reqVarType,
			Location:   location,
			ReqVarName: reqDetail.ReqVarName,
		}
		req = append(req, reqParam)

		return true
	})

	// 解析响应参数（最后一行）
	if lastStmt := fn.Body.List[len(fn.Body.List)-1]; lastStmt != nil {
		if expr, ok := lastStmt.(*ast.ExprStmt); ok {
			if call, ok := expr.X.(*ast.CallExpr); ok {
				if len(call.Args) > 1 {
					resp = getTypeName(pkg, call.Args[1])
				}
			}
		}
	}
	log.Printf("%v %s ", req, resp)
	return
}

const (
	shouldBindPrefix = "ShouldBind"
	bindPrefix       = "Bind"
	queryPrefix      = "Query"
	postFormPrefix   = "PostForm"
)

func getReqInfoList(sel *ast.Ident, args []ast.Expr) *ReqParam {

	switch {
	case strings.HasPrefix(sel.Name, shouldBindPrefix):
		return &ReqParam{
			Location: strings.TrimPrefix(sel.Name, shouldBindPrefix),
		}
	case strings.HasPrefix(sel.Name, bindPrefix):
		return &ReqParam{
			Location: strings.TrimPrefix(sel.Name, bindPrefix),
		}
	case strings.HasPrefix(sel.Name, queryPrefix), strings.HasPrefix(sel.Name, postFormPrefix):
		if len(args) > 0 {
			arg := args[0]
			switch arg := arg.(type) {
			case *ast.BasicLit:
				// 如果是字面量（如 "id"），直接返回值
				return &ReqParam{
					ReqVarType: "string",
					Location:   "Query",
					ReqVarName: strings.Trim(arg.Value, `"`),
				}
			case *ast.Ident:
				// 如果是变量（如 id），返回变量名
				return &ReqParam{
					ReqVarType: "string",
					Location:   "Query",
					ReqVarName: strings.Trim(arg.Name, `"`),
				}
			}
		}

	}

	return nil
}

func getReqVarType(pkg *packages.Package, call *ast.CallExpr) string {
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

func getTypeName(pkg *packages.Package, expr ast.Expr) string {
	tv := pkg.TypesInfo.Types[expr]
	return parseType(tv.Type)
}
func parseType(tv interface{}) string {
	switch t := tv.(type) {
	case *types.Named:
		if pkg := t.Obj().Pkg(); pkg == nil {
			return t.Obj().Name()
		} else {
			return fmt.Sprintf("%s.%s", pkg.Name(), t.Obj().Name())
		}

	case *types.Slice:
		return "[]" + parseType(t.Elem())
	case *types.Pointer:
		return parseType(t.Elem())
	default:
		return fmt.Sprint(tv)

	}
}

func getParamLocation(bindMethod string) string {
	switch bindMethod {
	case "Query":
		return "query"
	case "JSON":
		return "body"
	default:
		return ""
	}
}
