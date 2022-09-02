package bundler

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"text/template"
)

// Bundle bundles the provided bundler.Settings into a runnable application. It is not guaranteed to be compilable
// without errors.
func Bundle(set Settings) error {
	mainFile, err := os.Create(path.Join(set.Path, "main.go"))
	if err != nil {
		return fmt.Errorf("error creating main.go: %w", err)
	}
	defer mainFile.Close()
	modFile, err := os.Create(path.Join(set.Path, "go.mod"))
	if err != nil {
		return fmt.Errorf("error creating go.mod: %w", err)
	}
	defer modFile.Close()

	err = mainTemplate.Execute(mainFile, set)
	if err != nil {
		return fmt.Errorf("error writing main.go: %w", err)
	}
	err = modTemplate.Execute(modFile, set)
	if err != nil {
		return fmt.Errorf("error writing go.mod: %w", err)
	}
	return nil
}

var (
	//go:embed main.templ
	mainTemplateString string
	mainTemplate       *template.Template

	//go:embed mod.templ
	modTemplateString string
	modTemplate       *template.Template
)

func init() {
	var err error
	mainTemplate, err = template.New("main.go").Parse(mainTemplateString)
	if err != nil {
		panic(err)
	}
	modTemplate, err = template.New("go.mod").Parse(modTemplateString)
	if err != nil {
		panic(err)
	}
}
