package main

import (
	"fmt"
	"github.com/Jagerente/gocfg"
	"github.com/Jagerente/gocfg/pkg/docgens"
	"os"
	"workshop/internal/config"
)

const outputFile = ".env.dist"

func main() {
	cfg := new(config.Config)

	file, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("error creating %s file: %v", outputFile, err))
	}

	cfgManager := gocfg.NewDefault()
	if err := cfgManager.GenerateDocumentation(cfg, docgens.NewEnvDocGenerator(file)); err != nil {
		panic(err)
	}
}
