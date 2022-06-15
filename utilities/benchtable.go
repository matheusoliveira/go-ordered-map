// +build exclude

// This is not part of the final package, this is just a script to build benchmarks.md doc file
// easily from the bench.txt output.
// Please, do not be critical about these line of codes, was just a simple way to execute with
// `go run benchtable.go` without any dependencies.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const importPath = "./pkg/omap/"

const introductionText =
`# Benchmarks

This is the result of a series of benchmarks on each implementation to validate the assumptions
about design decision of each. Bellow is a formatted version of the results with my own
conclusions, you can see the raw results at [bench.txt](bench.txt) file.

To run the benchmarks your self, just do

` + "```sh" + `
make bench
` + "```" + `

You can generate the formatted output as bellow with [utilities/benchtable.go], the command is provided
in the Makefile as well (must run after generating the ` + "`bench.txt`" + ` with previous command):

` + "```sh" + `
make doc-bench
` + "```" + `
`

func readComments() (map[string]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, importPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	omapPkg, ok := pkgs["omap_test"]
	if !ok {
		return nil, errors.New("package \"omap\" not found")
	}
	omapDoc := doc.New(omapPkg, importPath, doc.AllDecls)
	ret := make(map[string]string, len(omapDoc.Funcs))
	for _, fct := range omapDoc.Funcs {
		ret[fct.Name] = fct.Doc
	}
	return ret, nil
}

// Simple Markdown-like Table Builder
type AlignmentType int
const (
	AlignLeft AlignmentType = iota
	AlignRight
	AlignCenter
)

type TableColumn struct {
	Title      string
	Align AlignmentType
	VMerge     bool
	maxSize    int
}

type Table struct {
	Columns []TableColumn
	rows    [][]string
}

func (t *Table) AddRow(cells ...any) {
	if len(cells) != len(t.Columns) {
		for i := len(cells); i < len(t.Columns); i++ {
			t.Columns = append(t.Columns, TableColumn{})
		}
	}
	if t.rows == nil {
		t.rows = make([][]string, 1, 2)
		t.rows[0] = make([]string, len(t.Columns))
		for i, c := range t.Columns {
			if len(c.Title) > c.maxSize {
				c.maxSize = len(c.Title)
				t.Columns[i] = c
			}
			t.rows[0][i] = c.Title
		}
	}
	row := make([]string, len(cells))
	for i, cell := range cells {
		col := t.Columns[i]
		str := fmt.Sprint(cell)
		if len(str) > col.maxSize {
			col.maxSize = len(str)
			t.Columns[i] = col
		}
		row[i] = str
	}
	t.rows = append(t.rows, row)
}

