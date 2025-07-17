package main

import (
	"flag"
	"fmt"
	"os"

	"viter/internal/viter"

	"github.com/spf13/afero"
)

func main() {
	var path string
	var create bool
	var help bool

	flag.StringVar(&path, "path", ".", "Path to the book directory")
	flag.BoolVar(&create, "create", false, "Create a new book")
	flag.BoolVar(&help, "help", false, "Print the help message")
	flag.Parse()

	if help {
		printHelp()
	} else if create {
		createBook(path)
	} else {
		printHelp()
	}
}

func createBook(path string) {
	fs := afero.NewOsFs()

	_, err := viter.CreateBook(fs, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating book: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created book at: %s\n", path)
}

func printHelp() {
	fmt.Println("Viter - A book writing tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  viter [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -path string")
	fmt.Println("        Path to the book directory (default \".\")")
	fmt.Println("  -create")
	fmt.Println("        Create a new book")
	fmt.Println("  -help")
	fmt.Println("        Print this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  viter -create -path ./my-book")
	fmt.Println("  viter -create")
	fmt.Println("  viter")
}
