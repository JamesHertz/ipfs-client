BIN_DIR := bin
BIN     := $(BIN_DIR)/ipfs-client

all: $(BIN)

$(BIN): ./**/*.go
	go build -o $@

.PHONNY: clean
clean:
	@rm -rf $(BIN_DIR)
