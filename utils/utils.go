package utils

import "fmt"

type CIDType int

const (
	Normal CIDType = iota 
	Secure
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