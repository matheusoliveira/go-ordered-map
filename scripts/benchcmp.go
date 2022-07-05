package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"bitbucket.org/creachadair/shell"
	"github.com/bitfield/script"
	"github.com/google/subcommands"
)

func getCurrentGitPos() (string, error) {
	rev, _ := script.Exec("git rev-parse --abbrev-ref HEAD").String()
	if rev == "HEAD" {
		rev, err := script.File(".git/HEAD").First(0).String()
		if err != nil || rev == "" {
			return "", errors.New("could not find current git pos")
		}
	}
	return strings.TrimSpace(rev), nil
}

func runBenchmark(tag string, take int, appendFile string, params *benchCmpCmd) error {
	outStr, err := script.Exec("git checkout " + shell.Quote(tag)).String()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w. Output: %q", err, outStr)
	}
	log.Printf("Running benchmark on version %q, take %d", tag, take)
	outStr, err = script.
		Exec("go test -bench="+shell.Quote(params.bench)+" -benchtime="+shell.Quote(params.benchtime)+" -benchmem ./...").
		Replace("github.com/matheusoliveira/go-ordered-map/pkg/", "github.com/matheusoliveira/go-ordered-map/").
		String()
	if err != nil {
		return fmt.Errorf("go test failed: %w. Output: %q", err, outStr)
	}
	_, err = script.Echo(outStr).AppendFile(appendFile)
	if err != nil {
		return err
	}
	return nil
}

type benchPackageVersion struct {
	version  string
	filename string
}

func execBenchCmp(versions []benchPackageVersion, params *benchCmpCmd) error {
	gitOriginalPos, err := getCurrentGitPos()
	if err != nil {
		log.Printf("could not get current git position, will continue without it. Error: %v", err)
	}

	// Remove output files
	for _, v := range versions {
		if !params.noRemoveBench {
			if err := script.IfExists(v.filename).Exec("rm " + shell.Quote(v.filename)).Close(); err != nil {
				return err
			}
		}
		log.Printf("%s: %s", v.version, v.filename)
	}

	// Run benchmarks
	for i := 1; i <= params.nruns; i++ {
		for _, v := range versions {
			if err := runBenchmark(v.version, i, v.filename, params); err != nil {
				return err
			}
		}
	}

	// Checkout git back to old position
	if err := script.Exec("git checkout " + shell.Quote(gitOriginalPos)).Close(); err != nil {
		log.Printf("Failed to checkout git back to %q: %v", gitOriginalPos, err)
	}

	// benchstat
	allVersions := ""
	for _, v := range versions {
		allVersions += "-" + v.version
	}
	benchstatOutput := shell.Quote(params.outputDir) + "/benchstat" + allVersions + ".txt"
	log.Printf("benchstat results %q", benchstatOutput)
	benchstatCmd := fmt.Sprintf("benchstat %s %s", shell.Quote(versions[0].filename), shell.Quote(versions[1].filename))
	if _, err := script.Exec(benchstatCmd).WriteFile(benchstatOutput); err != nil {
		return fmt.Errorf("benchstat failed: %w", err)
	}

	// Filter only lines with delta variation
	fmt.Println("\nFiltered high variance results only:")
	script.
		File(benchstatOutput).
		Reject(" ~ ").
		FilterLine(func(in string) string { return "    " + in }).
		Stdout()

	return nil
}

type benchCmpCmd struct {
	benchtime     string
	bench         string
	outputDir     string
	nruns         int
	noRemoveBench bool
}

func (*benchCmpCmd) Name() string {
	return "benchcmp"
}

func (*benchCmpCmd) Synopsis() string {
	return "Run benchmarks on different versions and compare results."
}

func (*benchCmpCmd) Usage() string {
	return "benchcmp [flags] <old-version> <new-version>\n"
}

func (p *benchCmpCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.bench, "bench", ".", "\"-bench\" flag for go test")
	f.StringVar(&p.benchtime, "benchtime", "2s", "\"-benchtime\" flag for go test")
	f.StringVar(&p.outputDir, "output-dir", os.TempDir(), "output directory, default is \"os.TempDir()\"")
	f.IntVar(&p.nruns, "nruns", 5, "number of benchmark runs")
	f.BoolVar(&p.noRemoveBench, "no-remove-bench", false, "do not remove bench*.txt files before running, useful to run only benchstat (e.g. with -nruns=0)")
}

func (p *benchCmpCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) < 2 {
		log.Printf("No version given.\n    Usage: %s", p.Usage())
		return subcommands.ExitUsageError
	}
	versions := make([]benchPackageVersion, len(f.Args()))
	for i, v := range f.Args() {
		versions[i] = benchPackageVersion{v, fmt.Sprintf("%s/bench-%s.txt", shell.Quote(p.outputDir), v)}
	}

	if err := execBenchCmp(versions, p); err != nil {
		log.Println(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
