package xerr

// OK 成功返回
const OK uint32 = 200

// 全局错误码

const ServerCommonError uint32 = 500
const RequestParamError uint32 = 400
const NotLoginError uint32 = 401
const UnauthorizedError uint32 = 403
const NotFoundError uint32 = 404

/**
 * 自定义错误码
 *
 * 规则：模块代码（2位） 功能代码（2位）具体错误码（4位）
 * 总长度8位：例如10010001，10 01 0001（10:模块代码（10-99），01:功能代码(01-99)，0001:具体错误码（0001-9999））
 */

const CommonLockError uint32 = 10000001
const RecordNotFoundError uint32 = 20000001
const RecordNotOperationError uint32 = 20000002