func (t *Table) Write(w io.Writer) (int, error) {
	var n int
	format := ""
	for _, col := range t.Columns {
		format += "| "
		if col.Align == AlignRight {
			format += "%" + strconv.Itoa(col.maxSize) + "s "
		} else {
			format += "%-" + strconv.Itoa(col.maxSize) + "s "
		}
	}
	format += "|\n"
	for i, row := range t.rows {
		strAsAny := make([]any, len(row))
		if i == 1 {
			// dotted line after title
			for j, col := range t.Columns {
				if col.Align == AlignRight {
					strAsAny[j] = strings.Repeat("-", col.maxSize - 1) + ":"
				} else {
					strAsAny[j] = strings.Repeat("-", col.maxSize)
				}
			}
			ln, err := fmt.Fprintf(w, format, strAsAny...)
			n += ln
			if err != nil {
				return n, err
			}
		}
		for j, cell := range row {
			if t.Columns[j].VMerge && i > 1 && t.rows[i-1][j] == cell {
				strAsAny[j] = ""
			} else {
				strAsAny[j] = cell
			}
		}
		ln, err := fmt.Fprintf(w, format, strAsAny...)
		n += ln
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (t *Table) Erase() {
	t.rows = nil
}

type BenchLine struct {
	Name     string
	SubName  string
	NRuns    int
	NsOp     int
	BOp      int
	AllocsOp int
}

func fmtInt(i int) string {
	if i == 0 {
		return "0"
	}
	str := ""
	if i < 0 {
		str = "-"
		i *= -1
	}
	for i != 0 {
		div := i % 1000
		i = i / 1000
		var divStr string
		if i == 0 {
			divStr = fmt.Sprintf("%d", div)
		} else {
			divStr = fmt.Sprintf("%03d", div)
		}
		if str == "" {
			str = divStr
		} else {
			str = divStr + "," + str
		}
	}
	return str
}

func processBenchmark(tbl *Table, w io.Writer, benchs []BenchLine, doc map[string]string) {
	if len(benchs) == 0 {
		return
	}
	var conclusionComment string
	const conclusionTag = "\nConclusion: "
	fmt.Fprintf(w, "## Benchmark %s\n\n", benchs[0].Name)
	if doc != nil {
		name := benchs[0].Name
		if comm, ok := doc["Benchmark" + name]; ok {
			// split comment and info
			s := strings.Split(comm, conclusionTag)
			fmt.Fprintln(w, s[0])
			if len(s) > 1 {
				fmt.Fprintln(w)
				conclusionComment = conclusionTag + strings.Join(s[1:], conclusionTag)
			}
		}
	}
	baseline := 0
	for i, b := range benchs {
		var perfCmp string
		if i == 0 && (b.SubName == "map" || b.SubName == "Builtin") {
			baseline = b.NsOp
			perfCmp = "baseline"
		} else if baseline != 0 {
			perfCmpF := float64(baseline)/float64(b.NsOp) - 1
			perfCmp = fmt.Sprintf("%.2f %%", perfCmpF*100)
		}
		tbl.AddRow(/*b.Name,*/ b.SubName, fmtInt(b.NRuns), fmtInt(b.NsOp), fmtInt(b.BOp), fmtInt(b.AllocsOp), perfCmp)
	}
	tbl.Write(w)
	fmt.Fprintln(w, conclusionComment)
	tbl.Erase()
}

func getMainReaderWritter() (io.Reader, io.Writer, error) {
	// in
	inputFilename := "docs/bench.txt"
	if len(os.Args) >= 2 {
		inputFilename = os.Args[1]
	}
	inputReader, err := os.Open(inputFilename)
	if err != nil {
		return nil, nil, err
	}
	// out
	outputFilename := "docs/benchmarks.md"
	if len(os.Args) >= 3 {
		outputFilename = os.Args[2]
	}
	outputWriter, err := os.Create(outputFilename)
	if err != nil {
		return nil, nil, err
	}
	return inputReader, outputWriter, nil
}

func processLines(in io.Reader) <-chan []BenchLine {
	outCh := make(chan []BenchLine)
	// Bench output processor
	reIsBench := regexp.MustCompile("^Benchmark")
	reProcBench := regexp.MustCompile(`^Benchmark([^/]+)/([^-]+)-\d+\s+(\d+)((\s+(\d+)\s(\S+))*)$`)
	reProcParams := regexp.MustCompile(`\s*(\d+)\s(\S+)`)
	go func() {
		benchs := []BenchLine{}
		scanner := bufio.NewScanner(in)
		lastBench := ""
		for scanner.Scan() {
			line := scanner.Text()
			if !reIsBench.MatchString(line) {
				continue
			}
			matches := reProcBench.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				b := BenchLine{}
				b.Name = m[1]
				b.SubName = m[2]
				b.NRuns, _ = strconv.Atoi(m[3])
				params := m[4]
				pmatches := reProcParams.FindAllStringSubmatch(params, -1)
				for _, pm := range pmatches {
					v, _ := strconv.Atoi(pm[1])
					switch pm[2] {
					case "ns/op":
						b.NsOp = v
					case "B/op":
						b.BOp = v
					case "allocs/op":
						b.AllocsOp = v
					}
				}
				if lastBench != b.Name {
					outCh <- benchs
					benchs = []BenchLine{b}
				} else {
					benchs = append(benchs, b)
				}
				lastBench = b.Name
			}
		}
		if len(benchs) > 0 {
			outCh <- benchs
		}
		close(outCh)
	}()
	return outCh
}

func main() {
	var err error
	// in/out
	in, out, err := getMainReaderWritter()
	if err != nil {
		log.Fatal(err)
	}
	// Process documentation
	doc, err := readComments()
	if err != nil {
		log.Printf("Could not read documentation. Error: %v\n\n. Proceding without doc...", err)
	}
	// Introduction text
	fmt.Fprintln(out, introductionText)
	// Table columns setup
	tbl := Table{
		Columns: []TableColumn{
			//{Title: "Bench", VMerge: true},
			{Title: "Implemenation"},
			{Title: "Nruns", Align: AlignRight},
			{Title: "ns/op", Align: AlignRight},
			{Title: "B/op", Align: AlignRight},
			{Title: "allocs/op", Align: AlignRight},
			{Title: "% perf relative", Align: AlignRight},
		},
	}
	// Process lines
	for benchs := range processLines(in) {
		processBenchmark(&tbl, out, benchs, doc)
		fmt.Fprintln(out)
	}
}
