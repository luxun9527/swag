package request

type GetAccountInfoReq struct {
	Id string `form:"id"`
}
type UpdateAccountReq struct {
	Id   string `form:"id"`
	Name string `form:"name"`
}
