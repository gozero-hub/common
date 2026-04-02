package mail

import (
	"fmt"

	"gitee.com/scholar-hub/go-zero-common/xerr"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type VerifyParam struct {
	Redis *redis.Redis
	Email string
	Code  string
}

func VerifyEmailCode(v VerifyParam) (err error) {
	key := fmt.Sprintf("sendCodeEmail:%s", v.Email)
	storedCode, err := v.Redis.Get(key)
	if err != nil {
		return err
	}
	if storedCode != v.Code {
		return xerr.NewErrMsg("incorrect verification code")
	}

	// 验证成功，删除验证码
	_, _ = v.Redis.Del(key)
	return nil
}
