# LCA Compressor & Archiver

A high-performance, multi-threaded file compression and recursive directory archiving utility written in Go.

## Features
- **Recursive Directory Archiving:** Pack and unpack entire directory structures seamlessly using `archive/tar`.
- **Stream-Based Architecture:** Uses `io.Pipe()` to stream tar-packaged data directly into the compression pipeline without intermediate disk storage.
- **Multi-Threaded Processing:** Concurrent chunk processing across configurable worker threads.
- **Hybrid Compression Engine:** Combines sliding-window LZ77 tokenization with dynamic Huffman coding.
- **Data Integrity:** Built-in CRC32 checksum verification per block.

## Installation & Building
Make sure you have Go installed, then build the binary:
```powershell
go build -o lca.exe cmd/lca/main.go
