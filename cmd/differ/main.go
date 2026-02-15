package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jbonatakis/differ/internal/classify"
	"github.com/jbonatakis/differ/internal/config"
	"github.com/jbonatakis/differ/internal/filter"
	"github.com/jbonatakis/differ/internal/gitdiff"
	"github.com/jbonatakis/differ/internal/output"
	"github.com/jbonatakis/differ/internal/parser"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X main.Version=..."
var Version string

// Exit codes per spec.
const (
	exitSuccess       = 0
	exitRuntimeError  = 1
	exitInvalidConfig = 2
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(exitRuntimeError)
	}
}

func newRootCmd() *cobra.Command {
	var (
		base     string
		head     string
		empty    string
		list     bool
		listOnly bool
		format   string
		include  []string
		exclude  []string
		category []string
		sort     string
		noColor  bool
	)

	cmd := &cobra.Command{
		Use:   "differ [rev-range] [flags] [-- pathspec...]",
		Short: "Git-aware line-of-code churn reporting",
		Long: `differ is a Git-aware CLI tool that reports line-of-code churn between two refs,
grouped into practical categories (docs, tests, source, generated, other).

Examples:
  differ                                          # auto-detect base ref
  differ main...HEAD                              # explicit rev-range
  differ --base main --head feature/my-branch     # explicit refs
  differ --empty include -l                       # include empty lines, show file list
  differ --format json --exclude 'vendor/**'      # JSON output, exclude vendor
  differ -- docs/ internal/                       # restrict to pathspecs`,
		Args: cobra.ArbitraryArgs,
		// Silence default Cobra error/usage printing so we control exit codes.
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, runOpts{
				base:     base,
				head:     head,
				empty:    empty,
				list:     list,
				listOnly: listOnly,
				format:   format,
				include:  include,
				exclude:  exclude,
				category: category,
				sort:     sort,
				noColor:  noColor,
				runner:   gitdiff.DefaultRunner,
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&base, "base", "", "base ref")
	flags.StringVar(&head, "head", "", "head ref")
	flags.StringVar(&empty, "empty", "exclude", "count empty/whitespace-only changed lines (include|exclude)")
	flags.BoolVarP(&list, "list", "l", false, "show summary plus per-file list")
	flags.BoolVarP(&listOnly, "list-only", "L", false, "show per-file list only")
	flags.StringVar(&format, "format", "text", "output format (text|json)")
	flags.StringArrayVar(&include, "include", nil, "include path glob (repeatable)")
	flags.StringArrayVar(&exclude, "exclude", nil, "exclude path glob (repeatable)")
	flags.StringArrayVar(&category, "category", nil, "restrict to category (docs|tests|source|generated|other, repeatable)")
	flags.StringVar(&sort, "sort", "churn", "file list ordering (churn|path)")
	flags.BoolVar(&noColor, "no-color", false, "disable colorized text output")

	return cmd
}

type runOpts struct {
	base     string
	head     string
	empty    string
	list     bool
	listOnly bool
	format   string
	include  []string
	exclude  []string
	category []string
	sort     string
	noColor  bool
	runner   gitdiff.CommandRunner
}

func run(cmd *cobra.Command, args []string, opts runOpts) error {
	// Validate --empty flag value.
	if opts.empty != "include" && opts.empty != "exclude" {
		fmt.Fprintf(os.Stderr, "Error: --empty must be 'include' or 'exclude', got %q\n", opts.empty)
		os.Exit(exitInvalidConfig)
	}

	// Validate --format flag value.
	if opts.format != "text" && opts.format != "json" {
		fmt.Fprintf(os.Stderr, "Error: --format must be 'text' or 'json', got %q\n", opts.format)
		os.Exit(exitInvalidConfig)
	}

	// Validate --sort flag value.
	if opts.sort != "churn" && opts.sort != "path" {
		fmt.Fprintf(os.Stderr, "Error: --sort must be 'churn' or 'path', got %q\n", opts.sort)
		os.Exit(exitInvalidConfig)
	}

	// Split args into rev-range (before --) and pathspecs (after --).
	var revRange string
	var pathspecs []string
	dashIdx := cmd.ArgsLenAtDash()
	if dashIdx >= 0 {
		if dashIdx > 1 {
			fmt.Fprintln(os.Stderr, "Error: at most one positional rev-range argument allowed")
			os.Exit(exitRuntimeError)
		}
		if dashIdx == 1 {
			revRange = args[0]
		}
		pathspecs = args[dashIdx:]
	} else {
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Error: at most one positional rev-range argument allowed")
			os.Exit(exitRuntimeError)
		}
		if len(args) == 1 {
			revRange = args[0]
		}
	}

	autoRefMode := opts.base == "" && opts.head == "" && revRange == ""
	worktreeMode := false

	// 1. Load config with CLI overrides.
	cliOverrides := config.Config{
		Include: opts.include,
		Exclude: opts.exclude,
		Empty:   opts.empty,
		Sort:    opts.sort,
	}

	// Determine repo root for config loading.
	repoRoot, _ := os.Getwd()
	cfg, err := config.Load(repoRoot, cliOverrides)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: loading config: %v\n", err)
		os.Exit(exitInvalidConfig)
	}

	// 2. Resolve refs.
	refRange, err := gitdiff.ResolveRefs(opts.runner, opts.base, opts.head, revRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitRuntimeError)
	}

	// In auto mode, prefer showing local edits when the working tree is dirty by
	// diffing from merge-base to the current worktree.
	if autoRefMode {
		if dirty, err := gitdiff.WorktreeDirty(opts.runner); err == nil && dirty {
			baseRef, headRef := parseRefRange(refRange)
			if baseRef != "" && headRef != "" {
				if mergeBase, err := gitdiff.MergeBase(opts.runner, baseRef, headRef); err == nil {
					refRange = mergeBase
					worktreeMode = true
				}
			}
		}
	}

	// 3. Run git diff.
	diffResult, err := gitdiff.RunDiff(opts.runner, refRange, pathspecs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: running git diff: %v\n", err)
		os.Exit(exitRuntimeError)
	}

	// 4. Parse diff output.
	parsed, err := parser.Parse(diffResult.Stdout, cfg.Empty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: parsing diff: %v\n", err)
		os.Exit(exitRuntimeError)
	}

	if err := diffResult.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitRuntimeError)
	}

	// 5. Classify files.
	classifier := classify.New(cfg)

	// 6. Filter.
	filterCfg := filter.FilterConfig{
		Include:    cfg.Include,
		Exclude:    cfg.Exclude,
		Categories: opts.category,
	}
	filtered := filter.Filter(parsed, filterCfg, func(path string) string {
		cat, _ := classifier.Classify(path)
		return cat
	})

	// 7. Build summary.
	// Parse base and head from refRange for meta.
	metaBase, metaHead := parseRefRange(refRange)
	if worktreeMode {
		metaHead = "WORKTREE"
	}

	fileStats := make([]output.FileStat, 0, len(filtered))
	catTotals := make(map[string]output.CategoryTotal)

	var totalAdded, totalDeleted, totalFiles int
	for _, fs := range filtered {
		cat, lang := classifier.Classify(fs.Path)
		fileStats = append(fileStats, output.FileStat{
			Path:     fs.Path,
			Added:    fs.Added,
			Deleted:  fs.Deleted,
			Churn:    fs.Churn,
			Category: cat,
			Language: lang,
		})

		ct := catTotals[cat]
		ct.Added += fs.Added
		ct.Deleted += fs.Deleted
		ct.Churn += fs.Churn
		ct.FileCount++
		catTotals[cat] = ct

		totalAdded += fs.Added
		totalDeleted += fs.Deleted
		totalFiles++
	}

	summary := output.Summary{
		Totals: output.CategoryTotal{
			Added:     totalAdded,
			Deleted:   totalDeleted,
			Churn:     totalAdded + totalDeleted,
			FileCount: totalFiles,
		},
		CategoryTotals: catTotals,
		FileStats:      fileStats,
		Meta: output.Meta{
			Base:      metaBase,
			Head:      metaHead,
			Empty:     cfg.Empty,
			Pathspecs: pathspecs,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	// 8. Render output.
	if opts.format == "json" {
		if err := output.RenderJSON(os.Stdout, summary); err != nil {
			fmt.Fprintf(os.Stderr, "Error: rendering JSON: %v\n", err)
			os.Exit(exitRuntimeError)
		}
	} else {
		output.RenderText(os.Stdout, summary, output.OutputOpts{
			List:     opts.list,
			ListOnly: opts.listOnly,
			Sort:     cfg.Sort,
			NoColor:  opts.noColor,
		})
	}

	return nil
}

// parseRefRange splits "base...head" into base and head parts.
func parseRefRange(refRange string) (string, string) {
	if parts := strings.SplitN(refRange, "...", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	if parts := strings.SplitN(refRange, "..", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return refRange, ""
}
