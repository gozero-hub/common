package tool

import (
	"context"
	"fmt"

	"gitee.com/scholar-hub/go-zero-common/xerr"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Lock struct {
	Rds *redis.Redis
}

// TryLock 注册锁
// key: 锁唯一标识
// ttl: 锁过期时间
// 返回 true 表示获取锁成功，false 表示锁已存在
func (l *Lock) TryLock(ctx context.Context, key string, ttl int) (err error) {
	ok, err := l.Rds.SetnxExCtx(ctx, key, "1", ttl)
	if err != nil {
		return xerr.NewErrCode(xerr.ServerCommonError)
	}
	if !ok {
		return xerr.NewErrCode(xerr.CommonLockError)
	}
	return nil
}

// Unlock 释放锁
func (l *Lock) Unlock(ctx context.Context, key string) {
	_, _ = l.Rds.DelCtx(ctx, key)
}

// LimitParam 限流检查参数
type LimitParam struct {
	Ctx     context.Context
	Redis   *redis.Redis
	Key     string // 唯一 key，比如 sendcode:ip:register:127.0.0.1
	Limit   int    // 最大次数
	Expired int    // 过期时间（秒），如果为 0，默认当天结束
}

// CheckLimit 检查某个 key 在过期时间内是否超过限制
func CheckLimit(param LimitParam) error {
	count, err := param.Redis.IncrCtx(param.Ctx, param.Key)
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	// 第一次设置过期时间
	if count == 1 {
		expire := param.Expired
		if expire == 0 {
			expire = GetTodayExpireAt()
		}
		_ = param.Redis.ExpireCtx(param.Ctx, param.Key, expire)
	}

	// 请求限制
	limit := param.Limit
	if limit == 0 {
		limit = 10
	}
	if count > int64(param.Limit) {
		return fmt.Errorf("the request limit has been exceeded")
	}
	return nil
}

// CheckLimitIp 专门的 IP 限制方法
func CheckLimitIp(ctx context.Context, rds *redis.Redis, vType string) error {
	ip := GetIPFromCtx(ctx)
	if ip == "" {
		return xerr.NewErrMsg("unable to obtain the requested IP address")
	}

	// 构造限制 key
	limitKey := fmt.Sprintf("CheckLimitIp:%s:%s", vType, ip)

	// 调用通用限流方法
	return CheckLimit(LimitParam{
		Ctx:   ctx,
		Redis: rds,
		Key:   limitKey,
		Limit: 10, // 每天最多 10 次
	})
}
