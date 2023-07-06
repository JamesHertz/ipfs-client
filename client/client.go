package client

import (
	// "bytes"
	// "encoding/json"
	// "fmt"
	// "io"
	// "net/http"
	// "os"

	"context"
	"errors"
	"log"
	"regexp"
	"time"

	// "math/rand"

	recs "github.com/JamesHertz/webmaster/record"
	shell "github.com/ipfs/go-ipfs-api"
	// utils "github.com/ipfs/kubo/cmd/ipfs/util"

	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	MultiAddrMatcher = regexp.MustCompile(
		"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$",
	)
	Localhost = "127.0.0.1"
)

// var (
// 	WebmasterBaseUrl = "http://webmaster/%s"
// 	CidEndpoint      = fmt.Sprintf(WebmasterBaseUrl, "cids")
// 	PeersEndpoint    = fmt.Sprintf(WebmasterBaseUrl, "peers")

// 	// content type I will use :)
// 	ContentTypeJSON = "application/json"
// )

var ErrNoAddrFound = errors.New("No addrs found")

const InterResolveTimeout = 10 * time.Second

type IpfsClientNode struct {
	*shell.Shell
	mode            recs.IpfsMode
	localCids       map[string]bool
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

func (ipfs *IpfsClientNode) UploadCids() error {
	// files_dir := os.Getenv("FILES_DIR")
	// files, _ := ioutil.ReadDir(files_dir)

	// cidsLog := utils.NewLogger("cids.log")
	// var cids []recs.CidRecord

	// for _, file := range files {
	// 	if file.Mode().IsRegular() {
	// 		full_file_name := fmt.Sprintf("%s/%s", files_dir, file.Name())
	// 		file_reader, _ := os.Open(full_file_name)

	// 		cid, err := ipfs.Add(file_reader)
	// 		if err != nil {
	// 			return err // :(
	// 		}

	// 		log.Printf("File %s [ CID: %s ] added", full_file_name, cid)
	// 		ipfs.localCids[cid] = true
	// 		if ipfs.bootstrapable() {
	// 			rec, _ := recs.NewCidRecord(cid, ipfs.mode)
	// 			cidsLog.Printf(`{"cid": "%s", "type": "%s"}`, rec.Cid, rec.ProviderType)
	// 			cids = append(cids, *rec)
	// 		}

	// 	}
	// }

	// if len(cids) > 0 {
	// 	log.Println("Uploading cids to webmaster")
	// 	data, _ := json.Marshal(cids)
	// 	res, err := http.Post(
	// 		CidEndpoint,
	// 		ContentTypeJSON,
	// 		bytes.NewBuffer(data),
	// 	)

	// 	if err != nil {
	// 		return err
	// 	}

	// 	defer res.Body.Close()

	// 	if res.StatusCode != http.StatusOK {
	// 		return fmt.Errorf("Request error: %s", res.Status)
	// 	}
	// 	// report added cids :)
	// 	log.Printf("%d cids uploaded", len(cids))
	// }

	return nil
}

func (ipfs *IpfsClientNode) RunExperiment(ctx context.Context) error {
	log.Println("Starting experiment...")

	log.Println("Waiting 1 minute...")
	time.Sleep(time.Minute * 1)

	if err := ipfs.UploadCids(); err != nil {
		log.Fatalf("Error uploading cids: %v", err)
	}

	log.Println("Uploading complete. Waiting 30 seconds")
	time.Sleep(time.Second * 30)

	// start resolving :)
	var (
		cids []string
		// cids []recs.CidRecord
		// err  error
	)

	for {
		if len(cids) == 0 {
			// TODO: change this :)
			// cids, err = ipfs.pullCids()
			// if err != nil {
			// 	return err
			// }
		}


		next := cids[0]
		cids = cids[1:]

		if ipfs.localCids[next] {
			log.Printf("Ups one of my CID came to scare me: %s ", next)
			continue
		}

		// log.Printf("Resolving { CID: %s ; type: %s }", next.Cid, next.ProviderType)
		log.Printf("Resolving CID: %s", next)

		res, err := ipfs.FindProviders(next)

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
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return nil
			}
			return ctx.Err()
		case <-time.After(InterResolveTimeout):
		}
	}

}

func (ipfs *IpfsClientNode) FindProviders(cid string) ([]shell.PeerInfo, error) {
	var peers struct{ Responses []shell.PeerInfo }
	req := ipfs.Request("dht/findprovs", cid).Option("verbose", false).Option("num-providers", 1)
	return peers.Responses, req.Exec(context.Background(), &peers)
}

func (ipfs *IpfsClientNode) Provide(cid string) ([]shell.PeerInfo, error) {
	var peers struct{ Responses []shell.PeerInfo }
	req := ipfs.Request("dht/provide", cid).Option("verbose", false).Option("recursive", false)
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

func (ipfs *IpfsClientNode) pullCids() ([]recs.CidRecord, error) {
	// var records []recs.CidRecord
	// res, err := http.Get(CidEndpoint)

	// if err != nil {
	// 	return nil, err
	// }

	// defer res.Body.Close()

	// if res.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("Request error: %s", res.Status)
	// }

	// data, _ := io.ReadAll(res.Body)

	// return records, json.Unmarshal(data, &records)

	return nil, nil
}
