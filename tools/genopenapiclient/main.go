package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/deepmap/oapi-codegen/v2/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
)

var (
	specFile    string
	packageName string
	output      string
)

func main() {
	flag.StringVar(&specFile, "specfile", "", "openapi specification file")
	flag.StringVar(&packageName, "packagename", "", "package name")
	flag.StringVar(&output, "output", "", "output of the file name")
	flag.Parse()

	if specFile == "" {
		log.Fatalln("specfile was not provided")
	}

	if packageName == "" {
		log.Fatalln("packageName was not provided")
	}

	if output == "" {
		log.Fatalln("output was not provided")
	}

	genOpts := codegen.GenerateOptions{
		Client: true,
		Models: true,
	}

	cfg := codegen.Configuration{
		PackageName:   packageName,
		Generate:      genOpts,
		OutputOptions: codegen.OutputOptions{},
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadFromFile(specFile)
	if err != nil {
		log.Fatalf("error loading swagger spec: %v", err)
	}

	codegen.TemplateFunctions["genResponseTypeName"] = genResponseTypeName
	codegen.TemplateFunctions["genResponsePayload"] = genResponsePayload

	code, err := codegen.Generate(swagger, cfg)
	if err != nil {
		log.Fatalf("error generating code: %v", err)
	}

	outputDir := filepath.Dir(output)
	err = os.MkdirAll(outputDir, 0750)
	if err != nil && os.IsNotExist(err) {
		log.Fatalf("failed to create output dir: %s", outputDir)
	}

	err = os.WriteFile(output, []byte(code), 0600)
	if err != nil {
		log.Fatalf("error writing generated code to file: %v", err)
	}
}

func genResponseTypeName(operationID string) string {
	return fmt.Sprintf("%s%s", codegen.UppercaseFirstCharacter(operationID), "_Response")
}

func genResponsePayload(operationID string) string {
	var buffer = bytes.NewBufferString("")

	// Here is where we build up a response:
	fmt.Fprintf(buffer, "&%s{\n", genResponseTypeName(operationID))
	fmt.Fprintf(buffer, "Body: bodyBytes,\n")
	fmt.Fprintf(buffer, "HTTPResponse: rsp,\n")
	fmt.Fprintf(buffer, "}")

	return buffer.String()
}
