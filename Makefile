NES_EMULATOR = nes-emulator
ROM_INFO = rom-info
INSPECT_PPU = inspect-ppu
ASCII_RENDER = ascii-render
DETAILED_RENDER = detailed-render
VERIFY_COLORS = verify-colors
WATCH_GAME = watch-game

WASM_DIR = cmd/wasm-display
WASM_BINARY = $(WASM_DIR)/nes.wasm

BINARIES = $(NES_EMULATOR) $(ROM_INFO) $(INSPECT_PPU) $(ASCII_RENDER) $(DETAILED_RENDER) $(VERIFY_COLORS) $(WATCH_GAME)

RELEASE_FLAGS = -ldflags="-s -w"

.PHONY: all clean test nes-emulator tools release wasm wasm-serve

all: $(BINARIES)

release:
	go build $(RELEASE_FLAGS) -o $(NES_EMULATOR) ./cmd/sdl-display

$(NES_EMULATOR):
	go build -o $(NES_EMULATOR) ./cmd/sdl-display

$(ROM_INFO):
	go build -o $(ROM_INFO) ./cmd/rom-info

$(INSPECT_PPU):
	go build -o $(INSPECT_PPU) ./cmd/inspect-ppu

$(ASCII_RENDER):
	go build -o $(ASCII_RENDER) ./cmd/ascii-render

$(DETAILED_RENDER):
	go build -o $(DETAILED_RENDER) ./cmd/detailed-render

$(VERIFY_COLORS):
	go build -o $(VERIFY_COLORS) ./cmd/verify-colors

$(WATCH_GAME):
	go build -o $(WATCH_GAME) ./cmd/watch-game

tools: $(ROM_INFO) $(INSPECT_PPU) $(ASCII_RENDER) $(DETAILED_RENDER) $(VERIFY_COLORS) $(WATCH_GAME)

test:
	go test ./...

test-race:
	go test -race ./...

deps:
	go mod download

wasm:
	GOOS=js GOARCH=wasm go build -o $(WASM_BINARY) ./cmd/wasm-display
	@GOROOT=$$(go env GOROOT); \
	if [ -f "$$GOROOT/misc/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/misc/wasm/wasm_exec.js" $(WASM_DIR)/; \
	elif [ -f "$$GOROOT/lib/wasm/wasm_exec.js" ]; then \
		cp "$$GOROOT/lib/wasm/wasm_exec.js" $(WASM_DIR)/; \
	else \
		echo "Error: wasm_exec.js not found in GOROOT"; exit 1; \
	fi

wasm-serve: wasm
	@echo "Starting server at http://localhost:8080"
	cd $(WASM_DIR) && python3 -m http.server 8080

clean:
	rm -f $(BINARIES) $(WASM_BINARY) $(WASM_DIR)/wasm_exec.js
