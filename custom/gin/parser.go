package gin

import (
	"golang.org/x/tools/go/packages"
	"log"
)

func ParseDetail(dir string) []*FileDetail {

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports,
		Dir: dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal("包加载失败:", err)
	}

	var (
		fdList     []*FileDetail
		routerInfo = map[string]*RouterInfo{}
	)
	for _, v := range pkgs {
		fileDetail, routerDetail := processPackage(v)
		for k, v1 := range routerDetail {
			routerInfo[k] = v1
		}
		fdList = append(fdList, fileDetail...)
	}
	for _, v := range fdList {
		for _, v1 := range v.FuncDetailList {
			r, ok := routerInfo[v1.FuncName]
			if ok {
				v1.Router = r.BuildPath()
			}
		}
	}
	return fdList
}
