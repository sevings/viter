package main

import (
	"flag"
	"fmt"

	"viter/internal/neural"
	"viter/internal/prompts"
	"viter/internal/viter"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	var path, lang string
	var configPath string
	var debug bool
	var create, meta bool
	var plan, chapter, score int
	var help bool

	flag.StringVar(&path, "path", ".", "Path to the book directory")
	flag.StringVar(&lang, "lang", "en", "Language of the book")
	flag.StringVar(&configPath, "config", "viter.toml", "Path to the configuration file")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&create, "create", false, "Create a new book")
	flag.BoolVar(&meta, "meta", false, "Update book metadata")
	flag.IntVar(&plan, "plan", 0, "Update book plan with given count of chapters")
	flag.IntVar(&chapter, "chapter", 0, "Update chapter with given number")
	flag.IntVar(&score, "score", 99, "Target score of the book")
	flag.BoolVar(&help, "help", false, "Print the help message")
	flag.Parse()

	if help {
		printHelp()
		return
	}

	if !create && !meta && plan == 0 && chapter == 0 {
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
	tg, ok := neural.NewLLM(cfg.Ai)
	if !ok {
		return
	}
	if debug {
		tg.EnableResponseLogging(fs, path)
	}

	var pp viter.PromptProvider
	switch lang {
	case "ru":
		pp = prompts.NewRuPrompts()
	default:
		pp = prompts.NewEnPrompts()
	}

	v, ok := viter.NewViter(cfg, pp, tg)
	if !ok {
		return
	}

	if create {
		ok = v.CreateBook(fs, path)
	} else {
		ok = v.LoadBook(fs, path)
	}
	if !ok {
		return
	}

	if meta {
		ok = v.UpdateMeta(score)
		if !ok {
			return
		}
	}

	if plan > 0 {
		ok = v.UpdatePlan(plan, score)
		if !ok {
			return
		}
	}

	if chapter > 0 {
		ok = v.UpdateChapter(chapter, score)
		if !ok {
			return
		}
	}
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
	fmt.Println("  -lang string")
	fmt.Println("        Language of the book (default \"en\")")
	fmt.Println("  -create")
	fmt.Println("        Create a new book")
	fmt.Println("  -meta")
	fmt.Println("        Update book metadata")
	fmt.Println("  -plan int")
	fmt.Println("        Update book plan with given count of chapters")
	fmt.Println("  -chapter int")
	fmt.Println("        Update chapter with given number")
	fmt.Println("  -score int")
	fmt.Println("        Target score of the book")
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
