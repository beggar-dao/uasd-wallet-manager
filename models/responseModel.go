package models

type BaseResonseModel struct {
	Code    int         `json:"code";form:"code"`       //状态
	Message string      `json:"message";form:"message"` // 描述
	Data    interface{} `json:"data";form:"data"`       //查询结果
}
