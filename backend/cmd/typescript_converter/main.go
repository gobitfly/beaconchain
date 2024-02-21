package main

import (
	"flag"
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gzuidhof/tygo/tygo"
	"golang.org/x/tools/go/packages"
)

const (
	packagePath    = "github.com/gobitfly/beaconchain/pkg/api/types"
	fallbackType   = "any"
	commonFileName = "common"
	indent         = "    "
)

var typeMappings = map[string]string{
	"decimal.Decimal": "string",
}

// Expects the following flags:
// -out: Output folder for the generated TypeScript file

// Standard usage (execute in backend folder): go run cmd/typescript-converter/main.go -out ../frontend/types/api

func main() {
	var out string
	flag.StringVar(&out, "out", "", "Output folder for the generated TypeScript file")
	flag.Parse()

	if out == "" {
		utils.LogFatal(nil, "Output folder not provided", 0)
	}

	if !strings.HasSuffix(out, "/") {
		out += "/"
	}

	// Load package
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}, packagePath)

	if err != nil {
		utils.LogFatal(err, "Failed to load package", 0)
	}
	if packages.PrintErrors(pkgs) > 0 {
		utils.LogFatal(nil, "Failed to load package", 0)
	}

	// Find all common types and their usages
	commonTypes := getCommonTypes(pkgs)
	// Find all usages of common types
	usage := getCommonUsages(pkgs, commonTypes)

	tygos := []*tygo.Tygo{}
	// Generate Tygo for common.go
	tygos = append(tygos, tygo.New(getTygoConfig(out, commonFileName, "")))
	// Generate Tygo for each file
	for file, typesUsed := range usage {
		importStr := "import { " + strings.Join(typesUsed, ", ") + " } from './" + commonFileName + ".ts';\n\n"
		tygos = append(tygos, tygo.New(getTygoConfig(out, file, importStr)))
	}

	// Generate TypeScript
	for _, tygo := range tygos {
		err := tygo.Generate()
		if err != nil {
			utils.LogFatal(err, "Failed to generate TypeScript", 0)
		}
	}
}

func getTygoConfig(out, file, frontmatter string) *tygo.Config {
	return &tygo.Config{
		Packages: []*tygo.PackageConfig{
			{
				Path:         packagePath,
				TypeMappings: typeMappings,
				FallbackType: fallbackType,
				IncludeFiles: []string{file + ".go"},
				OutputPath:   out + file + ".ts",
				Frontmatter:  frontmatter,
				Indent:       indent,
			},
		},
	}
}

// Parse common.go to find all common types
func getCommonTypes(pkgs []*packages.Package) map[string]bool {
	commonTypes := make(map[string]bool)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			filename := strings.TrimSuffix(filepath.Base(pkg.Fset.File(file.Pos()).Name()), ".go")
			if filepath.Base(filename) != commonFileName {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if typeSpec, ok := n.(*ast.TypeSpec); ok {
					commonTypes[typeSpec.Name.Name] = true
				}
				return true
			})
			return commonTypes
		}
	}
	return nil
}

// Parse all files to find which common types are for each file
func getCommonUsages(pkgs []*packages.Package, commonTypes map[string]bool) map[string][]string {
	usage := make(map[string][]string) // Map from file to list of commonTypes used
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			filename := strings.TrimSuffix(filepath.Base(pkg.Fset.File(file.Pos()).Name()), ".go")
			if filepath.Base(filename) == commonFileName {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				ident, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				if !commonTypes[ident.Name] {
					return true
				}
				if _, exists := usage[filename]; !exists {
					usage[filename] = make([]string, 0)
				}
				if !slices.Contains[[]string](usage[filename], ident.Name) {
					usage[filename] = append(usage[filename], ident.Name)
				}
				return true
			})
		}
	}
	return usage
}
