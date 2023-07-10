BASE_MODULE  := $(shell go list -m)
BIN_DIR      := bin
BOOT_NODE    := $(BIN_DIR)/boot-client
DEFAULT_NODE := $(BIN_DIR)/ipfs-client
SETUP_NODE   := $(BIN_DIR)/cfg-bootstrap
BINS         := $(DEFAULT_NODE) $(BOOT_NODE) $(SETUP_NODE)

DEPS := client/*.go go.*

all: $(BINS)

$(DEFAULT_NODE): $(DEPS) default-node/*.go experiments/*.go
	go build -o $@ $(BASE_MODULE)/default-node

$(BOOT_NODE): $(DEPS) bootstrap-node/*.go
	go build -o $@ $(BASE_MODULE)/bootstrap-node

$(SETUP_NODE): $(DEPS) setup/*.go
	go build -o $@ $(BASE_MODULE)/setup

.PHONNY: clean
clean:
	@rm -rf $(BIN_DIR)
