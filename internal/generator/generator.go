package generator

import (
	"bytes"
	"io"
)

type Method struct {
	Name       string
	InputType  string
	OutputType string
}

type TemplateData struct {
	PackageName  string
	ProtoPackage string
	ServiceName  string
	Methods      []Method
}

func (m Method) Signature() string {
	return "(" + "ctx context.Context, req *" + m.InputType + ") *" + m.OutputType
}

func (m Method) CallSignature() string {
	return "(ctx, req)"
}

func GenerateClient(w io.Writer, data TemplateData) error {
	var buf bytes.Buffer
	if err := clientStubTemplate.Execute(&buf, data); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func GenerateServer(w io.Writer, data TemplateData) error {
	var buf bytes.Buffer
	if err := serverStubTemplate.Execute(&buf, data); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// AppendMissingMethods appends only the methods that don't exist in the service file
func AppendMissingMethods(existingContent string, data TemplateData, existingMethods map[string]*ExistingMethod) (string, error) {
	var buf bytes.Buffer

	// Write existing content without the last '}'
	content := existingContent //strings.TrimRight(existingContent, "}\n ")
	buf.WriteString(content)

	// Add only missing methods
	for _, method := range data.Methods {
		if _, exists := existingMethods[method.Name]; !exists {
			buf.WriteString("\n\n")
			methodData := struct {
				ServiceName string
				Method
			}{
				ServiceName: data.ServiceName,
				Method:      method,
			}
			if err := methodTemplate.Execute(&buf, methodData); err != nil {
				return "", err
			}
		}
	}

	// Close the struct
	buf.WriteString("\n")
	return buf.String(), nil
}
