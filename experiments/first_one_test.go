package experiments

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/JamesHertz/ipfs-client/utils"
	. "github.com/JamesHertz/ipfs-client/utils"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	mode      string
	cids_conf struct {
		filename string
		cids     []string
	}
	node_seq     int
	total_nodes  int
	cid_per_node int
}

func TestExperimentSetup(t *testing.T) {

	rawCids := func(cids []CidInfo) []string {
		aux := make([]string, len(cids))
		for i, cid := range cids {
			aux[i] = cid.Cid
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
		mode:         "Normal",
		node_seq:     1,
		total_nodes:  4,
		cid_per_node: 2,
	}

	cfg.cids_conf.filename = "/tmp/cids.txt"
	cfg.cids_conf.cids = cids

	setupEnviroment(t, cfg)
	defer os.Remove(cfg.cids_conf.filename)

	// simple test 1
	aux, err := NewResolveExperiment()
	require.Nil(t, err)

	exp, ok := aux.(*ResolveExperiment)
	require.True(t, ok)

	require.EqualValues(t, cids[2:4], rawCids(exp.localCids), "Incorrect local cids")
	require.EqualValues(t, cids[:2], rawCids(exp.externalCids), "Incorrect external cids")

	// simple test 2
	cfg.mode = "Secure"
	cfg.node_seq = 2
	setupEnviroment(t, cfg)

	aux, err = NewResolveExperiment()
	require.Nil(t, err)

	exp, ok = aux.(*ResolveExperiment)
	require.True(t, ok)

	require.EqualValues(t, cids[4:6], rawCids(exp.localCids), "Incorrect local cids")
	require.Equal(t, len(cids)-cfg.cid_per_node, len(exp.externalCids), "Incorrect external cids size")
	require.EqualValues(t, cids[:4], rawCids(exp.externalCids[:4]))
	require.EqualValues(t, cids[6:], rawCids(exp.externalCids[4:]))
}

func setupEnviroment(t *testing.T, cfg *testConfig) {

	err := os.WriteFile(
		cfg.cids_conf.filename,
		[]byte(strings.Join(cfg.cids_conf.cids, "\n")),
		0644,
	)

	require.Nil(t, err, "Error writing to %s", cfg.cids_conf.filename)

	env := []struct{ name, value string }{
		{
			name:  "NODE_SEQ_NUM",
			value: strconv.Itoa(cfg.node_seq),
		},
		{
			name:  "EXP_TOTAL_NODES",
			value: strconv.Itoa(cfg.total_nodes),
		},
		{
			name:  "EXP_CIDS_PER_NODE",
			value: strconv.Itoa(cfg.cid_per_node),
		},
		{
			name:  "EXP_CIDS_FILE",
			value: cfg.cids_conf.filename,
		},
		{
			name:  "MODE",
			value: cfg.mode,
		},
	}

	for _, e := range env {
		err := os.Setenv(e.name, e.value)
		require.Nil(t, err, "Error setting variable: %s", e.name)
	}
}

func TestMarshaler( t * testing.T){


	cids := []utils.CidInfo{
		{
			Cid:  "cid-normal-1",
			Type: utils.Normal,
		},
		{
			Cid:  "cid-secure-1",
			Type: utils.Secure,
		},
	}

	expected := []string{"cid-normal-1", "cid-secure-1"}

	bytes, err := json.Marshal(cids)
	require.Nil(t, err)

	var res []string
	require.Nil(t, json.Unmarshal(bytes, &res))

	require.EqualValues(t, expected, res)
}
