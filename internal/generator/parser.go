package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type ExistingMethod struct {
	Name         string
	Parameters   []string
	Results      []string
	Body         string
	ContextType  string
	RequestType  string
	ResponseType string
}

// SignatureMatches checks if the method signature matches the proto definition
func (m *ExistingMethod) SignatureMatches(protoMethod Method) bool {
	return m.ContextType == "Context" &&
		strings.HasSuffix(m.RequestType, protoMethod.InputType) &&
		strings.HasSuffix(m.ResponseType, protoMethod.OutputType)
}

// ParseExistingService parses an existing service file and returns the methods
func ParseExistingService(filePath string) (map[string]*ExistingMethod, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, nil // Return nil if file cannot be parsed (likely doesn't exist yet)
	}

	methods := make(map[string]*ExistingMethod)

	// Find the service type and its methods
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasSuffix(typeSpec.Name.Name, "Service") {
				_, ok := typeSpec.Type.(*ast.StructType)
				if ok {
					return true
				}
			}
		}

		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil {
				method := &ExistingMethod{
					Name: funcDecl.Name.Name,
				}

				// Parse parameters
				if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) == 2 {
					// First parameter should be context.Context
					ctxParam := funcDecl.Type.Params.List[0]
					if sel, ok := ctxParam.Type.(*ast.SelectorExpr); ok {
						method.ContextType = sel.Sel.Name
					}

					// Second parameter should be the request type
					reqParam := funcDecl.Type.Params.List[1]
					if star, ok := reqParam.Type.(*ast.StarExpr); ok {
						if sel, ok := star.X.(*ast.SelectorExpr); ok {
							method.RequestType = sel.Sel.Name
						}
					}
				}

				// Parse result
				if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) == 1 {
					result := funcDecl.Type.Results.List[0]
					if star, ok := result.Type.(*ast.StarExpr); ok {
						if sel, ok := star.X.(*ast.SelectorExpr); ok {
							method.ResponseType = sel.Sel.Name
						}
					}
				}

				// Store method body including the function signature
				startPos := fset.Position(funcDecl.Pos()).Offset
				endPos := fset.Position(funcDecl.End()).Offset
				method.Body = string(content[startPos:endPos])
				methods[method.Name] = method
			}
		}
		return true
	})

	return methods, nil
}
