package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"math/rand"

	"io/ioutil"

	recs "github.com/JamesHertz/webmaster/record"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	MultiAddrMatcher = regexp.MustCompile(
		"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$",
	)
	Localhost = "127.0.0.1"
)

var (
	WebmasterBaseUrl = "http://webmaster/%s"
	CidEndpoint      = fmt.Sprintf(WebmasterBaseUrl, "cids")
	PeersEndpoint    = fmt.Sprintf(WebmasterBaseUrl, "peers")

	// content type I will use :)
	ContentTypeJSON  = "application/json"
)

var ErrNoAddrFound   = errors.New("No addrs found")

const InterResolveTimeout = 10 * time.Second

type IpfsClientNode struct {
	*shell.Shell
	mode recs.IpfsMode
}

func NewClient(mode recs.IpfsMode) IpfsClientNode {
	return IpfsClientNode{
		Shell: shell.NewShell("localhost:5001"),
		mode:  mode,
	}
}

func (ipfs *IpfsClientNode) BootstrapNode() error {
	addrs, err := ipfs.getSuitableAddress()
	if err != nil {
		return err
	}

	// expect len(myaddrs) > 0
	data, _ := json.Marshal(addrs)
	res, err := http.Post(
		PeersEndpoint,
		ContentTypeJSON,
		bytes.NewBuffer(data),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Request error: %s", res.Status)
	}

	data, _ = io.ReadAll(res.Body)
	var bootstraps []peer.AddrInfo

	json.Unmarshal(data, &bootstraps)

	if len(bootstraps) != 0 {
		addrs := make([]string, len(bootstraps))
		// for each node return choose one of its random addr
		// and them you are go to
		for i, pi := range bootstraps {
			choosen := pi.Addrs[ rand.Intn(len(pi.Addrs)) ]
			addrs[i] = fmt.Sprintf("%s/p2p/%s", choosen, pi.ID.Pretty())
		}

		_, err = ipfs.BootstrapAdd(addrs)
	} // else: ops I was the first node :)

	return err
}

func (ipfs *IpfsClientNode) UploadFiles() error {
	files_dir := os.Getenv("FILES_DIR")
	files, _ := ioutil.ReadDir(files_dir)

	var cids []recs.CidRecord

	for _, file := range files {
		if file.Mode().IsRegular() {
			full_file_name := fmt.Sprintf("%s/%s", files_dir, file.Name())
			file_reader, _ := os.Open(full_file_name)

			log.Printf("Adding file %s to ipfs", full_file_name)
			cid, err := ipfs.Add(file_reader)
			if err != nil {
				return err // :(
			}

			if ipfs.shouldPublish() {
				rec, _ := recs.NewCidRecord(cid, ipfs.mode)
				cids = append(cids, *rec)
			}

		}
	}

	if len(cids) > 0 {
		log.Println("Uploading cids to webmaster")
		data, _ := json.Marshal(cids)
		res, err := http.Post(
			CidEndpoint,
			ContentTypeJSON,
			bytes.NewBuffer(data),
		)

		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("Request error: %s", res.Status)
		}

		log.Printf("%d cids uploaded", len(cids))
	}

	return nil
}

func (ipfs *IpfsClientNode) RunExperiment(ctx context.Context) error {
	log.Println("Starting experiment...")

	var (
		cids []recs.CidRecord
		err    error
	)

	for {
		if len(cids) == 0 {
			cids, err = ipfs.pullCids()
			if err != nil {
				return err
			}
		}

		next := cids[0]
		cids  = cids[1:]

		log.Printf("Resolving { CID: %s ; type: %s }", next.Cid, next.ProviderType)

		res, err := ipfs.DhtFindProvs(next.Cid.String())

		if err != nil {
			return err
		}

		if len(res) > 0 {
			log.Printf("Found %d providers.", len(res))
		} else {
			log.Printf("Unable to resolve CID")
		}

		// wait for some time or returns
		// because the experience is over
		select {
		case <- ctx.Done():
			return ctx.Err()
		case <- time.After(InterResolveTimeout):
		}
	}

}

func (ipfs *IpfsClientNode) DhtFindProvs(cid string) ([]shell.PeerInfo, error) {
	var peers struct{ Responses []shell.PeerInfo }
	// todo: ask about this...
	req := ipfs.Request("dht/findprovs", cid).Option("verbose", false).Option("num-providers", 1) 
	return peers.Responses, req.Exec(context.Background(), &peers)
}

func (ipfs *IpfsClientNode) getSuitableAddress() (*peer.AddrInfo, error) {
	pi, err := ipfs.ID()
	if err != nil {
		return nil, err
	}

	myaddrs := peer.AddrInfo{}

	for _, addr := range pi.Addresses {
		if suitableMultiAddress(addr) {
			aux, _ :=  peer.AddrInfoFromString(addr)
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
func suitableMultiAddress(maddr string) bool {
	res := MultiAddrMatcher.FindStringSubmatch(maddr)
	return res != nil && res[1] != Localhost
}

func (ipfs *IpfsClientNode) shouldPublish() bool {
	return ipfs.mode != recs.NONE
}

func (ipfs *IpfsClientNode) pullCids() ([]recs.CidRecord, error) {
	var records []recs.CidRecord
	res, err := http.Get(CidEndpoint)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request error: %s", res.Status)
	}

	data, _ := io.ReadAll(res.Body)

	return records, json.Unmarshal(data, &records) 

}