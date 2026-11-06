package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Laca1234/lca-compressor/internal/controller"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "compress":
		compressCmd := flag.NewFlagSet("compress", flag.ExitOnError)
		inputFlag := compressCmd.String("input", "", "Path to input file")
		outputFlag := compressCmd.String("output", "", "Path to output .lca file")
		threadsFlag := compressCmd.Int("threads", 4, "Number of parallel worker threads")

		_ = compressCmd.Parse(os.Args[2:])

		if *inputFlag == "" || *outputFlag == "" {
			fmt.Println("Error: --input and --output are required")
			os.Exit(1)
		}

		inFile, err := os.Open(*inputFlag)
		if err != nil {
			fmt.Printf("Failed to open input file: %v\n", err)
			os.Exit(1)
		}
		defer inFile.Close()

		outFile, err := os.Create(*outputFlag)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		pipe := controller.NewPipeline(1024*1024, *threadsFlag)
		fmt.Printf("Compressing %s -> %s using %d threads...\n", *inputFlag, *outputFlag, *threadsFlag)
		
		if err := pipe.CompressStream(inFile, outFile); err != nil {
			fmt.Printf("Compression failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Compression finished successfully!")

	case "decompress":
		decompressCmd := flag.NewFlagSet("decompress", flag.ExitOnError)
		inputFlag := decompressCmd.String("input", "", "Path to input .lca file")
		outputFlag := decompressCmd.String("output", "", "Path to output restored file")

		_ = decompressCmd.Parse(os.Args[2:])

		if *inputFlag == "" || *outputFlag == "" {
			fmt.Println("Error: --input and --output are required")
			os.Exit(1)
		}

		inFile, err := os.Open(*inputFlag)
		if err != nil {
			fmt.Printf("Failed to open input file: %v\n", err)
			os.Exit(1)
		}
		defer inFile.Close()

		outFile, err := os.Create(*outputFlag)
		if err != nil {
			fmt.Printf("Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		pipe := controller.NewPipeline(1024*1024, 1)
		fmt.Printf("Decompressing %s -> %s...\n", *inputFlag, *outputFlag)
		
		if err := pipe.DecompressStream(inFile, outFile); err != nil {
			fmt.Printf("Decompression failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Decompression finished successfully!")

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: lca <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  compress   --input <file> --output <file.lca> [--threads N]")
	fmt.Println("  decompress --input <file.lca> --output <file>")
}
