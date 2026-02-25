# NES Emulator - WebAssembly Display

This is the web-based frontend for the NES emulator, using HTML5 Canvas for rendering.

## Building

```bash
# Build the WASM binary
GOOS=js GOARCH=wasm go build -o cmd/wasm-display/nes.wasm ./cmd/wasm-display

# Or use make
make wasm
```

## Running

1. Copy the Go WASM support file:
   ```bash
   cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" cmd/wasm-display/
   ```

2. Serve the `cmd/wasm-display` directory with any HTTP server:
   ```bash
   # Using Python
   cd cmd/wasm-display && python3 -m http.server 8080

   # Using Go
   go run golang.org/x/tools/cmd/present@latest  # or any static server
   ```

3. Open http://localhost:8080 in your browser

4. Click "Load ROM" and select a `.nes` file

## Controls

| Key | Action |
|-----|--------|
| Arrow Keys | D-Pad |
| X | A Button |
| Z | B Button |
| Enter | Start |
| Shift | Select |
| P | Pause/Resume |

## Browser Compatibility

Tested on modern browsers with WebAssembly support:
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+
