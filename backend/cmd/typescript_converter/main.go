package typescript_converter

import (
	"flag"
	"go/ast"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gzuidhof/tygo/tygo"
	"golang.org/x/tools/go/packages"
)

const (
	packagePath    = "github.com/gobitfly/beaconchain/pkg/api/types"
	fallbackType   = "any"
	commonFileName = "common"
	lintDisable    = "/* eslint-disable */\n"
	goFileSuffix   = ".go"
	tsFileSuffix   = ".ts"
)

// Files that should not be converted to TypeScript
var ignoredFiles = []string{"data_access", "search_types", "archiver"}

var typeMappings = map[string]string{
	"decimal.Decimal": "string /* decimal.Decimal */",
	"time.Time":       "string /* time.Time */",
}

// Expects the following flags:
// -out: Output folder for the generated TypeScript file

// Standard usage (execute in backend folder): go run cmd/main.go typescript-converter -out ../frontend/types/api

func Run() {
	var out string
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.StringVar(&out, "out", "", "Output folder for the generated TypeScript file")
	_ = fs.Parse(os.Args[2:])

	if out == "" {
		log.Fatal(nil, "Output folder not provided", 0)
	}

	if !strings.HasSuffix(out, "/") {
		out += "/"
	}

	// delete everything in the output folder
	err := deleteFiles(out)
	if err != nil {
		log.Fatal(err, "Failed to delete files in output folder", 0)
	}

	// Load package
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}, packagePath)

	if err != nil {
		log.Fatal(err, "Failed to load package", 0)
	}
	if packages.PrintErrors(pkgs) > 0 {
		log.Fatal(nil, "Failed to load package", 0)
	}

	// Find all common types, i.e. types that are used in multiple files and must be imported in ts
	commonTypes := getCommonTypes(pkgs)
	// Find imports (usages of common types) for each file
	imports := getImports(pkgs, commonTypes)

	var configs []*tygo.Tygo
	// Generate Tygo config for each file
	for fileName, typesUsed := range imports {
		var importStr string
		if len(typesUsed) > 0 {
			importStr = "import type { " + strings.Join(typesUsed, ", ") + " } from './" + commonFileName + "'\n"
		}
		configs = append(configs, tygo.New(getTygoConfig(out, fileName, importStr)))
	}

	// Generate TypeScript
	for _, tygo := range configs {
		err := tygo.Generate()
		if err != nil {
			log.Fatal(err, "Failed to generate TypeScript", 0)
		}
	}

	log.Infof("Juhu!")
}

func deleteFiles(out string) error {
	files, err := filepath.Glob(out + "*" + tsFileSuffix)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func getTygoConfig(outDir, fileName, frontmatter string) *tygo.Config {
	return &tygo.Config{
		Packages: []*tygo.PackageConfig{
			{
				Path:         packagePath,
				TypeMappings: typeMappings,
				FallbackType: fallbackType,
				IncludeFiles: []string{fileName + goFileSuffix},
				OutputPath:   outDir + fileName + tsFileSuffix,
				Frontmatter:  lintDisable + frontmatter,
			},
		},
	}
}

// Iterate over all file names and files in the packages
func allFiles(pkgs []*packages.Package) iter.Seq2[string, *ast.File] {
	return func(yield func(string, *ast.File) bool) {
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				fileName := filepath.Base(pkg.Fset.File(file.Pos()).Name())
				if !yield(fileName, file) {
					return
				}
			}
		}
	}
}

// Parse common.go to find all common types
func getCommonTypes(pkgs []*packages.Package) map[string]struct{} {
	var commonFile *ast.File
	// find common file
	for fileName, file := range allFiles(pkgs) {
		fileName = strings.TrimSuffix(fileName, goFileSuffix)
		if filepath.Base(fileName) == commonFileName {
			commonFile = file
			break
		}
	}
	if commonFile == nil {
		log.Fatal(nil, "common.go not found", 0)
	}
	commonTypes := make(map[string]struct{})
	// iterate over all types in common file and add them to the map
	for node := range ast.Preorder(commonFile) {
		if typeSpec, ok := node.(*ast.TypeSpec); ok {
			commonTypes[typeSpec.Name.Name] = struct{}{}
		}
	}
	return commonTypes
}

// Parse all files to find used common types for each file
// Returns a map with file name as key and a set of common types used in the file as value
func getImports(pkgs []*packages.Package, commonTypes map[string]struct{}) map[string][]string {
	imports := make(map[string][]string) // Map from file to set of commonTypes used
	imports[commonFileName] = []string{} // Add common file to map with empty set
	for fileName, file := range allFiles(pkgs) {
		fileName = strings.TrimSuffix(fileName, goFileSuffix)
		if filepath.Base(fileName) == commonFileName || slices.Contains(ignoredFiles, fileName) {
			continue
		}
		var currentFileImports []string
		// iterate over all struct fields in the file
		for node := range ast.Preorder(file) {
			ident, ok := node.(*ast.Ident)
			if !ok {
				continue
			}
			_, isCommonType := commonTypes[ident.Name]
			if isCommonType && !slices.Contains(currentFileImports, ident.Name) {
				currentFileImports = append(currentFileImports, ident.Name)
			}
		}
		imports[fileName] = currentFileImports
	}
	return imports
}
