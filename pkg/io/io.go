package io

import (
	"bufio"
	"strings"

	"github.com/rotisserie/eris"
	"golang.org/x/text/transform"
)

func ScannerSkipUntil(scanner *bufio.Scanner, token string, skipMax int) bool {
	i := 0
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), token) {
			return true
		}
		i++
		if i > skipMax {
			return false
		}
	}

	return false
}

func ReaderSkipUntil(reader *bufio.Reader, token string, skipMax int) error {
	i := 0
	for {
		buf, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(buf), token) {
			return nil
		}
		i++
		if i > skipMax {
			return eris.New("Token not found")
		}
	}
}

type Transformer func(string) string

type replacingTrasnformer struct {
	from, to string
}

func (t *replacingTrasnformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	nSrc = len(src)
	nDst = nSrc
	copy(dst, []byte(strings.ReplaceAll(string(src), t.from, t.to)))
	err = nil
	return
}

func (t *replacingTrasnformer) Reset() {}

func NewReplacingTransformer(from, to string) transform.Transformer {
	return &replacingTrasnformer{
		from: from,
		to:   to,
	}
}
