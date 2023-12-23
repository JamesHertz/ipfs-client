package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"

	// "strconv"
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

	resolveWaitTime 	time.Duration
	publishWaitTime 	time.Duration
	// other timers and stuffs c:
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

func NewResolveExperiment(cfg * NodeConfig) (Experiment, error) {

	seqNum := cfg.NodeSeqNr
	totalNodes := cfg.TotalNodes
	experienceCidNr := cfg.CidsPerNode * cfg.TotalNodes

	allCids, err := loadCids(cfg.CidFileName, experienceCidNr)
	if err != nil {
		return nil, fmt.Errorf("Error get cids from %s: %s", cfg.CidFileName, err)
	}

	if seqNum >= totalNodes {
		return nil, fmt.Errorf("Invalid node sequence number (%d) it should be [0,%d[", seqNum, totalNodes)
	}

	if !cfg.ResolveAll() {
		allCids = allCids[:experienceCidNr/2]
	}

	var externalCids, localCids []CidInfo

	startCid := seqNum * cfg.CidsPerNode
	endCid   := startCid + cfg.CidsPerNode

	localCids = append(localCids, allCids[startCid:endCid]...)
	externalCids = append(externalCids, allCids[:startCid]...)
	externalCids = append(externalCids, allCids[endCid:]...)

	return &ResolveExperiment{
		localCids:    localCids,
		externalCids: externalCids,
		resolveWaitTime:    cfg.ResolveWaitTime,
		publishWaitTime:   cfg.PublishWaitTime,
	}, nil
}

func (exp *ResolveExperiment) Start(ipfs *client.IpfsClientNode, ctx context.Context) error {

	log.Println("Starting experiment...")

	for _, cid := range exp.localCids {
		if _, err := ipfs.Provide(cid); err != nil {
			return fmt.Errorf("Error upload cid %s: %s", cid.Cid, err)
		}
		time.Sleep(exp.publishWaitTime)
	}

	// save cids in a file :)
	cidsLog := utils.NewLogger("cids.log")
	aux, _  := json.Marshal(exp.localCids)
	cidsLog.Print(string(aux))

	log.Printf("Uploaded %d Cids", len(exp.localCids))

	totalExternalCids := len(exp.externalCids)
	for {

		target := exp.externalCids[rand.Intn(totalExternalCids)]

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
		case <-time.After(exp.resolveWaitTime):
		}

	}
}
