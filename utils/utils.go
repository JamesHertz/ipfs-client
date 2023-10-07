package utils

import "fmt"

// some usefult constants for testint sake
type CIDType int

const (
	Normal CIDType = iota + 1
	Secure
)

const (
	CidTypeOptionName = "cid-type"
)

type CidInfo struct {
	Content string
	CidType CIDType 
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