package service

import (
	"io"
)

type BitWriter struct {
	w   io.Writer
	buf byte
	n   uint8
}

func NewBitWriter(w io.Writer) *BitWriter {
	return &BitWriter{w: w}
}

func (bw *BitWriter) WriteBits(val uint32, count uint8) error {
	for i := int(count - 1); i >= 0; i-- {
		bit := byte((val >> uint(i)) & 1)
		bw.buf = (bw.buf << 1) | bit
		bw.n++
		if bw.n == 8 {
			if _, err := bw.w.Write([]byte{bw.buf}); err != nil {
				return err
			}
			bw.buf = 0
			bw.n = 0
		}
	}
	return nil
}

func (bw *BitWriter) Flush() error {
	if bw.n > 0 {
		bw.buf <<= (8 - bw.n)
		if _, err := bw.w.Write([]byte{bw.buf}); err != nil {
			return err
		}
		bw.buf = 0
		bw.n = 0
	}
	return nil
}

type BitReader struct {
	r   io.Reader
	buf byte
	n   uint8
}

func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{r: r}
}

func (br *BitReader) ReadBits(count uint8) (uint32, error) {
	var result uint32
	for i := uint8(0); i < count; i++ {
		if br.n == 0 {
			b := make([]byte, 1)
			_, err := br.r.Read(b)
			if err != nil {
				return 0, err
			}
			br.buf = b[0]
			br.n = 8
		}
		bit := uint32((br.buf >> (br.n - 1)) & 1)
		result = (result << 1) | bit
		br.n--
	}
	return result, nil
}
