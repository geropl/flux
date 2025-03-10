package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/dave/jennifer/jen"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/token"
	"github.com/influxdata/flux/parser"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go source from Flux source",
	Long: `This utility creates Go sources files from Flux source files.
The process is to parse directories recursively and within each directory
write out a single file with the Flux AST representation of the directory source.
`,
	RunE: generate,
}

var (
	pkgName,
	rootDir,
	outDir,
	importFile,
	ignoreFile string
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVar(&pkgName, "go-pkg", "", "The fully qualified Go package name of the root package.")
	generateCmd.Flags().StringVar(&rootDir, "root-dir", ".", "The root level directory for all packages.")
	generateCmd.Flags().StringVar(&outDir, "out-dir", ".", "The directory to write compiled packages.")
	generateCmd.Flags().StringVar(&importFile, "import-file", "builtin_gen.go", "Location relative to root-dir to place a file to import all generated packages.")
	generateCmd.Flags().StringVar(&ignoreFile, "ignoreFile-file", ".fluxignore", "Location relative to root-dir of file containing packages to ignore one per line.")
}

func generate(cmd *cobra.Command, args []string) error {
	ignored, err := readIgnoreFile(ignoreFile)
	if err != nil {
		return err
	}
	var goPackages, testPackages []string
	err = walkDirs(rootDir, func(dir string) error {
		// Determine the absolute flux package path
		fluxPath, err := filepath.Rel(rootDir, dir)
		if err != nil {
			return err
		}
		if contains(fluxPath, ignored) {
			return nil
		}

		fset := new(token.FileSet)
		pkgs, err := parser.ParseDir(fset, dir)
		if err != nil {
			return err
		}

		// Annotate the packages with the absolute flux package path.
		for _, pkg := range pkgs {
			pkg.Path = fluxPath
		}

		var fluxPkg, testPkg *ast.Package
		switch len(pkgs) {
		case 0:
			return nil
		case 1:
			for k, v := range pkgs {
				if strings.HasSuffix(k, "_test") {
					testPkg = v
				} else {
					fluxPkg = v
				}
			}
		case 2:
			for k, v := range pkgs {
				if strings.HasSuffix(k, "_test") {
					testPkg = v
					continue
				}
				fluxPkg = v
			}
			if fluxPkg == nil {
				return fmt.Errorf("cannot have two Flux test packages in the same directory")
			}
			if testPkg == nil {
				return fmt.Errorf("cannot have two distinct non-test Flux packages in the same directory")
			}
		default:
			keys := make([]string, 0, len(pkgs))
			for k := range pkgs {
				keys = append(keys, k)
			}
			return fmt.Errorf("found more than 2 flux packages in directory %s; packages %v", dir, keys)
		}

		if fluxPkg != nil {
			if ast.Check(fluxPkg) > 0 {
				return errors.Wrapf(ast.GetError(fluxPkg), codes.Inherit, "failed to parse package %q", fluxPkg.Package)
			}
			// Assign import path
			fluxPkg.Path = fluxPath
			// Track go import path
			goPath := path.Join(pkgName, dir)
			if goPath != pkgName {
				goPackages = append(goPackages, goPath)
			}
			// Write the ast file
			if err := generateFluxASTFile(dir, fluxPkg); err != nil {
				return err
			}
			// Compile the flux package and put the result into the out directory.
			if err := compilePackage(outDir, fluxPath, fluxPkg); err != nil {
				return err
			}
		}
		if testPkg != nil {
			// Strip out test files with the testcase statement.
			validFiles := []*ast.File{}
			for _, file := range testPkg.Files {
				valid := true
				for _, item := range file.Body {
					if _, ok := item.(*ast.TestCaseStatement); ok {
						valid = false
					}
				}
				if valid {
					validFiles = append(validFiles, file)
				}
			}
			if len(validFiles) < len(testPkg.Files) {
				testPkg.Files = validFiles
			}

			if ast.Check(testPkg) > 0 {
				return errors.Wrapf(ast.GetError(testPkg), codes.Inherit, "failed to parse test package %q", testPkg.Package)
			}
			// Validate test package file use _test.flux suffix for the file name
			for _, f := range testPkg.Files {
				if !strings.HasSuffix(f.Name, "_test.flux") {
					return fmt.Errorf("flux test files must use the _test.flux suffix in their file name, found %q", path.Join(dir, f.Name))
				}
			}
			// Track go import path
			importPath := path.Join(pkgName, dir)
			if importPath != pkgName {
				testPackages = append(testPackages, importPath)
			}
			// Generate test AST file using non *_test package name since this is Go code that needs to be part of the normal build.
			if err := generateTestASTFile(dir, strings.TrimSuffix(testPkg.Package, "_test"), []*ast.Package{testPkg}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := generateTestPkgList(testPackages); err != nil {
		return err
	}

	// Write the import file
	f := jen.NewFile(path.Base(pkgName))
	f.HeaderComment("// DO NOT EDIT: This file is autogenerated via the builtin command.")
	f.Anon(goPackages...)
	return f.Save(filepath.Join(rootDir, importFile))
}

func generateFluxASTFile(dir string, pkg *ast.Package) error {
	file := jen.NewFile(pkg.Package)
	file.HeaderComment(`// DO NOT EDIT: This file is autogenerated via the builtin command.
//
// This file is empty but its existence ensures that each Flux package has a
// corresponding Go package. This simplifies importing Go packages with or without
// builtin values.
`)
	return file.Save(filepath.Join(dir, "flux_gen.go"))
}

func generateTestPkgList(imports []string) error {
	stmts := make([]jen.Code, len(imports)+2)
	// var pkgs []*ast.Package
	stmts[0] = jen.
		Var().
		Id("pkgs").
		Index().
		Op("*").
		Qual("github.com/influxdata/flux/ast", "Package")

	for i, path := range imports {
		// pkgs = append(pkgs, <imported_package>.FluxTestPackages...)
		stmts[i+1] = jen.Id("pkgs").Op("=").Id("append").Call(
			jen.Id("pkgs"), jen.Qual(path, "FluxTestPackages").Op("..."),
		)
	}

	// return pkgs
	stmts[len(stmts)-1] = jen.Return(jen.Id("pkgs"))

	file := jen.NewFile(path.Base(pkgName))
	file.HeaderComment("// DO NOT EDIT: This file is autogenerated via the builtin command.")
	// var FluxTestPackages = func() []*ast.Package {
	//     statements ...
	// }
	file.
		Var().
		Id("FluxTestPackages").
		Op("=").
		Func().
		Params().
		Index().
		Op("*").
		Qual("github.com/influxdata/flux/ast", "Package").
		Block(stmts...).
		Call()
	return file.Save(filepath.Join(rootDir, "test_packages.go"))
}

func generateTestASTFile(dir, pkg string, pkgs []*ast.Package) error {
	file := jen.NewFile(pkg)
	file.HeaderComment("// DO NOT EDIT: This file is autogenerated via the builtin command.")
	v, err := constructValue(reflect.ValueOf(pkgs))
	if err != nil {
		return err
	}
	file.Var().Id("FluxTestPackages").Op("=").Add(v)
	return file.Save(filepath.Join(dir, "flux_test_gen.go"))
}

func readIgnoreFile(fn string) ([]string, error) {
	f, err := os.Open(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var ignored []string
	for scanner.Scan() {
		ignored = append(ignored, strings.TrimSpace(scanner.Text()))
	}
	return ignored, scanner.Err()
}

func contains(s string, list []string) bool {
	for _, l := range list {
		if s == l {
			return true
		}
	}
	return false
}

func walkDirs(path string, f func(dir string) error) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	if err := f(path); err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			if err := walkDirs(filepath.Join(path, file.Name()), f); err != nil {
				return err
			}
		}
	}
	return nil
}

// indirectType returns a code statement that represents the type expression
// for the given type.
func indirectType(typ reflect.Type) *jen.Statement {
	switch typ.Kind() {
	case reflect.Map:
		c := jen.Index(indirectType(typ.Key()))
		c.Add(indirectType(typ.Elem()))
		return c
	case reflect.Ptr:
		c := jen.Op("*")
		c.Add(indirectType(typ.Elem()))
		return c
	case reflect.Array, reflect.Slice:
		c := jen.Index()
		c.Add(indirectType(typ.Elem()))
		return c
	default:
		return jen.Qual(typ.PkgPath(), typ.Name())
	}
}

// constructValue returns a Code value for the given value.
func constructValue(v reflect.Value) (jen.Code, error) {
	switch v.Kind() {
	case reflect.Array:
		s := indirectType(v.Type())
		values := make([]jen.Code, v.Len())
		for i := 0; i < v.Len(); i++ {
			val, err := constructValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
		s.Values(values...)
		return s, nil
	case reflect.Slice:
		if v.IsNil() {
			return jen.Nil(), nil
		}
		s := indirectType(v.Type())
		values := make([]jen.Code, v.Len())
		for i := 0; i < v.Len(); i++ {
			val, err := constructValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
		s.Values(values...)
		return s, nil
	case reflect.Interface:
		if v.IsNil() {
			return jen.Nil(), nil
		}
		return constructValue(v.Elem())
	case reflect.Ptr:
		if v.IsNil() {
			return jen.Nil(), nil
		}
		s := jen.Op("&")
		val, err := constructValue(reflect.Indirect(v))
		if err != nil {
			return nil, err
		}
		return s.Add(val), nil
	case reflect.Map:
		if v.IsNil() {
			return jen.Nil(), nil
		}
		s := indirectType(v.Type())
		keys := v.MapKeys()
		values := make(jen.Dict, v.Len())
		for _, k := range keys {
			key, err := constructValue(k)
			if err != nil {
				return nil, err
			}
			val, err := constructValue(v.MapIndex(k))
			if err != nil {
				return nil, err
			}
			values[key] = val
		}
		s.Values(values)
		return s, nil
	case reflect.Struct:
		switch v.Type().Name() {
		case "DateTimeLiteral":
			lit := v.Interface().(ast.DateTimeLiteral)
			fmtTime := lit.Value.Format(time.RFC3339Nano)
			return constructStructValue(v, map[string]*jen.Statement{
				"Value": jen.Qual("github.com/influxdata/flux/internal/parser", "MustParseTime").Call(jen.Lit(fmtTime)),
			})
		case "RegexpLiteral":
			lit := v.Interface().(ast.RegexpLiteral)
			regexString := lit.Value.String()
			return constructStructValue(v, map[string]*jen.Statement{
				"Value": jen.Qual("regexp", "MustCompile").Call(jen.Lit(regexString)),
			})
		}
		return constructStructValue(v, nil)
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String:
		typ := types[v.Kind()]
		cv := v.Convert(typ)
		return jen.Lit(cv.Interface()), nil
	default:
		return nil, fmt.Errorf("unsupport value kind %v", v.Kind())
	}
}

func constructStructValue(v reflect.Value, replace map[string]*jen.Statement) (*jen.Statement, error) {
	typ := v.Type()
	s := indirectType(typ)
	values := make(jen.Dict, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := typ.Field(i).Name
		if !field.CanInterface() {
			// Ignore private fields
			continue
		}
		if s, ok := replace[name]; ok {
			values[jen.Id(name)] = s
			continue
		}
		val, err := constructValue(field)
		if err != nil {
			return nil, err
		}
		values[jen.Id(name)] = val
	}
	return s.Values(values), nil
}

// types is map of reflect.Kind to reflect.Type for the primitive types
var types = map[reflect.Kind]reflect.Type{
	reflect.Bool:       reflect.TypeOf(false),
	reflect.Int:        reflect.TypeOf(int(0)),
	reflect.Int8:       reflect.TypeOf(int8(0)),
	reflect.Int16:      reflect.TypeOf(int16(0)),
	reflect.Int32:      reflect.TypeOf(int32(0)),
	reflect.Int64:      reflect.TypeOf(int64(0)),
	reflect.Uint:       reflect.TypeOf(uint(0)),
	reflect.Uint8:      reflect.TypeOf(uint8(0)),
	reflect.Uint16:     reflect.TypeOf(uint16(0)),
	reflect.Uint32:     reflect.TypeOf(uint32(0)),
	reflect.Uint64:     reflect.TypeOf(uint64(0)),
	reflect.Uintptr:    reflect.TypeOf(uintptr(0)),
	reflect.Float32:    reflect.TypeOf(float32(0)),
	reflect.Float64:    reflect.TypeOf(float64(0)),
	reflect.Complex64:  reflect.TypeOf(complex64(0)),
	reflect.Complex128: reflect.TypeOf(complex128(0)),
	reflect.String:     reflect.TypeOf(""),
}
