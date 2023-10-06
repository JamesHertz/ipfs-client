package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JamesHertz/ipfs-client/client"

	utils "github.com/ipfs/kubo/cmd/ipfs/util"
	. "github.com/JamesHertz/ipfs-client/utils"
)

type ResolveExperiment struct {
	localCids 	  []CidInfo
	externalCids  []CidInfo

	mode string
}


// TODO: accept a ctx as argument wich is filled with the env value
func loadCids(filename string) ([]CidInfo, error){
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cids := strings.Split(
		strings.Trim(string(data), "\n"), "\n",
	)

	cidsInfo := make([]CidInfo, len(cids))

	for i, cid := range cids {
		var cidType CIDType
		// TODO: find a better solution for this
		if i < len(cids) / 2 {
			cidType = Normal
		} else {
			cidType = Secure
		}

		cidsInfo[i] = CidInfo{
			Content: cid,
			CidType: cidType,
		}
	}

	return cidsInfo, nil
}

const InterResolveTimeout = 10 * time.Second

func NewResolveExperiment() (Experiment, error) {
	node_seq_var      :=  os.Getenv("NODE_SEQ_NUM") 
	total_nodes_var   :=  os.Getenv("EXP_TOTAL_NODES") 
	cids_per_node_var :=  os.Getenv("EXP_CIDS_PER_NODE") 
	cids_file         :=  os.Getenv("EXP_CIDS_FILE")
	node_type         :=  os.Getenv("MODE")

	// ?? invalid mode?
	if node_type == "" {
		return nil, fmt.Errorf("MODE not set")
	}

	seqNum, err := strconv.Atoi( node_seq_var )
	if err != nil {
		return nil, fmt.Errorf("Error parsing NODE_SEQ_NUM (%s): %s", node_seq_var, err)
	}

	totalNodes, err := strconv.Atoi( total_nodes_var )
	if err != nil {
		return nil, fmt.Errorf("Error parsing EXP_TOTAL_NODES (%s): %s", total_nodes_var,  err)
	}

	cidsPerNode, err := strconv.Atoi( cids_per_node_var )
	if err != nil {
		return nil, fmt.Errorf("Error parsing CIDS_PER_NODE (%s): %s", cids_per_node_var, err)
	}

	// data, err := os.ReadFile(cids_file)
	all_cids, err := loadCids(cids_file)
	if err != nil {
		return nil, fmt.Errorf("Error get cids from EXP_CIDS_FILE (%s): %s", cids_file, err)
	}

	total_exp_cids := cidsPerNode * totalNodes

	if seqNum >= totalNodes {
		return nil, fmt.Errorf("Invalid NODE_SEQ_NUMBER %d it should be [0,%d[", seqNum, totalNodes)
	} 

	if len(all_cids) !=  total_exp_cids {
		return nil, fmt.Errorf("Expected %d cids but found %d nodes that.", total_exp_cids, len(all_cids))
	}

	// TODO: add tests for this thing
	if node_type == "Normal" {
		all_cids = all_cids[:total_exp_cids / 2]
	} 

	var externalCids, localCids []CidInfo

	start_cid := seqNum*cidsPerNode 
	end_cid   := start_cid + cidsPerNode

	localCids     = append(localCids, all_cids[start_cid:end_cid]...)
	externalCids  = append(externalCids, all_cids[:start_cid]...)
	externalCids  = append(externalCids, all_cids[end_cid:]...)


	fmt.Printf("localCids: %v\n", localCids)
	fmt.Printf("externalCids: %v\n", externalCids)

	return &ResolveExperiment{
		localCids: localCids,
		externalCids: externalCids,
	}, nil
}

	
func(exp *ResolveExperiment) Start(ipfs *client.IpfsClientNode, ctx context.Context) error {
	log.Println("Starting experiment...")

	log.Println("Waiting 5 minute...")
	time.Sleep(time.Minute * 5)


	for _, cid :=  range exp.localCids {
		if _, err := ipfs.Provide(cid); err != nil {
			return fmt.Errorf("Error upload cid %s: %s", cid.Content, err) 
		}
	}

	// save cids in a file :)
	cidsLog := utils.NewLogger("cids.log")
	aux, _  := json.Marshal(exp.localCids)
	cidsLog.Print(string(aux))

	log.Printf("Upload %d Cids", len(exp.localCids))

	total_ext_cids := len(exp.externalCids)
	for {

		target := exp.externalCids[ rand.Intn( total_ext_cids ) ]

		peers, err := ipfs.FindProviders(target)

		if err != nil {
			return err
		}

		if len(peers) == 0 {
			log.Printf("Unable to resolve %v", target)
		} else {
			log.Printf("Found %d providers of %v", len(peers), target)
		}

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