package utils

import (
	"encoding/json"
	"fmt"
	"time"
	"os"
	"math/rand"
	"reflect"
)


// some usefult constants for testint sake
type CIDType int

const (
	Normal CIDType = iota
	Secure
)

// TODO: rename to CidRecord
// TODO: think about the DEFAULT type and find a better solution for this
type CidInfo struct {
	Cid  string
	Type CIDType
}

func (cidType CIDType) String() string {
	switch cidType {
	case Normal:
		return "Normal"
	case Secure:
		return "Secure"
	default:
		panic(
			fmt.Sprintf("Invalid CIDType %d", cidType),
		)
	}
}

func (info *CidInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(info.Cid)
}


type NodeConfig struct {
	// cid configs
	CidsPerNode int `koanf:"EXP_CIDS_PER_NODE"`
	CidFileName string `koanf:"EXP_CIDS_FILE"`

	// membership configs
	BootDirectory string `koanf:"EXP_BOOT_DIR"`
	BootFileName string `koanf:"EXP_BOOT_FILE"`
	TotalNodes int `koanf:"EXP_TOTAL_NODES"`

	// timers
	ExpDuration  time.Duration `koanf:"EXP_DURATION"`
	ResolveWaitTime time.Duration `koanf:"EXP_RESOLVE_WAIT_TIME"`
	StartTime int64 `koanf:"EXP_START_TIME"`
	GracePeriod time.Duration `koanf:"EXP_GRACE_PERIOD"`

	// node configs
	NodeSeqNr int `koanf:"NODE_SEQ_NUM"`
	Mode string `koanf:"NODE_MODE"`
	Role string `koanf:"NODE_ROLE"`

	resolveAll bool
	isBootstrap bool
}

// TODO: try to think of a better solution
func (cfg *NodeConfig) Validate() error {

	if cfg.CidsPerNode <= 0 {
		return fmt.Errorf("Invalid EXP_CIDS_PER_NODE (%d): should be greather than 0", cfg.CidsPerNode);
	}

	if cfg.CidFileName == "" {
		return fmt.Errorf("Invalid EXP_CIDS_FILE (%s): should not be empty", cfg.CidFileName);
	}

	if cfg.TotalNodes <= 0 {
		return fmt.Errorf("Invalid EXP_TOTAL_NODES (%d): should be greather than 0", cfg.TotalNodes);
	}

	if cfg.ExpDuration <= 0 {
		return fmt.Errorf("Invalid EXP_DURATION (%d): should be greather than 0", cfg.ExpDuration);
	}

	if cfg.GracePeriod <= 0 {
		return fmt.Errorf("Invalid EXP_GRACE_PERIOD (%d): should be greather than 0", cfg.GracePeriod);
	}

	if cfg.NodeSeqNr < 0 {
		return fmt.Errorf("Invalid NODE_SEQ_NUM (%d): should be greather than or equal to 0", cfg.NodeSeqNr);
	}

	if cfg.StartTime <= time.Now().Unix() {
		return fmt.Errorf(
			"Invalid EXP_START_TIME (%d): should be greather than the current time", cfg.StartTime,
		);
	}

	switch cfg.Role {
	case "bootstrap":
		if cfg.BootDirectory == "" {
			return fmt.Errorf("Invalid EXP_BOOT_DIR: should not be empty");
		}
		cfg.isBootstrap = true
	case "worker":
		if cfg.BootFileName == "" {
			return fmt.Errorf("Invalid EXP_BOOT_FILE: should not be empty");
		}
		cfg.isBootstrap = false
	default:
		return fmt.Errorf(
			"Invalid NODE_ROLE ('%s'): should be one of [bootstrap, worker]", cfg.Role,
		);
	}

	switch cfg.Mode {
	case "normal":
		cfg.resolveAll = false
	case "secure", "default":
		cfg.resolveAll = true
	default:
		return fmt.Errorf(
			"Invalid NODE_MODE ('%s'): should be one of [normal, secure, default]", cfg.Mode,
		);
	}

	return nil
}

func (cfg * NodeConfig) IsBootstrap() bool {
	return cfg.isBootstrap
}

func (cfg * NodeConfig) ResolveAll() bool {
	return cfg.resolveAll
}


func (cfg * NodeConfig) Print(){

	val := reflect.ValueOf(cfg).Elem()
    for i:=0; i<val.NumField();i++{

		field := val.Type().Field(i)
		tag   := field.Tag.Get("koanf")
		if tag != "" {
			value := val.Field(i).Interface()
			fmt.Printf("%-25s: %v\n", tag, value)
		}
    }


}

func (cfg * NodeConfig) LoadBootstraps() ([]string, error){

	data, err := os.ReadFile(cfg.BootFileName)
	if err != nil {
		return nil, fmt.Errorf("Error reading boot file (%s): %s\n", cfg.BootFileName, err)
	}

	var (
		nodes [][]string
		chosen  [] string
	)

	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("Error unmarshalling '%s': %s\n", cfg.BootFileName, err)
	}

	for _, node := range nodes {
		chosen = append(chosen, node[ rand.Intn(len(node)) ])
	}

	if len(chosen) == 0 {
		return nil, fmt.Errorf("Not nodes found '%s'\n", cfg.BootFileName)
	}

	return chosen, nil
}