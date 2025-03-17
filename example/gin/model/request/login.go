package request

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
