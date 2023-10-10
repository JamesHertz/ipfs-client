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

	. "github.com/JamesHertz/ipfs-client/utils"
	utils "github.com/ipfs/kubo/cmd/ipfs/util"
)

// TODO: a structe called context or config that has the values
//       for each one of the enviroment variables

type ResolveExperiment struct {
	localCids    []CidInfo
	externalCids []CidInfo

	mode string
}

// TODO: accept a ctx as argument wich is filled with the env value
func loadCids(filename string, expCidsNumber int) ([]CidInfo, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cids := strings.Split(
		strings.Trim(string(data), "\n"), "\n",
	)

	if len(cids) < expCidsNumber {
		return nil, fmt.Errorf("Expected at least %d cids but found %d nodes that.", expCidsNumber, len(cids))
	}

	cids = cids[:expCidsNumber]

	cidsInfo := make([]CidInfo, len(cids))

	for i, cid := range cids {
		var cidType CIDType
		// TODO: find a better solution for this
		if i < len(cids)/2 {
			cidType = Normal
		} else {
			cidType = Secure
		}

		cidsInfo[i] = CidInfo{
			Cid:  cid,
			Type: cidType,
		}
	}

	return cidsInfo, nil
}

const InterResolveTimeout = 10 * time.Second

func NewResolveExperiment() (Experiment, error) {
	// TODO: replace this with a global config file
	//       and use the information already avalable
	// 	     on the node about the mode :)
	node_seq_var := os.Getenv("NODE_SEQ_NUM")
	total_nodes_var := os.Getenv("EXP_TOTAL_NODES")
	cids_per_node_var := os.Getenv("EXP_CIDS_PER_NODE")
	cids_file := os.Getenv("EXP_CIDS_FILE")
	node_type := os.Getenv("MODE")

	if node_type == "" {
		return nil, fmt.Errorf("MODE not set")
	}

	seqNum, err := strconv.Atoi(node_seq_var)
	if err != nil {
		return nil, fmt.Errorf("Error parsing NODE_SEQ_NUM (%s): %s", node_seq_var, err)
	}

	totalNodes, err := strconv.Atoi(total_nodes_var)
	if err != nil {
		return nil, fmt.Errorf("Error parsing EXP_TOTAL_NODES (%s): %s", total_nodes_var, err)
	}

	cidsPerNode, err := strconv.Atoi(cids_per_node_var)
	if err != nil {
		return nil, fmt.Errorf("Error parsing CIDS_PER_NODE (%s): %s", cids_per_node_var, err)
	}

	// number of cids expected
	total_exp_cids := cidsPerNode * totalNodes

	all_cids, err := loadCids(cids_file, total_exp_cids)
	if err != nil {
		return nil, fmt.Errorf("Error get cids from EXP_CIDS_FILE (%s): %s", cids_file, err)
	}

	if seqNum >= totalNodes {
		return nil, fmt.Errorf("Invalid NODE_SEQ_NUMBER %d it should be [0,%d[", seqNum, totalNodes)
	}

	// ...
	if node_type == "Normal" {
		all_cids = all_cids[:total_exp_cids/2]
	}

	var externalCids, localCids []CidInfo

	start_cid := seqNum * cidsPerNode
	end_cid   := start_cid + cidsPerNode

	localCids = append(localCids, all_cids[start_cid:end_cid]...)
	externalCids = append(externalCids, all_cids[:start_cid]...)
	externalCids = append(externalCids, all_cids[end_cid:]...)

	return &ResolveExperiment{
		localCids:    localCids,
		externalCids: externalCids,
	}, nil
}

func (exp *ResolveExperiment) Start(ipfs *client.IpfsClientNode, ctx context.Context) error {
	grace_period_var := os.Getenv("EXP_GRADE_PERIOD")
	grace_period, err := strconv.Atoi(grace_period_var)

	if err != nil {
		return fmt.Errorf("Error parsing EXP_GRADE_PERIOD (%s): %s", grace_period_var, err)
	}

	log.Println("Starting experiment...")

	log.Printf("Waiting %d minute...", grace_period)
	time.Sleep(time.Minute * time.Duration(grace_period))

	for _, cid := range exp.localCids {
		if _, err := ipfs.Provide(cid); err != nil {
			return fmt.Errorf("Error upload cid %s: %s", cid.Cid, err)
		}
	}

	// save cids in a file :)
	cidsLog := utils.NewLogger("cids.log")
	aux, _  := json.Marshal(exp.localCids)
	cidsLog.Print(string(aux))

	log.Printf("Upload %d Cids", len(exp.localCids))

	total_ext_cids := len(exp.externalCids)
	for {

		target := exp.externalCids[rand.Intn(total_ext_cids)]

		peers, err := ipfs.FindProviders(target)

		if err != nil {
			return err
		}

		if len(peers) == 0 {
			log.Printf("Unable to resolve %v", target.Cid)
		} else {
			log.Printf("Found %d providers of %v", len(peers), target.Cid)
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
