package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"viter/internal/neural"
	"viter/internal/prompts"
	"viter/internal/viter"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func NewPrompts(lang string) viter.PromptProvider {
	// If lang is empty, detect OS locale
	if lang == "" {
		if locale := os.Getenv("LC_ALL"); locale != "" {
			lang = strings.SplitN(locale, ".", 2)[0]
		} else if locale := os.Getenv("LC_MESSAGES"); locale != "" {
			lang = strings.SplitN(locale, ".", 2)[0]
		} else if locale := os.Getenv("LANG"); locale != "" {
			lang = strings.SplitN(locale, ".", 2)[0]
		} else {
			lang = "en" // fallback to English
		}

		// Extract language code (first part before underscore)
		if strings.Contains(lang, "_") {
			lang = strings.SplitN(lang, "_", 2)[0]
		}
	}

	var pp viter.PromptProvider
	switch lang {
	case "ru":
		pp = prompts.NewRuPrompts()
	default:
		pp = prompts.NewEnPrompts()
	}

	return pp
}

func main() {
	var path, lang string
	var configPath string
	var debug bool
	var simple string
	var create, meta, chapters, book, correctall bool
	var md, html, epub, archive bool
	var plan, chapter, iterCount, maxLen, correct int
	var help bool

	flag.StringVar(&path, "path", ".", "Path to the book directory")
	flag.StringVar(&lang, "lang", "", "Language of the book")
	flag.StringVar(&configPath, "config", "viter.toml", "Path to the configuration file")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.StringVar(&simple, "simple", "", "Generate the whole book using the given prompt")
	flag.BoolVar(&create, "create", false, "Create a new book")
	flag.BoolVar(&meta, "meta", false, "Update book metadata")
	flag.IntVar(&plan, "plan", 0, "Update book plan with given count of chapters")
	flag.IntVar(&chapter, "chapter", 0, "Update chapter with given number")
	flag.BoolVar(&chapters, "chapters", false, "Update all book chapters")
	flag.BoolVar(&book, "book", false, "Update the whole book")
	flag.BoolVar(&correctall, "correctall", false, "Correct all chapters")
	flag.IntVar(&correct, "correct", 0, "Correct chapter with given number")
	flag.IntVar(&iterCount, "iter", 5, "Update given number of times")
	flag.IntVar(&maxLen, "maxlen", 10000, "Maximum length of a text part to correct")
	flag.BoolVar(&archive, "archive", false, "Make an archive copy of each file version")
	flag.BoolVar(&md, "md", false, "Export to markdown file")
	flag.BoolVar(&html, "html", false, "Export to html file")
	flag.BoolVar(&epub, "epub", false, "Export to epub file")
	flag.BoolVar(&help, "help", false, "Print the help message")
	flag.Parse()

	if help {
		printHelp()
		return
	}

	if simple != "" {
		create = true
		meta = true
		chapters = true
		book = true
		if plan == 0 {
			plan = 10
		}
	}

	if !create && !meta && !chapters && !book && !correctall && plan == 0 && chapter == 0 && correct == 0 && !md && !html && !epub {
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

	pp := NewPrompts(lang)
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
	if simple != "" {
		v.SetPlot(simple)
	}
	if debug || archive {
		v.EnableArchiving()
	}

	if meta {
		ok = v.UpdateMeta(iterCount)
		if !ok {
			return
		}
	}

	if plan > 0 {
		ok = v.UpdatePlan(plan, iterCount)
		if !ok {
			return
		}
	}

	if chapter > 0 {
		ok = v.UpdateChapter(chapter, iterCount)
		if !ok {
			return
		}
	}

	if chapters {
		ok = v.UpdateAllChapters(iterCount)
		if !ok {
			return
		}
	}

	if book {
		ok = v.UpdateBook(iterCount)
		if !ok {
			return
		}
	}

	if correct > 0 {
		ok = v.CorrectChapter(correct, maxLen)
		if !ok {
			return
		}
	}

	if correctall {
		ok = v.CorrectAllChapters(maxLen)
		if !ok {
			return
		}
	}

	if md {
		ok = v.ExportMarkdown()
		if !ok {
			return
		}
	}

	if html {
		ok = v.ExportHTML()
		if !ok {
			return
		}
	}

	if epub {
		ok = v.ExportEPUB()
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
	fmt.Println("        Language of the book (defaults to OS language)")
	fmt.Println("  -simple string")
	fmt.Println("        Generate the whole book using the given prompt")
	fmt.Println("  -create")
	fmt.Println("        Create a new book")
	fmt.Println("  -meta")
	fmt.Println("        Update book metadata")
	fmt.Println("  -plan int")
	fmt.Println("        Update book plan with given count of chapters")
	fmt.Println("  -chapter int")
	fmt.Println("        Update chapter with given number")
	fmt.Println("  -chapters")
	fmt.Println("        Update all book chapters")
	fmt.Println("  -book")
	fmt.Println("        Update the whole book")
	fmt.Println("  -correct int")
	fmt.Println("        Correct chapter with given number")
	fmt.Println("  -correctall")
	fmt.Println("        Correct all chapters")
	fmt.Println("  -iter int")
	fmt.Println("        Update given number of times (default 5)")
	fmt.Println("  -maxlen int")
	fmt.Println("        Maximum length of a text part to correct (default 10000)")
	fmt.Println("  -archive")
	fmt.Println("        Make an archive copy of each file version")
	fmt.Println("  -epub")
	fmt.Println("        Export to epub file")
	fmt.Println("  -md")
	fmt.Println("        Export to markdown file")
	fmt.Println("  -html")
	fmt.Println("        Export to html file")
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
