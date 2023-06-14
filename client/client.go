package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"math/rand"

	"io/ioutil"

	recs "github.com/JamesHertz/webmaster/record"
	shell "github.com/ipfs/go-ipfs-api"
)

var (
	RE = regexp.MustCompile(
		"^/ip4/([.0-9]+)/(tcp|udp)/\\d+(/quic-v1|/quic)?/p2p/\\w+$",
	)
	LOCALHOST = "127.0.0.1"
)

var (
	SERVER_BASE_URL = "http://webmaster/%s"
	CIDS_URL        = fmt.Sprintf(SERVER_BASE_URL, "cids")
	PEERS_URL       = fmt.Sprintf(SERVER_BASE_URL, "peers")

	// content types I will use :)
	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain; charset=utf-8"
)

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

func (ipfs *IpfsClientNode) RunExperiment() error {
	log.Println("Starting experiment...")


	return nil
}

func (ipfs * IpfsClientNode) pullCids() ([]string, error){
	return nil, nil
}

func (ipfs * IpfsClientNode) DhtResolve(cid string) ([]shell.PeerInfo, error){
	var peers struct{ Responses []shell.PeerInfo }
	req := ipfs.Request("dht/findprovs", cid).Option("verbose", false).Option("num-providers", 1)
	return peers.Responses, req.Exec(context.Background(), &peers)
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
		log.Println("Uploading files to webmaster")
		data, err := json.Marshal(cids)
		if err != nil {
			log.Fatal("Unable to marshal CidRecords")
		}
		res, err := http.Post(
			CIDS_URL,
			ContentTypeJSON,
			bytes.NewBuffer(data),
		)

		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("Request error: %s", res.Status)
		}

		log.Printf("%d files uploaded", len(cids))
	}

	return nil
}

func (ipfs *IpfsClientNode) BootstrapNode() error {
	addrs, err := ipfs.GetSuitableAddress()
	if err != nil {
		return err
	}

	// expect len(myaddrs) > 0
	chosen_addr := addrs[rand.Intn(len(addrs))]
	res, err := http.Post(
		PEERS_URL,
		ContentTypeText,
		bytes.NewBuffer([]byte(chosen_addr)),
	)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Request error: %s", res.Status)
	}

	data, _ := io.ReadAll(res.Body)
	var bootstraps []string

	json.Unmarshal(data, &bootstraps)

	if len(bootstraps) != 0 {
		_, err = ipfs.BootstrapAdd(bootstraps)
	} // else: ops I was the first node :)

	return err
}

func (ipfs *IpfsClientNode) GetSuitableAddress() ([]string, error) {
	pi, err := ipfs.ID()
	if err != nil {
		return nil, err
	}

	myaddrs := []string{}

	for _, addr := range pi.Addresses {
		if suitableMultiAddress(addr) {
			myaddrs = append(myaddrs, addr)
		}
	}

	return myaddrs, nil
}

func (ipfs *IpfsClientNode) shouldPublish() bool {
	return ipfs.mode != recs.NONE
}

// address that does not have the localhost as ip
// aren't address for webtransport or webrtc stuffs
func suitableMultiAddress(maddr string) bool {
	res := RE.FindStringSubmatch(maddr)
	return res != nil && res[1] != LOCALHOST
}
