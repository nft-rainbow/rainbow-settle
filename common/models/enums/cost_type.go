package enums

import (
	"errors"
	"fmt"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
)

type CostType uint

const (
	COST_TYPE_RAINBOW_FREE CostType = iota + 1
	COST_TYPE_RAINBOW_NORMAL
	COST_TYPE_RAINBOW_MINT
	COST_TYPE_RAINBOW_DEPLOY
	COST_TYPE_CONFURA_NOMRAL = 10
	COST_TYPE_SCAN_NORMAL    = 20
)

func GetCostType(isTestnet bool, method string, path string) CostType {
	if !isTestnet {
		if utils.IsMint(method, path) {
			return COST_TYPE_RAINBOW_MINT
		}
		if utils.IsDeploy(method, path) {
			return COST_TYPE_RAINBOW_DEPLOY
		}
	}
	return COST_TYPE_RAINBOW_NORMAL
}

var (
	CostTypeValue2StrMap map[CostType]string
	CostTypeStr2ValueMap map[string]CostType
)

var (
	ErrUnkownCostType = errors.New("unknown cost type")
)

func init() {
	CostTypeValue2StrMap = map[CostType]string{
		COST_TYPE_RAINBOW_FREE:   "rainbow_free",
		COST_TYPE_RAINBOW_NORMAL: "rainbow_normal",
		COST_TYPE_RAINBOW_MINT:   "rainbow_mint",
		COST_TYPE_RAINBOW_DEPLOY: "rainbow_deploy",
		COST_TYPE_CONFURA_NOMRAL: "confura_normal",
		COST_TYPE_SCAN_NORMAL:    "scan_normal",
	}

	CostTypeStr2ValueMap = make(map[string]CostType)
	for k, v := range CostTypeValue2StrMap {
		CostTypeStr2ValueMap[v] = k
	}
}

func (t CostType) String() string {
	v, ok := CostTypeValue2StrMap[t]
	if ok {
		return v
	}
	return "unkown"
}

func (t CostType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *CostType) UnmarshalText(data []byte) error {
	v, ok := CostTypeStr2ValueMap[string(data)]
	if ok {
		*t = v
		return nil
	}
	return fmt.Errorf("unknown cost type %v", string(data))
}

func ParseCostType(str string) (*CostType, error) {
	v, ok := CostTypeStr2ValueMap[str]
	if !ok {
		return nil, fmt.Errorf("unknown cost type %v", str)
	}
	return &v, nil
}
