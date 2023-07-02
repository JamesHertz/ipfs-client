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
	"strings"
	"time"

	"math/rand"

	"io/ioutil"

	recs "github.com/JamesHertz/webmaster/record"
	shell "github.com/ipfs/go-ipfs-api"
	utils "github.com/ipfs/kubo/cmd/ipfs/util"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	MultiAddrMatcher = regexp.MustCompile(
		"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$",
	)
	Localhost  = "127.0.0.1"

	// this is because of docker swarm networks 
	// the containers that are created using such networks
	// always have an eth1 interace which is not reachable for the 
	// other containers and which ip range is 172.18.0.0/8 
	// so this was the quickest solution I found for this
	BadPreffix = "172.18.0." 
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
	localCids map[string] bool
	shouldBootstrap bool
}

func NewClient(opts ...Option) (*IpfsClientNode, error) {
	cfg := defaultConfig()

	if err := cfg.Apply(opts...) ; err != nil {
		return nil, err
	}
	return &IpfsClientNode{
		Shell: shell.NewShell(cfg.apiUrl),
		mode:  cfg.mode,
		localCids: make(map[string]bool),
		shouldBootstrap: cfg.shouldBootstrap,
	}, nil
}

func (ipfs *IpfsClientNode) BootstrapNode() error {
	addrs, err := ipfs.getSuitableAddress()
	if err != nil {
		return err
	}

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
			log.Printf("Connecting to: %s", pi.ID)
		}

		_, err = ipfs.BootstrapAdd(addrs)
	} else {
		log.Println("I was the first node")
	}

	return err
}

func (ipfs *IpfsClientNode) UploadFiles() error {
	files_dir := os.Getenv("FILES_DIR")
	files, _  := ioutil.ReadDir(files_dir)

	cidsLog   := utils.NewLogger("cids.log")
	var cids []recs.CidRecord

	for _, file := range files {
		if file.Mode().IsRegular() {
			full_file_name := fmt.Sprintf("%s/%s", files_dir, file.Name())
			file_reader, _ := os.Open(full_file_name)

			cid, err := ipfs.Add(file_reader)
			if err != nil {
				return err // :(
			}

			log.Printf("File %s [ CID: %s ] added", full_file_name, cid)
			ipfs.localCids[cid] = true
			if ipfs.bootstrapable() {
				rec, _ := recs.NewCidRecord(cid, ipfs.mode)
				cidsLog.Printf(`{"cid": "%s", "type": "%s"}`, rec.Cid, rec.ProviderType)
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
		// report added cids :)
		log.Printf("%d cids uploaded", len(cids))
	}

	return nil
}

func (ipfs *IpfsClientNode) RunExperiment(ctx context.Context) error {
	log.Println("Starting experiment...")

	// do some inital things :)
	if ipfs.bootstrapable() {
		if err := ipfs.BootstrapNode(); err != nil {
			return fmt.Errorf("Error bootstrap the client: %v", err)
		}
		log.Println("Bootstrap complete.")
	}

	log.Println("Waiting 1 minute...")
	time.Sleep(time.Minute * 1)

	if err := ipfs.UploadFiles(); err != nil {
		log.Fatalf("Error uploading files: %v", err)
	}

	log.Println("Uploading complete. Waiting 30 seconds")
	time.Sleep(time.Second * 30)

	// start resolving :)
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

		if ipfs.localCids[next.Cid.String()] {
			log.Printf("Ups one of my CID came to scare me: %s ", next.Cid)
			continue
		}

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
			if ctx.Err() == context.DeadlineExceeded {
				return nil
			}
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
	return res != nil && res[1] != Localhost && !strings.HasPrefix(res[1], BadPreffix)
}

func (ipfs *IpfsClientNode) bootstrapable() bool {
	return ipfs.shouldBootstrap
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