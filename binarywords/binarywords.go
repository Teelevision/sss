// Package binarywords provides an encoding of binary into words inspired by
// BIP39 and Diceware. Is uses a German wordlist from
// github.com/dys2p/wordlists-de.
package binarywords

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"iter"
	"slices"
	"strings"
)

//go:embed dys2p/wordlists-de/de-2048-v1.txt
var wordlistRaw []byte

var wordlist []string

const (
	bits = 11 // BIP39 uses 11 bits per word
	size = 1 << bits
	mask = size - 1
)

func init() {
	scanner := bufio.NewScanner(bytes.NewReader(wordlistRaw))
	for scanner.Scan() {
		wordlist = append(wordlist, strings.TrimSpace(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Sprintf("Error reading wordlist: %v", err))
	}
	if len(wordlist) != size {
		panic(fmt.Sprintf("Wordlist must contain exactly %d words.", size))
	}
}

// Encodes binary data to BIP39 mnemonic using the loaded wordlist
func Encode(data []byte) []string {
	return slices.Collect(EncodeToWords(data))
}

func EncodeToWords(data []byte) iter.Seq[string] {
	return func(yield func(string) bool) {
		for word := range encodeToIndexes(data) {
			if !yield(wordlist[word]) {
				return
			}
		}
	}
}

func encodeToIndexes(data []byte) iter.Seq[int] {
	return func(yield func(int) bool) {
		if len(data) == 0 {
			return
		}
		buf := 0
		n := 0
		for _, b := range data {
			buf = (buf << 8) | int(b)
			n += 8
			for n >= bits {
				n -= bits
				if !yield((buf >> n) & mask) {
					return
				}
				buf &= 1<<n - 1
			}
		}
		// Handle remaining bits (pad with a one and then zeros)
		if n > 0 {
			yield(((buf<<1 + 1) << (bits - n - 1)) & mask)
		} else {
			yield(1 << (bits - 1))
		}
	}
}

func Decode(words []string) []byte {
	indexes := make([]int, 0, len(words))
	for _, word := range words {
		indexes = append(indexes, slices.Index(wordlist, word))
	}
	return decodeFromIndexes(indexes)
}

func decodeFromIndexes(indexes []int) []byte {
	data := make([]byte, 0, (len(indexes)*bits+7)/8)
	var buf, n int
	for _, index := range indexes {
		buf = (buf << bits) | index
		n += bits
		for n >= 8 {
			n -= 8
			data = append(data, byte((buf>>n)&0xff))
			buf &= 1<<n - 1 // Clear the bits we've used
		}
	}
	if buf == 0 && len(data) > 0 && data[len(data)-1] == 1<<7 {
		data = data[:len(data)-1] // Remove the padding byte if it was added
	}
	return data
}
