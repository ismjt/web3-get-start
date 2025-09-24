package models

// 基础响应结构体
type BaseResponse struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
}

// 包含单条数据
type DataResponse[T any] struct {
	BaseResponse
	Payload T `json:"payload"`
}

// 包含分页信息
type PageResponse[T any] struct {
	BaseResponse
	Payload []T   `json:"payload"` // 列表数据
	Total   int64 `json:"total"`   // 总条数
	Current int64 `json:"current"` // 当前页
	Size    int64 `json:"size"`    // 每页条数
}

// 登录成功返回数据体
type LoginData struct {
	Token     string `json:"token"`      // JWT
	ExpiresAt int64  `json:"expires_at"` // 过期时间（时间戳）
	UserID    uint   `json:"user_id"`    // 用户ID
	Username  string `json:"username"`   // 用户名
}
