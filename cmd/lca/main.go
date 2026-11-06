package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Laca1234/lca-compressor/internal/controller"
)

func main() {
	compressCmd := flag.NewFlagSet("compress", flag.ExitOnError)
	compressInput := compressCmd.String("input", "", "Input file or directory to archive and compress")
	compressOutput := compressCmd.String("output", "", "Output .lca archive file")
	compressThreads := compressCmd.Int("threads", 4, "Number of parallel compression threads")

	decompressCmd := flag.NewFlagSet("decompress", flag.ExitOnError)
	decompressInput := decompressCmd.String("input", "", "Input .lca archive file")
	decompressOutput := decompressCmd.String("output", "", "Output directory to restore files into")
	decompressThreads := decompressCmd.Int("threads", 4, "Number of parallel decompression threads")

	if len(os.Args) < 2 {
		fmt.Println("Usage: lca <command> [arguments]")
		fmt.Println("  compress   Compress a file or directory into a .lca archive")
		fmt.Println("  decompress Decompress a .lca archive into a target directory")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compress":
		compressCmd.Parse(os.Args[2:])
		if *compressInput == "" || *compressOutput == "" {
			fmt.Println("Error: --input and --output flags are required.")
			compressCmd.Usage()
			os.Exit(1)
		}

		fmt.Printf("Archiving and compressing '%s' -> '%s' using %d threads...\n", *compressInput, *compressOutput, *compressThreads)

		outFile, err := os.Create(*compressOutput)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		// Set up a pipe to stream packaged tar data directly into the compressor
		pr, pw := io.Pipe()
		errChan := make(chan error, 1)

		go func() {
			defer pw.Close()
			err := controller.PackFolder(*compressInput, pw)
			if err != nil {
				pw.CloseWithError(err)
			}
			errChan <- err
		}()

		// Run your compression pipeline using the pipe reader as input
		err = controller.CompressStream(pr, outFile, *compressThreads)
		if err != nil {
			fmt.Printf("Compression failed: %v\n", err)
			os.Exit(1)
		}

		if packErr := <-errChan; packErr != nil {
			fmt.Printf("Archiving failed: %v\n", packErr)
			os.Exit(1)
		}

		fmt.Println("Archive compression finished successfully!")

	case "decompress":
		decompressCmd.Parse(os.Args[2:])
		if *decompressInput == "" || *decompressOutput == "" {
			fmt.Println("Error: --input and --output flags are required.")
			decompressCmd.Usage()
			os.Exit(1)
		}

		fmt.Printf("Decompressing and unpacking '%s' -> '%s'...\n", *decompressInput, *decompressOutput)

		inputFile, err := os.Open(*decompressInput)
		if err != nil {
			fmt.Printf("Failed to open input archive: %v\n", err)
			os.Exit(1)
		}
		defer inputFile.Close()

		// Set up a pipe to stream decompressed tar data directly into the unpacker
		pr, pw := io.Pipe()
		errChan := make(chan error, 1)

		go func() {
			defer pw.Close()
			err := controller.UnpackArchive(pr, *decompressOutput)
			if err != nil {
				pr.CloseWithError(err)
			}
			errChan <- err
		}()

		// Run your decompression pipeline, writing the raw tar stream into the pipe writer
		err = controller.DecompressStream(inputFile, pw, *decompressThreads)
		if err != nil {
			fmt.Printf("Decompression failed: %v\n", err)
			os.Exit(1)
		}

		if unpackErr := <-errChan; unpackErr != nil {
			fmt.Printf("Unpacking failed: %v\n", unpackErr)
			os.Exit(1)
		}

		fmt.Println("Decompression and restoration finished successfully!")

	default:
		fmt.Printf("Unknown subcommand '%s'\n", os.Args[1])
		os.Exit(1)
	}
}
