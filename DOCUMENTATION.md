================================================================================

LCA COMPRESSOR TECHNICAL DOCUMENTATION

================================================================================



1. ARCHITECTURE OVERVIEW

--------------------------------------------------------------------------------

The `lca-compressor` pipeline is engineered around zero-disk-footprint streaming 

and concurrent worker pools.



[Directory / File] 

       │ (archive/tar)

       ▼

  [io.Pipe()] 

       │ (stream)

       ▼

[Multi-Threaded Chunking (LZ77 + Huffman)] 

       │ (blocks + CRC32)

       ▼

[.lca Archive File]





2. CORE COMPONENTS

--------------------------------------------------------------------------------



### A. Archiving Layer (`internal/controller/archive.go`)

- PackFolder(srcDir string, w io.Writer): Recursively walks the source directory, 

  normalizes file paths to relative Unix-style paths, and streams `tar` headers 

  and payloads into the writer.

- UnpackArchive(r io.Reader, destDir string): Reads the tar archive stream, safely 

  resolves destination paths, recreates directory trees, and restores files 

  byte-for-byte.



### B. Pipeline & Streaming (`internal/controller/pipeline.go`)

- Coordinates asynchronous chunk dispatchers (`ChunkJob` and `ChunkResult`).

- Buffers streams through `io.Pipe()` connecting the tar archiver directly to 

  the compression engine.



### C. Compression Engine (`internal/engine/`)

- LZ77: Sliding window pattern matching for redundancy reduction.

- Huffman Coding: Frequency-based variable-length bit encoding with serialized 

  tree headers.



### D. Service Layer (`internal/service/`)

- Manages global headers, block headers, and bitwise readers/writers.

- Implements IEEE CRC32 checksum validation to catch corruption during decompression.

