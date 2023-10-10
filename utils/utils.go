package utils

import (
	"encoding/json"
	"fmt"
)

// some usefult constants for testint sake
type CIDType int

const (
	Normal CIDType = iota
	Secure
)

// TODO: rename to CidRecord
// TODO: think about the DEFAULT type and find a better solution for this
type CidInfo struct {
	Cid  string
	Type CIDType
}

func (cidType CIDType) String() string {
	switch cidType {
	case Normal:
		return "Normal"
	case Secure:
		return "Secure"
	default:
		panic(
			fmt.Sprintf("Invalid CIDType %d", cidType),
		)
	}
}

func (info *CidInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(info.Cid)
}