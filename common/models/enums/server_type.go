package enums

import (
	"errors"
	"fmt"
)

type ServerType uint

const (
	SERVER_TYPE_RAINBOW ServerType = iota + 1
	SERVER_TYPE_CONFURA_CSPACE
	SERVER_TYPE_CONFURA_ESPACE
	SERVER_TYPE_SCAN_CSPACE
	SERVER_TYPE_SCAN_ESPACE
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
		SERVER_TYPE_RAINBOW:        "rainbow",
		SERVER_TYPE_CONFURA_CSPACE: "confura_main_cspace",
		SERVER_TYPE_CONFURA_ESPACE: "confura_test_espace",
		SERVER_TYPE_SCAN_CSPACE:    "scan_main_cspace",
		SERVER_TYPE_SCAN_ESPACE:    "scan_main_espace",
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
