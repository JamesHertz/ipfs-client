package experiments

import (
	"context"
	"github.com/JamesHertz/ipfs-client/client"
)

type Experiment interface {
	Start(ipfs *client.IpfsClientNode, ctx context.Context) error
}