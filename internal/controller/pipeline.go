package controller

import (
	"bytes"
	"hash/crc32"
	"io"
	"runtime"
	"sort"
	"sync"

	"github.com/Laca1234/lca-compressor/internal/engine"
	"github.com/Laca1234/lca-compressor/internal/service"
)

type ChunkJob struct {
	ID   uint32
	Data []byte
}

type ChunkResult struct {
	ID               uint32
	UncompressedSize uint32
	CompressedSize   uint32
	Checksum         uint32
	Payload          []byte
}

type Pipeline struct {
	ChunkSize int
	Threads   int
}

func NewPipeline(chunkSize int, threads int) *Pipeline {
	if threads <= 0 {
		threads = runtime.NumCPU()
	}
	return &Pipeline{
		ChunkSize: chunkSize,
		Threads:   threads,
	}
}

func (p *Pipeline) CompressStream(r io.Reader, w io.Writer) error {
	if err := service.WriteGlobalHeader(w, service.FlagCRC32); err != nil {
		return err
	}

	jobs := make(chan ChunkJob, p.Threads*2)
	results := make(chan ChunkResult, p.Threads*2)

	var wg sync.WaitGroup
	for i := 0; i < p.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				checksum := crc32.ChecksumIEEE(job.Data)
				
				// LZ77 Match
				tokens := engine.CompressLZ77(job.Data, 4096)
				
				// Transform tokens to byte buffer
				var tokBuf bytes.Buffer
				for _, t := range tokens {
					if t.IsMatch {
						tokBuf.WriteByte(1)
						tokBuf.WriteByte(byte(t.Offset >> 8))
						tokBuf.WriteByte(byte(t.Offset & 0xFF))
						tokBuf.WriteByte(byte(t.Length))
					} else {
						tokBuf.WriteByte(0)
						tokBuf.WriteByte(t.Literal)
					}
				}
				
				rawBlock := tokBuf.Bytes()
				root, codes := engine.BuildHuffmanTree(rawBlock)

				var blockPayload bytes.Buffer
				_ = engine.SerializeTree(root, &blockPayload)

				var compressedBitBuf bytes.Buffer
				bw := service.NewBitWriter(&compressedBitBuf)
				for _, b := range rawBlock {
					c := codes[b]
					_ = bw.WriteBits(c.Val, c.NBits)
				}
				_ = bw.Flush()

				blockPayload.Write(compressedBitBuf.Bytes())

				results <- ChunkResult{
					ID:               job.ID,
					UncompressedSize: uint32(len(job.Data)),
					CompressedSize:   uint32(blockPayload.Len()),
					Checksum:         checksum,
					Payload:          blockPayload.Bytes(),
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		var id uint32 = 0
		for {
			buf := make([]byte, p.ChunkSize)
			n, err := io.ReadFull(r, buf)
			if n > 0 {
				jobs <- ChunkJob{ID: id, Data: buf[:n]}
				id++
			}
			if err != nil {
				break
			}
		}
		close(jobs)
	}()

	pending := make(map[uint32]ChunkResult)
	var nextID uint32 = 0

	for res := range results {
		pending[res.ID] = res
		for {
			item, exists := pending[nextID]
			if !exists {
				break
			}
			delete(pending, nextID)

			bh := service.BlockHeader{
				BlockID:          item.ID,
				UncompressedSize: item.UncompressedSize,
				CompressedSize:   item.CompressedSize,
				Checksum:         item.Checksum,
			}
			if err := service.WriteBlockHeader(w, bh); err != nil {
				return err
			}
			if _, err := w.Write(item.Payload); err != nil {
				return err
			}
			nextID++
		}
	}

	return nil
}

func (p *Pipeline) DecompressStream(r io.Reader, w io.Writer) error {
	_, err := service.ReadGlobalHeader(r)
	if err != nil {
		return err
	}

	for {
		bh, err := service.ReadBlockHeader(r)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}

		payload := make([]byte, bh.CompressedSize)
		if _, err := io.ReadFull(r, payload); err != nil {
			return err
		}

		payloadReader := bytes.NewReader(payload)
		root, err := engine.DeserializeTree(payloadReader)
		if err != nil {
			return err
		}

		// Read remaining bytes as bitstream
		bitstreamBytes := payload[payload.Len()-payloadReader.Len():]
		br := service.NewBitReader(bytes.NewReader(bitstreamBytes))

		var rawBlock []byte
		// Traverse Huffman tree to decode rawBlock bytes
		for i := uint32(0); i < bh.UncompressedSize * 2; i++ { // safe upper bound or token count
			curr := root
			if curr == nil {
				break
			}
			for !curr.IsLeaf {
				bit, err := br.ReadBits(1)
				if err != nil {
					break
				}
				if bit == 0 {
					curr = curr.Left
				} else {
					curr = curr.Right
				}
				if curr == nil {
					break
				}
			}
			if curr != nil && curr.IsLeaf {
				rawBlock = append(rawBlock, curr.Byte)
			}
			if len(rawBlock) >= int(bh.UncompressedSize * 3) { // estimated token stream limit
				break
			}
		}

		// Reconstruct LZ77 tokens
		var tokens []engine.Token
		for idx := 0; idx < len(rawBlock); {
			flag := rawBlock[idx]
			idx++
			if flag == 1 {
				if idx+3 > len(rawBlock) {
					break
				}
				offset := uint16(rawBlock[idx])<<8 | uint16(rawBlock[idx+1])
				length := uint16(rawBlock[idx+2])
				idx += 3
				tokens = append(tokens, engine.Token{
					IsMatch: true,
					Offset:  offset,
					Length:  length,
				})
			} else {
				if idx >= len(rawBlock) {
					break
				}
				lit := rawBlock[idx]
				idx++
				tokens = append(tokens, engine.Token{
					IsMatch: false,
					Literal: lit,
				})
			}
		}

		originalData := engine.DecompressLZ77(tokens)
		if uint32(len(originalData)) > bh.UncompressedSize {
			originalData = originalData[:bh.UncompressedSize]
		}

		if _, err := w.Write(originalData); err != nil {
			return err
		}
	}

	return nil
}
