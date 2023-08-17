package enums

import (
	"errors"
	"fmt"

	"github.com/nft-rainbow/conflux-gin-helper/utils"
)

type SettleType uint

const (
	SETTLE_TYPE_BALANCE SettleType = iota
	SETTLE_TYPE_QUOTA_RESET
	SETTLE_TYPE_QUOTA_ROLLOVER
)

func GetSettleType(isTestnet bool, method string, path string) SettleType {
	if !isTestnet {
		if utils.IsMint(method, path) {
			return SETTLE_TYPE_QUOTA_RESET
		}
		if utils.IsDeploy(method, path) {
			return SETTLE_TYPE_QUOTA_ROLLOVER
		}
	}
	return SETTLE_TYPE_BALANCE
}

var (
	SettleTypeValue2StrMap map[SettleType]string
	SettleTypeStr2ValueMap map[string]SettleType
)

var (
	ErrUnkownSettleType = errors.New("unknown settle type")
)

func init() {
	SettleTypeValue2StrMap = map[SettleType]string{
		SETTLE_TYPE_BALANCE:        "balance",
		SETTLE_TYPE_QUOTA_RESET:    "reset",
		SETTLE_TYPE_QUOTA_ROLLOVER: "rollover",
	}

	SettleTypeStr2ValueMap = make(map[string]SettleType)
	for k, v := range SettleTypeValue2StrMap {
		SettleTypeStr2ValueMap[v] = k
	}
}

func (t SettleType) String() string {
	v, ok := SettleTypeValue2StrMap[t]
	if ok {
		return v
	}
	return "unkown"
}

func (t SettleType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *SettleType) UnmarshalText(data []byte) error {
	v, ok := SettleTypeStr2ValueMap[string(data)]
	if ok {
		*t = v
		return nil
	}
	return fmt.Errorf("unknown settle type %v", string(data))
}

func ParseSettleType(str string) (*SettleType, error) {
	v, ok := SettleTypeStr2ValueMap[str]
	if !ok {
		return nil, fmt.Errorf("unknown settle type %v", str)
	}
	return &v, nil
}
