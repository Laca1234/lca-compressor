package engine

type Token struct {
	IsMatch bool
	Literal byte
	Offset  uint16
	Length  uint16
}

func CompressLZ77(data []byte, windowSize uint16) []Token {
	var tokens []Token
	n := len(data)
	if n == 0 {
		return tokens
	}

	for i := 0; i < n; {
		matchOffset := uint16(0)
		matchLength := uint16(0)

		start := 0
		if i > int(windowSize) {
			start = i - int(windowSize)
		}

		for j := start; j < i; j++ {
			length := 0
			for i+length < n && data[j+length] == data[i+length] && length < 255 {
				length++
			}

			if uint16(length) > matchLength {
				matchLength = uint16(length)
				matchOffset = uint16(i - j)
			}
		}

		if matchLength >= 3 {
			tokens = append(tokens, Token{
				IsMatch: true,
				Offset:  matchOffset,
				Length:  matchLength,
			})
			i += int(matchLength)
		} else {
			tokens = append(tokens, Token{
				IsMatch: false,
				Literal: data[i],
			})
			i++
		}
	}

	return tokens
}

func DecompressLZ77(tokens []Token) []byte {
	var out []byte
	for _, tok := range tokens {
		if !tok.IsMatch {
			out = append(out, tok.Literal)
		} else {
			start := len(out) - int(tok.Offset)
			for i := 0; i < int(tok.Length); i++ {
				out = append(out, out[start+i])
			}
		}
	}
	return out
}
