package client

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	utils "github.com/JamesHertz/ipfs-client/utils"
	recs "github.com/JamesHertz/webmaster/record"
	cidlib "github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

func TestSuitable(t *testing.T) {

	suitable := []string{
		"/ip4/10.0.0.1/tcp/9000/p2p/customPID",
		"/ip4/10.0.0.1/udp/4001/p2p/customPID",
		"/ip4/1.2.3.4/udp/5000/quic-v1/p2p/customPID",
		"/ip4/3.6.7.9/udp/5001/quic/p2p/customPID",
	}

	unsuitable := []string{
		"/ip4/127.0.0.1/tcp/9000/p2p/customPID",
		"/ip4/127.0.0.1/udp/4001/p2p/customPID",
		"/ip4/127.0.0.1/udp/5000/quic-v1/p2p/customPID",
		"/ip4/127.0.0.1/udp/5001/quic/p2p/customPID",
		"/ip4/10.0.0.1/tcp/5000/webtransport/certhash/customhash",
		"/ip4/172.18.0.7/tcp/9000/p2p/customPID",
		"/ip4/172.18.0.250/tcp/9000/p2p/customPID",
	}

	for _, st := range suitable {
		require.True(t, suitableMultiAddrs(st), "addr: %v", st)
	}

	for _, ust := range unsuitable {
		require.False(t, suitableMultiAddrs(ust))
	}

}

func TestFindProvidersAndProvide(t *testing.T) {
	t.Log("THIS TESTS ASSUMES THERE HAVE 2 IPFS NODES RUNNING THAT ARE CONNECTED TO ONE ANOTHER.")
	time.Sleep(20 * time.Second)

	ipfs, err := NewClient(Mode(recs.NONE))
	require.Nil(t, err, "Error intializing client")

	for i := 0; i < 10; i++ {
		content := bytes.NewBuffer(
			[]byte(fmt.Sprintf("ipfs-client-running-%d", i)),
		)

		cid, err := ipfs.Add(content)
		require.Nil(t, err, "Couldn't add a new file: %d", i)

		t.Logf("cid: %s", cid)

		provs, err := ipfs.FindProviders(utils.CidInfo{Cid: cid})
		require.Nil(t, err, "If this one fails it may be because of the time it waited for the node to start or because the CID is not longer provided.")

		require.True(t, len(provs) > 0)

		require.Nil(t, ipfs.Unpin(cid), "Unable to remove previously added cid")
	}

	var cids []string
	for i := 0; i < 10; i++ {
		mh, err := multihash.Sum(
			[]byte(fmt.Sprintf("")), multihash.SHA2_256, -1,
		)
		require.Nil(t, err)

		cid := cidlib.NewCidV0(mh).String()
		cids = append(cids, cid)
		_, err = ipfs.Provide(utils.CidInfo{Cid: cid})
		require.Nil(t, err)
	}

	for _, cid := range cids {
		peers, err := ipfs.FindProviders(utils.CidInfo{Cid: cid})
		require.Nil(t, err)
		require.Equal(t, 1, len(peers))
	}
}
