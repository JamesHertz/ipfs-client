package experiments

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JamesHertz/ipfs-client/utils"
	. "github.com/JamesHertz/ipfs-client/utils"
	"github.com/stretchr/testify/require"
)

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

	cfg := NodeConfig{
		Mode:         "normal",
		NodeSeqNr:   1,
		TotalNodes:  4,
		CidsPerNode: 2,
		CidFileName: "/tmp/cids.txt",

		// other configs:
		Role: "worker",
		BootFileName: "/path/to/boot/file",
		ExpDuration: 1,
		ResolveWaitTime: 1,
		GracePeriod: 1,
		StartTime: time.Now().Unix() + 120, // two minutes
	}


	err := os.WriteFile(
		cfg.CidFileName, []byte(strings.Join(cids, "\n")), 0666,
	)
	require.Nil(t, err)


	require.Nil(t, cfg.Validate())

	// simple test 1
	aux, err := NewResolveExperiment(&cfg)
	require.Nil(t, err)

	exp, ok := aux.(*ResolveExperiment)
	require.True(t, ok)

	require.EqualValues(t, cids[2:4], rawCids(exp.localCids), "Incorrect local cids")
	require.EqualValues(t, cids[:2], rawCids(exp.externalCids), "Incorrect external cids")

	// simple test 2
	cfg.Mode = "secure"
	cfg.NodeSeqNr = 2
	require.Nil(t, cfg.Validate())

	aux, err = NewResolveExperiment(&cfg)
	require.Nil(t, err)

	exp, ok = aux.(*ResolveExperiment)
	require.True(t, ok)

	require.EqualValues(t, cids[4:6], rawCids(exp.localCids), "Incorrect local cids")
	require.Equal(t, len(cids)-cfg.CidsPerNode, len(exp.externalCids), "Incorrect external cids size")
	require.EqualValues(t, cids[:4], rawCids(exp.externalCids[:4]))
	require.EqualValues(t, cids[6:], rawCids(exp.externalCids[4:]))
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
