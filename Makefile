BASE_MODULE  := $(shell go list -m)
BIN_DIR      := bin

DEPS := *.go ./**/*.go

all: $(DEPS)
	go build -o $(BIN_DIR)/ipfs-client

# BOOT_NODE    := $(BIN_DIR)/boot-client
# DEFAULT_NODE := $(BIN_DIR)/ipfs-client
# BINS         := $(DEFAULT_NODE) $(BOOT_NODE)
# all: $(BINS)
# $(DEFAULT_NODE): $(DEPS) default-node/*.go experiments/*.go
# 	go build -o $@ $(BASE_MODULE)/default-node

# $(BOOT_NODE): $(DEPS) bootstrap-node/*.go
# 	go build -o $@ $(BASE_MODULE)/bootstrap-node

.PHONNY: clean
clean:
	@rm -rf $(BIN_DIR)
