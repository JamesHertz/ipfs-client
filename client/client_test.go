package client

import (
	"bytes"
	"context"
	"testing"
	"time"

	"os/exec"

	recs "github.com/JamesHertz/webmaster/record"
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
	}

	for _, st := range suitable {
		require.True(t, suitableMultiAddress(st), "addr: %v", st)
	}

	for _, ust := range unsuitable {
		require.False(t, suitableMultiAddress(ust))
	}

}

func TestDhtResolve(t *testing.T) {
	// this test assumes that you have ipfs installed in your machine
	_, err := exec.LookPath("ipfs")
	if err == nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cmd := exec.CommandContext(ctx, "ipfs", "daemon")
		err := cmd.Start()

		require.Nil(t, err, "Failed running daemon")

		// lets wait a bit for it to start
		time.Sleep(5 * time.Second)

		ipfs := NewClient(recs.NONE)
		content := bytes.NewBuffer([]byte("ipfs-client running :)"))

		cid, err := ipfs.Add(content)
		require.Nil(t, err, "Coudln't add a new file")

		t.Logf("cid: %s", cid)

		provs, err := ipfs.DhtFindProvs(cid)
		require.Nil(t, err, "If this one fails it may be because of the time it waited for the node to start or because the CID is not longer provided.")

		t.Logf("provs: %v", provs)

		require.Nil(t, ipfs.Unpin(cid), "Unable to remove previously added cid")
	} else {
		t.Log("Unable to find ipfs in your machine")
	}

}
