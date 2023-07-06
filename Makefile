BASE_MODULE  := $(shell go list -m)
BIN_DIR      := bin
BOOT_NODE    := $(BIN_DIR)/boot-client
DEFAULT_NODE := $(BIN_DIR)/ipfs-client
BINS         := $(DEFAULT_NODE) $(BOOT_NODE)

DEPS := default-node/*.go go.*

all: $(BINS)

$(DEFAULT_NODE): $(DEPS) client/*.go
	go build -o $@ $(BASE_MODULE)/default-node

$(BOOT_NODE): $(DEPS) client/*.go
	go build -o $@ $(BASE_MODULE)/bootstrap-node

.PHONNY: clean
clean:
	@rm -rf $(BIN_DIR)
