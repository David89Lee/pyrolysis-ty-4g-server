package response

var MsgFlags = map[int]string{
	SUCCESS:                    "ok",
	ERROR:                      "fail",
	ServerError:                "服务端错误",
	InvalidParams:              "请求参数错误",
	ErrorAuthCheckTokenFail:    "token不合法",
	ErrorAuthCheckTokenTimeout: "token超时",
	ErrorGenerateToken:         "生成token失败",
	ErrorConnectWebSocket:      "连接websocket失败",
	ErrorReadWebSocket:         "读取websocket失败",

	ErrorCreateInfo:     "创建信息失败",
	ErrorUpdateInfo:     "更新信息失败",
	ErrorDeleteInfo:     "删除信息失败",
	ErrorFindInfo:       "查询信息失败",
	ErrorFindDtuRecords: "查询DTU记录失败",
	ErrorFindDtuHistory: "查询DTU历史数据失败",
	ErrorFindDtuCurve:   "查询DTU曲线数据失败",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
