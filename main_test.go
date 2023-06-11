package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)



func TestSuitable(t *testing.T){
	//"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$" 

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