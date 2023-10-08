package models

import (
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
)

type UserApiQuota struct {
	BaseModel
	UserId             uint           `json:"user_id"`
	CostType           enums.CostType `json:"cost_type"`
	CountReset         int            `json:"count_reset"`           // 会在指定重置时间后重置
	NextResetCountTime time.Time      `json:"next_reset_count_time"` // 下一次重置时间
	CountRollover      int            `json:"count_rollover"`        // 下个月顺延
}

func (u *UserApiQuota) Total() int {
	if u == nil {
		return 0
	}
	return u.CountReset + u.CountRollover
}


