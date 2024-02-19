package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

type FileStructs struct {
	FileName    string
	StructNames []string
}

// Expects the following flags:
// -in: Path to the Go package containing the API types
// -out: Output folder for the generated TypeScript file

// This script scans a Go package for struct definitions, generates a temporary Go program to convert these structs into TypeScript interfaces,
// and then replaces shared interfaces in the generated TypeScript files with imports from a common.ts file.

// Example usage (execute in backend folder): go run main.go -in /pkg/types/api -out ../frontend/types/api

func main() {
	// Defining flags for the API types package path and the output folder
	var inputPath, outputPath string
	flag.StringVar(&inputPath, "in", "", "Path to the Go package containing the API types")
	flag.StringVar(&outputPath, "out", ".", "Output folder for the generated TypeScript file")
	flag.Parse()

	// Ensure the package path is provided
	if inputPath == "" {
		fmt.Println("You must specify a package path using the -in flag.")
		return
	}

	if outputPath == "" {
		fmt.Println("You must specify a destination path using the -in flag.")
		return
	}

	// Step 1: Scan the input package for struct definitions
	structs, err := scanPackageForStructs(inputPath)
	if err != nil {
		panic(err)
	}

	// Step 2: Generate the Go program for typescriptify conversion
	programFileName := "temp.go"
	programSource := generateProgram(structs, outputPath)

	// Step 3: Write the program to a .go file
	if err := os.WriteFile(programFileName, []byte(programSource), 0644); err != nil {
		panic(err)
	}

	// Step 4: Execute the generated Go program
	cmd := exec.Command("go", "run", programFileName)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	fmt.Println("Output:", out.String())

	// Step 5: Delete the generated Go program file
	if err := os.Remove(programFileName); err != nil {
		panic(err)
	}

	// Step 6: Replace common interfaces with imports
	var commonInterfaces FileStructs
	for _, file := range structs {
		if file.FileName == "common" {
			commonInterfaces = file
		}
	}

	files, err := os.ReadDir(outputPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if strings.Contains(file.Name(), "common") || !strings.HasSuffix(file.Name(), ".ts") {
			continue
		}

		filePath := filepath.Join(outputPath, file.Name())
		err := replaceCommonWithImport(filePath, commonInterfaces.StructNames)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", filePath, err)
		}
	}
}

func replaceCommonWithImport(filePath string, commonInterfaces []string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	modifiedContent := string(content)
	var toImport []string

	for _, interfaceName := range commonInterfaces {
		regex, err := regexp.Compile(`(?ms)` + `^export interface ` + interfaceName + ` \{.*?\}$`)
		if err != nil {
			return err
		}
		if regex.MatchString(modifiedContent) {
			toImport = append(toImport, interfaceName)
			modifiedContent = regex.ReplaceAllString(modifiedContent, "")
		}
	}

	if len(toImport) > 0 {
		imports := "import { " + strings.Join(toImport, ", ") + " } from './common.ts';\n\n"
		modifiedContent = imports + modifiedContent
	}

	// Write the modified content back to the file
	return os.WriteFile(filePath, []byte(modifiedContent), 0644)
}

func scanPackageForStructs(inputPath string) ([]FileStructs, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, inputPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var filesStructs []FileStructs
	for _, pkg := range pkgs {
		for filePath, file := range pkg.Files {
			var structNames []string
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						structNames = append(structNames, typeSpec.Name.Name)
					}
				}
			}
			if len(structNames) > 0 {
				filename := filepath.Base(filePath)
				filesStructs = append(filesStructs, FileStructs{
					// Remove the .go extension
					FileName:    filename[:len(filename)-3],
					StructNames: structNames,
				})
			}
		}
	}
	return filesStructs, nil
}

func generateProgram(structs []FileStructs, outputPath string) string {
	var imports strings.Builder
	imports.WriteString(`import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
`)
	imports.WriteString(fmt.Sprintf("\tapitypes \"%s\"\n", "github.com/gobitfly/beaconchain/pkg/types/api"))
	imports.WriteString(")\n\n")

	var body strings.Builder
	body.WriteString(`func main() {
	var converter *typescriptify.TypeScriptify
	var err error` + "\n\n")

	for _, structs := range structs {
		body.WriteString(`
			converter = typescriptify.New()
			converter = converter.WithBackupDir("")
			converter.ManageType(time.Time{}, typescriptify.TypeOptions{TSType: "number"})
			converter.ManageType(decimal.Decimal{}, typescriptify.TypeOptions{TSType: "string"})
			converter.WithInterface(true)` + "\n")

		for _, structName := range structs.StructNames {
			body.WriteString(fmt.Sprintf("\tconverter = converter.Add(apitypes.%s{})\n", structName))
		}

		body.WriteString(fmt.Sprintf("\terr = converter.ConvertToFile(\"%s\")\n", outputPath+structs.FileName+".ts"))
		body.WriteString(`	if err != nil {
			panic(err.Error())
		}` + "\n\n")
	}
	body.WriteString("}\n")

	return fmt.Sprintf("package main\n\n%s%s", imports.String(), body.String())
}
