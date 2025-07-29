package main

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SSSaaS/sssa-golang"
	"github.com/skip2/go-qrcode"
	"github.com/teelevision/sss/binarywords"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: sss encode [min] [num] | sss decode")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "encode":
		encode(os.Args[2:])
	case "decode":
		decode()
	default:
		fmt.Fprintln(os.Stderr, "Usage: sss encode [min] [num] | sss decode")
		os.Exit(1)
	}
}

func encode(args []string) {
	minimum := 2
	numShares := 2

	if len(args) >= 1 {
		if v, err := strconv.Atoi(args[0]); err == nil && v >= 2 {
			minimum = v
			numShares = v
		} else {
			fmt.Fprintln(os.Stderr, "Error: min must be an integer >= 2.")
			os.Exit(1)
		}
	}
	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[1]); err == nil && v >= minimum {
			numShares = v
		} else {
			fmt.Fprintln(os.Stderr, "Error: num must be an integer >= min.")
			os.Exit(1)
		}
	}

	fmt.Fprint(os.Stderr, "Enter password: ")
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading password:", err)
		os.Exit(1)
	}
	password = strings.TrimRight(password, "\r\n")

	shares, err := sssa.Create(minimum, numShares, password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating shares:", err)
		os.Exit(1)
	}

	// Verify decoding works
	secret, err := sssa.Combine(shares)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error combining shares:", err)
		os.Exit(1)
	}
	if secret != password {
		fmt.Fprintln(os.Stderr, "Error: Decoded password does not match original.")
		os.Exit(1)
	}

	// Create output folder with current timestamp
	outputDir := fmt.Sprintf("output/%d", time.Now().Unix())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	data := make([]tmplData, 0, len(shares))
	for i, share := range shares {
		data = append(data, tmplData{
			Index: i + 1,
			Words: writeShare(i+1, share, outputDir),
			Share: share,
		})
		fmt.Println(share)
	}

	// Write index.html
	indexPath := fmt.Sprintf("%s/index.html", outputDir)
	indexFile, err := os.Create(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating index.html: %v\n", err)
		os.Exit(1)
	}
	defer indexFile.Close()

	if err := tmpl.Execute(indexFile, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing index template: %v\n", err)
		os.Exit(1)
	}
}

func writeShare(index int, share string, outputDir string) (words []string) {
	// Generate QR code for the share
	qrPath := fmt.Sprintf("%s/share_%d.png", outputDir, index)
	if err := qrcode.WriteFile(share, qrcode.Medium, 512, qrPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating QR code: %v\n", err)
		os.Exit(1)
	}

	// Prepare binarywords for the share
	for j := 0; j < len(share); j += 44 {
		end := j + 44
		if end > len(share) {
			end = len(share)
		}
		chunk := share[j:end]
		data, err := base64.URLEncoding.DecodeString(chunk)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding chunk: %v\n", err)
			os.Exit(1)
		}
		words = append(words, binarywords.Encode(data)...)
	}

	return words
}

func decode() {
	fmt.Fprintln(os.Stderr, "Enter shares (one per line, empty line to finish):")
	reader := bufio.NewReader(os.Stdin)
	var shares []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// Try base64 decode to check if it's a valid share
		if len(line) >= 44 {
			_, err = base64.URLEncoding.DecodeString(line[:44])
			if err == nil {
				shares = append(shares, line)
				continue
			}
		}

		// Try binarywords decode
		words := strings.Fields(strings.ToLower(line))
		var shareChunks []string
		for i := 0; i < len(words); i += 24 {
			end := i + 24
			if end > len(words) {
				end = len(words)
			}
			chunkWords := words[i:end]
			data, err := binarywords.Decode(chunkWords)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error decoding words: %v\n", err)
				os.Exit(1)
			}
			shareChunk := base64.URLEncoding.EncodeToString(data)
			shareChunks = append(shareChunks, shareChunk)
		}
		shares = append(shares, strings.Join(shareChunks, ""))
	}

	if len(shares) < 2 {
		fmt.Fprintln(os.Stderr, "Need at least 2 shares to decode.")
		os.Exit(1)
	}

	secret, err := sssa.Combine(shares)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error combining shares:", err)
		os.Exit(1)
	}

	fmt.Println("Recovered password:", secret)
}

//go:embed share.tpl.html
var htmlTemplate string

var tmpl = template.Must(template.New("share").Parse(htmlTemplate))

type tmplData struct {
	Index int
	Words []string
	Share string
}
