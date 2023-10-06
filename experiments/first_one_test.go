package experiments

import (
	// "strings"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/JamesHertz/ipfs-client/utils"
)


type testConfig struct {
	Mode string
	cids_conf struct {
		filename string
		cids []string
	}
	node_seq     int
	total_nodes  int
	cid_per_node int
}

func TestExperimentSetup(t *testing.T){

	rawCids := func(cids []CidInfo) []string{
		aux := make([]string, len(cids))
		for i, cid := range cids {
			aux[i] = cid.Content
		}
		return aux
	}

	cids := []string{
		"cid-normal-1",
		"cid-normal-2",
		"cid-normal-3",
		"cid-normal-4",
		"cid-secure-1",
		"cid-secure-2",
		"cid-secure-3",
		"cid-secure-4",
	}

	cfg := &testConfig{
		Mode: "Normal",
		node_seq: 1,
		total_nodes: 4,
		cid_per_node: 2,
	}

	cfg.cids_conf.filename = "/tmp/cids.txt"
	cfg.cids_conf.cids = cids

	setupEnviroment(t, cfg)
	defer os.Remove(cfg.cids_conf.filename)

	aux, err := NewResolveExperiment()
	require.Nil(t, err)

	exp, ok := aux.(*ResolveExperiment)
	require.True(t, ok)

	require.EqualValues(t, cids[2:4], rawCids(exp.localCids),  "Incorrect local cids")
	require.EqualValues(t, cids[:2], rawCids(exp.externalCids), "Incorrect external cids")

	t.Fail()
}


func setupEnviroment(t *testing.T, cfg *testConfig ){

/*
	node_seq_var      :=  os.Getenv("NODE_SEQ_NUM") 
	total_nodes_var   :=  os.Getenv("EXP_TOTAL_NODES") 
	cids_per_node_var :=  os.Getenv("EXP_CIDS_PER_NODE") 
	cids_file         :=  os.Getenv("EXP_CIDS_FILE")
	node_type         :=  os.Getenv("MODE")
*/

	err := os.WriteFile(
		cfg.cids_conf.filename, 
		[]byte(strings.Join(cfg.cids_conf.cids, "\n")), 
		0644,
	)

	require.Nil(t, err, "Error writing to %s", cfg.cids_conf.filename)

	env := []struct{ name, value string }{
		{
			name: "NODE_SEQ_NUM",
			value: strconv.Itoa(cfg.node_seq),
		},
		{
			name: "EXP_TOTAL_NODES",
			value: strconv.Itoa(cfg.total_nodes),
		},
		{
			name: "EXP_CIDS_PER_NODE",
			value: strconv.Itoa(cfg.cid_per_node),
		},
		{
			name: "EXP_CIDS_FILE",
			value: cfg.cids_conf.filename,
		},
		{
			name: "MODE",
			value: cfg.Mode,
		},
	}

	for _, e := range env {
		err := os.Setenv(e.name, e.value)
		require.Nil(t, err, "Error setting variable: %s", e.name)
	}
}
