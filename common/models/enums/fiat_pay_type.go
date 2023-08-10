package enums

import (
	"errors"
	"fmt"
)

/*
failed get sponsor status from db
sponsoring
io error
not enough cash
failed get sponsor info
sponsor not enough
*/

type UserPayType uint

const (
	USER_PAY_TYPE_PRE UserPayType = iota + 1
	USER_PAY_TYPE_POST
)

var (
	UserPayTypeValue2StrMap map[UserPayType]string
	UserPayTypeStr2ValueMap map[string]UserPayType
)

var (
	ErrUnkownUserPayType = errors.New("unknown user pay type")
)

func init() {
	UserPayTypeValue2StrMap = map[UserPayType]string{
		USER_PAY_TYPE_PRE:  "pre",
		USER_PAY_TYPE_POST: "post",
	}

	UserPayTypeStr2ValueMap = make(map[string]UserPayType)
	for k, v := range UserPayTypeValue2StrMap {
		UserPayTypeStr2ValueMap[v] = k
	}
}

func (t UserPayType) String() string {
	if t == 0 {
		return ""
	}

	v, ok := UserPayTypeValue2StrMap[t]
	if ok {
		return v
	}
	return "unkown"
}

func (t UserPayType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *UserPayType) UnmarshalText(data []byte) error {
	v, ok := UserPayTypeStr2ValueMap[string(data)]
	if ok {
		*t = v
		return nil
	}
	return fmt.Errorf("unknown user pay type %v", string(data))
}

func ParseFiatPayType(str string) (*UserPayType, error) {
	v, ok := UserPayTypeStr2ValueMap[str]
	if !ok {
		return nil, fmt.Errorf("unknown user pay type %v", str)
	}
	return &v, nil
}
