# NES Emulator Makefile

# Output binary names
NES_EMULATOR = nes-emulator
ROM_INFO = rom-info
INSPECT_PPU = inspect-ppu
ASCII_RENDER = ascii-render
DETAILED_RENDER = detailed-render
VERIFY_COLORS = verify-colors
WATCH_GAME = watch-game

# All binaries
BINARIES = $(NES_EMULATOR) $(ROM_INFO) $(INSPECT_PPU) $(ASCII_RENDER) $(DETAILED_RENDER) $(VERIFY_COLORS) $(WATCH_GAME)

# Release build flags (strip symbols, disable DWARF)
RELEASE_FLAGS = -ldflags="-s -w"

.PHONY: all clean test nes-emulator tools release

# Default: build everything
all: $(BINARIES)

# Release build (optimized, smaller binary)
release:
	go build $(RELEASE_FLAGS) -o $(NES_EMULATOR) ./cmd/sdl-display

# Main emulator with SDL display
$(NES_EMULATOR):
	go build -o $(NES_EMULATOR) ./cmd/sdl-display


# Utility tools
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

# Build all tools except nes-emulator
tools: $(ROM_INFO) $(INSPECT_PPU) $(ASCII_RENDER) $(DETAILED_RENDER) $(VERIFY_COLORS) $(WATCH_GAME)

# Run tests
test:
	go test ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Download dependencies
deps:
	go mod download

# Clean build artifacts
clean:
	rm -f $(BINARIES)
