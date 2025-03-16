package gin

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type GroupInfo struct {
	Path   string
	Routes []string
}
type RouterInfo struct {
	Path   string
	Method string
}

func (r *RouterInfo) BuildPath() string {
	return fmt.Sprintf("@Router %s [%s]", r.Path, r.Method)
}
func GetRouterInfo(fNode *ast.FuncDecl) map[string]*RouterInfo {
	if isGinRouterFunc(fNode) {
		// 提取函数体中的路由信息
		return extractRoutesFromFuncBody(fNode.Body)
	}
	return nil
}

// 检查函数参数类型是否为 gin.Engine 或 gin.RouterGroup
func isGinRouterFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return false
	}
	// 获取第一个参数的类型
	paramType := fn.Type.Params.List[0].Type
	switch t := paramType.(type) {
	case *ast.StarExpr: // 指针类型，如 *gin.RouterGroup
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			return isGinType(sel)
		}
	case *ast.SelectorExpr: // 非指针类型，如 gin.Engine
		return isGinType(t)
	}
	return false
}

// 检查类型是否为 gin.Engine 或 gin.RouterGroup
func isGinType(sel *ast.SelectorExpr) bool {
	if x, ok := sel.X.(*ast.Ident); ok && x.Name == "gin" {
		return sel.Sel.Name == "Engine" || sel.Sel.Name == "RouterGroup"
	}
	return false
}

// 从函数体中提取路由信息
func extractRoutesFromFuncBody(body *ast.BlockStmt) map[string]*RouterInfo {

	groups, routerInfo := make(map[string]*GroupInfo), make(map[string]*RouterInfo) // 存储每个 Group 的信息
	var currentGroupID = "default"                                                  // 默认 Group 的标识符

	// 初始化默认 Group
	groups[currentGroupID] = &GroupInfo{
		Path:   "/",
		Routes: []string{},
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt: // 查找 Group 赋值语句
			if len(x.Lhs) == 1 && len(x.Rhs) == 1 {
				if call, ok := x.Rhs[0].(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Group" {
						// 提取 Group 路径
						groupPath := extractStringLiteral(call.Args[0])
						// 生成 Group 标识符
						currentGroupID = fmt.Sprintf("group%d", len(groups))
						// 存储 Group 信息
						groups[currentGroupID] = &GroupInfo{
							Path:   groupPath,
							Routes: []string{},
						}
					}
				}
			}

		case *ast.CallExpr: // 查找路由注册方法
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && isRouterMethod(sel.Sel.Name) {
				// 提取路由路径和方法
				path := extractStringLiteral(x.Args[0])
				method := sel.Sel.Name

				funcName := extractMethodName(x.Args[1])

				// 构建完整路径
				fullPath := buildFullPath(groups[currentGroupID].Path, path)
				r := &RouterInfo{
					Path:   fullPath,
					Method: method,
				}
				routerInfo[funcName] = r
				// 将路由信息添加到当前 Group
				//groups[currentGroupID].Routes = append(groups[currentGroupID].Routes, fmt.Sprintf("%s -> %s", fullPath, method))

			}
		}
		return true
	})

	// 输出结果
	return routerInfo
}

// 判断是否是路由注册方法
func isRouterMethod(name string) bool {
	routerMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, method := range routerMethods {
		if name == method {
			return true
		}
	}
	return false
}

// 提取字符串字面量（路由路径）
func extractStringLiteral(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return strings.Trim(lit.Value, `"`)
	}
	return ""
}

// 提取方法名
func extractMethodName(expr ast.Expr) string {
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		//return fmt.Sprintf("%s.%s", extractTypeName(sel.X), sel.Sel.Name)
		return sel.Sel.Name
	}
	return ""
}

// 提取类型名
func extractTypeName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", extractTypeName(x.X), x.Sel.Name)
	default:
		return "unknown"
	}
}

// 构建完整路径
func buildFullPath(groupPath, path string) string {
	p := strings.TrimSuffix(groupPath, "/") + "/" + strings.TrimPrefix(path, "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
