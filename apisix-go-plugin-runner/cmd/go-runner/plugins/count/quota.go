package count

import "github.com/nft-rainbow/rainbow-fiat/common/models/enums"

var (
	quotaLimit map[enums.CostType]int
)

func InitQuotaLimit() {
	quotaLimit = make(map[enums.CostType]int)
	quotaLimit[enums.COST_TYPE_RAINBOW_NORMAL] = 10000
	quotaLimit[enums.COST_TYPE_RAINBOW_MINT] = 200
	quotaLimit[enums.COST_TYPE_RAINBOW_DEPLOY] = 200
}
