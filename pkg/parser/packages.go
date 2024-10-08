package parser

import (
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// GetPackages returns a list of all package directories in the project
//
// Example:
//
//	packages, err := parser.GetPackages()
//	if err != nil {
//		log.Fatalf("Error fetching packages: %v", err)
//	}
//	for _, pkg := range packages {
//		fmt.Printf("Package: %s\n", pkg)
//	}
func GetPackages() ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles, // We only need the file paths
	}

	rootDir, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, err
	}

	var packageDirs []string
	for _, pkg := range pkgs {
		if len(pkg.GoFiles) > 0 {
			// To get the package directory, we take the directory of the first Go file
			dir := filepath.Dir(pkg.GoFiles[0])

			// if the directory is the root of the project, we skip it
			if dir == rootDir {
				continue
			}
			packageDirs = append(packageDirs, dir)
		}
	}

	return packageDirs, nil
}
