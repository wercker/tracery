package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"runtime/pprof"
	"syscall"

	"github.com/wercker/tracery/mockery"
	"github.com/wercker/tracery/tracery"
)

const regexMetadataChars = "\\.+*?()|[]{}^$"

type Config struct {
	fName      string
	fTemplate  string
	fPrint     bool
	fOutput    string
	fOutpkg    string
	fDir       string
	fRecursive bool
	fAll       bool
	fIP        bool
	fTO        bool
	fCase      string
	fNote      string
	fProfile   string
	fVersion   bool
	quiet      bool
}

func main() {
	config := parseConfigFromArgs(os.Args)

	var recursive bool
	var filter *regexp.Regexp
	var err error
	var limitOne bool

	if config.quiet {
		// if "quiet" flag is set, set os.Stdout to /dev/null to suppress all output to Stdout
		os.Stdout = os.NewFile(uintptr(syscall.Stdout), os.DevNull)
	}

	if config.fVersion {
		fmt.Println(mockery.SemVer)
		return
	} else if config.fName != "" && config.fAll {
		fmt.Fprintln(os.Stderr, "Specify -name or -all, but not both")
		os.Exit(1)
	} else if config.fName != "" {
		recursive = config.fRecursive
		if strings.ContainsAny(config.fName, regexMetadataChars) {
			if filter, err = regexp.Compile(config.fName); err != nil {
				fmt.Fprintln(os.Stderr, "Invalid regular expression provided to -name")
				os.Exit(1)
			}
		} else {
			filter = regexp.MustCompile(fmt.Sprintf("^%s$", config.fName))
			limitOne = true
		}
	} else if config.fAll {
		recursive = true
		filter = regexp.MustCompile(".*")
	} else {
		fmt.Fprintln(os.Stderr, "Use -name to specify the name of the interface or -all for all interfaces found")
		os.Exit(1)
	}

	if config.fProfile != "" {
		f, err := os.Create(config.fProfile)
		if err != nil {
			os.Exit(1)
			return
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var osp tracery.OutputStreamProvider
	if config.fPrint {
		osp = &tracery.StdoutStreamProvider{}
	} else {
		osp = &tracery.FileOutputStreamProvider{
			BaseDir:   config.fOutput,
			InPackage: config.fIP,
			TestOnly:  config.fTO,
			Case:      config.fCase,
		}
	}

	tmpl := template.Must(tracery.GetTemplate(config.fTemplate))

	visitor := &tracery.GeneratorVisitor{
		InPackage:   config.fIP,
		Note:        config.fNote,
		Osp:         osp,
		PackageName: config.fOutpkg,
		Template:    tmpl,
	}

	walker := tracery.Walker{
		BaseDir:   config.fDir,
		Recursive: recursive,
		Filter:    filter,
		LimitOne:  limitOne,
	}

	generated := walker.Walk(visitor)

	if config.fName != "" && !generated {
		fmt.Printf("Unable to find %s in any go files under this path\n", config.fName)
		os.Exit(1)
	}
}

func parseConfigFromArgs(args []string) Config {
	config := Config{}

	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)

	flagSet.StringVar(&config.fName, "name", "", "name or matching regular expression of interface to generate mock for")
	flagSet.StringVar(&config.fTemplate, "tmpl", "template.go", "which template to use")
	flagSet.BoolVar(&config.fPrint, "print", false, "print the generated mock to stdout")
	flagSet.StringVar(&config.fOutput, "output", "./mocks", "directory to write mocks to")
	flagSet.StringVar(&config.fOutpkg, "outpkg", "mocks", "name of generated package")
	flagSet.StringVar(&config.fDir, "dir", ".", "directory to search for interfaces")
	flagSet.BoolVar(&config.fRecursive, "recursive", false, "recurse search into sub-directories")
	flagSet.BoolVar(&config.fAll, "all", false, "generates mocks for all found interfaces in all sub-directories")
	flagSet.BoolVar(&config.fIP, "inpkg", false, "generate a mock that goes inside the original package")
	flagSet.BoolVar(&config.fTO, "testonly", false, "generate a mock in a _test.go file")
	flagSet.StringVar(&config.fCase, "case", "camel", "name the mocked file using casing convention")
	flagSet.StringVar(&config.fNote, "note", "", "comment to insert into prologue of each generated file")
	flagSet.StringVar(&config.fProfile, "cpuprofile", "", "write cpu profile to file")
	flagSet.BoolVar(&config.fVersion, "version", false, "prints the installed version of mockery")
	flagSet.BoolVar(&config.quiet, "quiet", false, "suppress output to stdout")

	flagSet.Parse(args[1:])

	return config
}
