package service

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	LcaMagic   uint32 = 0x4C434100 // "LCA\0"
	LcaVersion uint8  = 0x01
	FlagCRC32  uint8  = 0x02
)

type GlobalHeader struct {
	Magic   uint32
	Version uint8
	Flags   uint8
}

type BlockHeader struct {
	BlockID          uint32
	UncompressedSize uint32
	CompressedSize   uint32
	Checksum         uint32
}

func WriteGlobalHeader(w io.Writer, flags uint8) error {
	header := GlobalHeader{
		Magic:   LcaMagic,
		Version: LcaVersion,
		Flags:   flags,
	}
	return binary.Write(w, binary.BigEndian, header)
}

func ReadGlobalHeader(r io.Reader) (GlobalHeader, error) {
	var header GlobalHeader
	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		return header, err
	}
	if header.Magic != LcaMagic {
		return header, fmt.Errorf("invalid magic bytes: 0x%X", header.Magic)
	}
	return header, nil
}

func WriteBlockHeader(w io.Writer, bh BlockHeader) error {
	return binary.Write(w, binary.BigEndian, bh)
}

func ReadBlockHeader(r io.Reader) (BlockHeader, error) {
	var bh BlockHeader
	err := binary.Read(r, binary.BigEndian, &bh)
	return bh, err
}
