package client

import (
	"context"
	"errors"
	"regexp"
	"time"

	recs "github.com/JamesHertz/webmaster/record"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p/core/peer"

	. "github.com/JamesHertz/ipfs-client/utils"
)

var (
	MultiAddrMatcher = regexp.MustCompile(
		"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$",
	)
	Localhost = "127.0.0.1"
)

var ErrNoAddrFound = errors.New("No addrs found")

// TODO: add this as a config option
const InterResolveTimeout = 10 * time.Second

type IpfsClientNode struct {
	*shell.Shell
	mode         recs.IpfsMode
	nodeSeqNum   int
	cidsCount    int
	localCids    map[string]bool
}

func NewClient(opts ...Option) (*IpfsClientNode, error) {
	var err error

	cfg := defaultConfig()
	if err = cfg.Apply(opts...); err != nil {
		return nil, err
	}

	ipfs := IpfsClientNode{
		Shell:           shell.NewShell(cfg.apiUrl),
		mode:            cfg.mode,
		localCids:       make(map[string]bool),
	}

	if len(cfg.bootstrapNodes) > 0 {
		_, err = ipfs.BootstrapAdd(
			cfg.bootstrapNodes,
		)
	}

	return &ipfs, err
}

func (ipfs *IpfsClientNode) FindProviders(cid CidInfo) ([]shell.PeerInfo, error) {
	var peers struct{ Responses []shell.PeerInfo }
	req := ipfs.Request("dht/findprovs", cid.Content).Option("verbose", false).Option("num-providers", 1)
	req.Option("cidtype", cid.CidType.String())
	return peers.Responses, req.Exec(context.Background(), &peers)
}

func (ipfs *IpfsClientNode) Provide(cid CidInfo) ([]shell.PeerInfo, error) {
	var peers struct{ Responses []shell.PeerInfo }
	req := ipfs.Request("dht/provide", cid.Content).Option("verbose", false).Option("recursive", false)
	return peers.Responses, req.Exec(context.Background(), &peers)
}

func (ipfs *IpfsClientNode) SuitableAddresses() (*peer.AddrInfo, error) {
	pi, err := ipfs.ID()
	if err != nil {
		return nil, err
	}

	myaddrs := peer.AddrInfo{}

	for _, addr := range pi.Addresses {
		if suitableMultiAddrs(addr) {
			aux, _ := peer.AddrInfoFromString(addr)
			myaddrs.ID = aux.ID
			myaddrs.Addrs = append(myaddrs.Addrs, aux.Addrs...)
		}
	}

	if len(myaddrs.Addrs) == 0 {
		return nil, ErrNoAddrFound
	}

	return &myaddrs, nil
}

// address that does not have the localhost as ip and that
// aren't address for webtransport or webrtc stuffs
func suitableMultiAddrs(maddr string) bool {
	res := MultiAddrMatcher.FindStringSubmatch(maddr)
	return res != nil && res[1] != Localhost
}