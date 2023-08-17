package enums

import (
	"errors"
	"fmt"
)

type ServerType uint

const (
	SERVER_TYPE_RAINBOW ServerType = iota + 1
	SERVER_TYPE_CONFURA
	SERVER_TYPE_SCAN
)

var (
	ServerTypeValue2StrMap map[ServerType]string
	ServerTypeStr2ValueMap map[string]ServerType
)

var (
	ErrUnkownServerType = errors.New("unknown server type")
)

func init() {
	ServerTypeValue2StrMap = map[ServerType]string{
		SERVER_TYPE_RAINBOW: "rainbow",
		SERVER_TYPE_CONFURA: "confura",
		SERVER_TYPE_SCAN:    "scan",
	}

	ServerTypeStr2ValueMap = make(map[string]ServerType)
	for k, v := range ServerTypeValue2StrMap {
		ServerTypeStr2ValueMap[v] = k
	}
}

func (t ServerType) String() string {
	v, ok := ServerTypeValue2StrMap[t]
	if ok {
		return v
	}
	return "unkown"
}

func (t ServerType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *ServerType) UnmarshalText(data []byte) error {
	v, ok := ServerTypeStr2ValueMap[string(data)]
	if ok {
		*t = v
		return nil
	}
	return fmt.Errorf("unknown server type %v", string(data))
}

func ParseServerType(str string) (*ServerType, error) {
	v, ok := ServerTypeStr2ValueMap[str]
	if !ok {
		return nil, fmt.Errorf("unknown server type %v", str)
	}
	return &v, nil
}
