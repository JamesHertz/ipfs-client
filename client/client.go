package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"math/rand"

	"io/ioutil"

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
	CIDS_URL  = fmt.Sprintf(SERVER_BASE_URL, "cids")
	PEERS_URL = fmt.Sprintf(SERVER_BASE_URL, "peers")

	// content types I will use :)
    ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain; charset=utf-8"
)

type IpfsClientNode struct {
	*shell.Shell
}

func NewClient() IpfsClientNode {
	return IpfsClientNode{
		Shell: shell.NewShell("localhost:5001"),
	}
}

func (ipfs * IpfsClientNode) UploadFiles() error {
	files_dir := os.Getenv("FILES_DIR")
	files, _  := ioutil.ReadDir(files_dir)

	for _, file := range files {
		if file.Mode().IsRegular() {
			full_file_name := fmt.Sprintf("%s/%s", files_dir, file.Name())
			file_reader , _   := os.Open(full_file_name)

			// TODO: save the CID and then send it to the webmaster
			/*cid*/_, err := ipfs.Add(file_reader)
			if err != nil {
				return err // :(
			}
		}
	}

	return nil
}

func (ipfs * IpfsClientNode) BootstrapNode() error {
	addrs, err := ipfs.GetSuitableAddress()
	if err != nil {
		return err
	}

	// expect len(myaddrs) > 0
	chosen_addr := addrs[ rand.Intn( len(addrs) ) ]
	res, err := http.Post(
		PEERS_URL, 
		ContentTypeText, 
		bytes.NewBuffer([]byte(chosen_addr)),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
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

func (ipfs * IpfsClientNode) GetSuitableAddress() ([]string, error) {
	pi, err := ipfs.ID()
	if err != nil {
		return nil, err
	}

	myaddrs := []string{}

	for _, addr := range pi.Addresses {
		if suitableMultiAddress(addr){
			myaddrs = append(myaddrs, addr)
		}
	}

	return myaddrs, nil
}


// address that does not have the localhost as ip
// aren't address for webtransport or webrtc stuffs
func suitableMultiAddress(maddr string) bool{
	res := RE.FindStringSubmatch(maddr)
	return res != nil && res[1] != LOCALHOST
}

// notify server
