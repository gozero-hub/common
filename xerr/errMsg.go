package xerr

var message map[uint32]string

func init() {
	message = make(map[uint32]string)
	message[OK] = "Success"
	message[ServerCommonError] = "the server is malfunctioning, please try again later"
	message[RequestParamError] = "parameter error"
	message[NotLoginError] = "not logged in, please login first"
	message[UnauthorizedError] = "unauthorized operation"
	message[NotFoundError] = "api not found"

	message[CommonLockError] = "the operation is too fast, please try again later"
	message[RecordNotFoundError] = "the record not found"
	message[RecordNotFoundError] = "the record not operation"
}

func MapErrMsg(errcode uint32) string {
	if msg, ok := message[errcode]; ok {
		return msg
	} else {
		return message[ServerCommonError]
	}
}

func IsCodeErr(errcode uint32) bool {
	if _, ok := message[errcode]; ok {
		return true
	} else {
		return false
	}
}
