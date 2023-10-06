package utils

type CIDType int

const (
	Normal CIDType = iota 
	Secure
)

type CidInfo struct {
	Content string
	CidType CIDType 
}