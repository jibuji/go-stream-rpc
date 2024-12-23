package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jibuji/go-stream-rpc/internal/generator"

	"google.golang.org/protobuf/compiler/protogen"
)

// getServiceDir returns the path to the service directory that will contain
// the generated service implementation skeleton
func getServiceDir(baseDir string) string {
	// Get the directory containing the proto file and add "service" sibling directory
	dir := strings.TrimSuffix(baseDir, filepath.Base(baseDir))
	return dir + "service"
}

// main is the entry point for the protoc-gen-stream-rpc plugin
// It generates:
// 1. Client code (_client.pb.go)
// 2. Server interfaces (_server.pb.go)
// 3. Service implementation skeletons in the service/ directory
func main() {
	var (
		flags flag.FlagSet
		_     = flags.String("go_out", "", "")
		_     = flags.String("go_opt", "", "")
	)

	protogen.Options{ParamFunc: flags.Set}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			// For each service in the file
			for _, service := range f.Services {
				// Generate client code
				clientFileName := strings.TrimSuffix(f.GeneratedFilenamePrefix, ".pb") + "_" + strings.ToLower(service.GoName) + "_client.pb.go"
				clientFile := gen.NewGeneratedFile(clientFileName, f.GoImportPath)

				methods := make([]generator.Method, 0)
				for _, method := range service.Methods {
					methods = append(methods, generator.Method{
						Name:       method.GoName,
						InputType:  method.Input.GoIdent.GoName,
						OutputType: method.Output.GoIdent.GoName,
					})
				}

				data := generator.TemplateData{
					PackageName:  string(f.GoPackageName),
					ProtoPackage: string(f.GoImportPath),
					ServiceName:  service.GoName,
					Methods:      methods,
				}

				// Generate client code
				if err := generator.GenerateClient(clientFile, data); err != nil {
					return err
				}

				// Generate server code
				serverFileName := strings.TrimSuffix(f.GeneratedFilenamePrefix, ".pb") + "_" + strings.ToLower(service.GoName) + "_server.pb.go"
				serverFile := gen.NewGeneratedFile(serverFileName, f.GoImportPath)
				if err := generator.GenerateServer(serverFile, data); err != nil {
					return err
				}

				// Generate service skeleton
				serviceDir := getServiceDir(f.GeneratedFilenamePrefix)
				skeletonFileName := serviceDir + "/" + strings.ToLower(service.GoName) + ".go"

				// Check if file already exists
				if _, err := os.Stat(skeletonFileName); err == nil {
					// File exists, parse it and only add missing methods
					existingMethods, err := generator.ParseExistingService(skeletonFileName)
					if err != nil {
						log.Printf("Warning: Could not parse existing service file: %v", err)
						continue // Skip generation if we can't parse existing file
					}

					// Read the entire file content
					content, err := os.ReadFile(skeletonFileName)
					if err != nil {
						return fmt.Errorf("failed to read existing file: %v", err)
					}

					// Generate new content with appended methods
					newContent, err := generator.AppendMissingMethods(string(content), data, existingMethods)
					if err != nil {
						return fmt.Errorf("failed to append methods: %v", err)
					}

					// Write back to the file if there were changes
					if newContent != string(content) {
						if err := os.WriteFile(skeletonFileName, []byte(newContent), 0644); err != nil {
							return fmt.Errorf("failed to write updated file: %v", err)
						}
					}
				} else if os.IsNotExist(err) {
					// File doesn't exist, generate new skeleton
					skeletonFile := gen.NewGeneratedFile(skeletonFileName, f.GoImportPath)
					if err := generator.GenerateSkeleton(skeletonFile, data, nil); err != nil {
						return fmt.Errorf("failed to generate skeleton: %v", err)
					}
				}
			}
		}
		return nil
	})
}
