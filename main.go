package main

import (
	"flag"
	"fmt"

	"viter/internal/viter"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	var path string
	var configPath string
	var debug bool
	var create bool
	var help bool

	flag.StringVar(&path, "path", ".", "Path to the book directory")
	flag.StringVar(&configPath, "config", "viter.toml", "Path to the configuration file")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&create, "create", false, "Create a new book")
	flag.BoolVar(&help, "help", false, "Print the help message")
	flag.Parse()

	if help {
		printHelp()
		return
	}

	if !create {
		printHelp()
		return
	}

	var zapLogger *zap.Logger
	var err error
	if debug {
		zapLogger, err = zap.NewDevelopment(zap.WithCaller(false))
	} else {
		zapLogger, err = zap.NewProduction(zap.WithCaller(false))
	}
	if err != nil {
		panic(err)
	}
	defer func() { _ = zapLogger.Sync() }()

	zap.ReplaceGlobals(zapLogger)
	zap.RedirectStdLog(zapLogger)
	logger := zapLogger.Sugar()

	cfg, err := viter.LoadConfig(configPath)
	if err != nil {
		logger.Errorw("Error loading config", "err", err)
		return
	}

	fs := afero.NewOsFs()
	v, ok := viter.NewViter(cfg)
	if !ok {
		return
	}

	ok = v.CreateBook(fs, path)
}

func printHelp() {
	fmt.Println("Viter - A book writing tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  viter [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -config string")
	fmt.Println("        Path to the configuration file (default \"viter.toml\")")
	fmt.Println("  -path string")
	fmt.Println("        Path to the book directory (default \".\")")
	fmt.Println("  -create")
	fmt.Println("        Create a new book")
	fmt.Println("  -debug")
	fmt.Println("        Enable debug mode")
	fmt.Println("  -help")
	fmt.Println("        Print this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  viter -create -path ./my-book")
	fmt.Println("  viter -create")
	fmt.Println("  viter")
}
