package validation

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"strings"
)

type Validator struct{
	path string
	logger logger
}

func NewValidator(path string) Validator {
	return Validator{
		path: path,
		logger:logger{},
	}
}

// Validate returns true as successful
func (v Validator) Validate() bool {
	moduleReader := reader{
		logger: v.logger,
	}

	hclFiles, err := moduleReader.read(v.path)
	if err != nil {
		v.logger.EmitError(err.Error())
		return false
	}

	isAllValid := true

	for _, hclFile := range hclFiles {
		body := hclFile.Body.(*hclsyntax.Body)
		for _, block := range body.Blocks {
			switch block.Type {
			case "data", "resource":
				isValid := v.validateAttribute(block.Body)
				if !isValid {
					isAllValid = false
				}
			default:
			}
		}
	}

	return isAllValid
}

func (v Validator) validateAttribute(body *hclsyntax.Body) bool {
	isAllValid := true
	for _, attribute := range body.Attributes {
		expression := attribute.Expr
		switch expression.(type) {
		// TODO other Expr types
		case *hclsyntax.ConditionalExpr:
		case *hclsyntax.FunctionCallExpr:
		default:
			for _, variable := range expression.Variables() {
				if variable.RootName() == "var" {
					variableName := variable[1].(hcl.TraverseAttr).Name
					propertyName := attribute.Name
					isValid := v.validate(variableName, propertyName, variable[0].SourceRange())
					if !isValid {
						isAllValid = false
					}
				}
			}
		}
	}

	for _, block := range body.Blocks {
		isValid := v.validateAttribute(block.Body)
		if !isValid {
			isAllValid = false
		}
	}

	return isAllValid
}

func (v Validator) validate(variableName string, propertyName string, sourceRange hcl.Range) bool {
	if strings.LastIndex(variableName, propertyName) == len(variableName) - len(propertyName) {
		return true
	} else {
		message := fmt.Sprintf("%s:Line:%d Column:%d; Property: %s; Variable: %s", sourceRange.Filename, sourceRange.Start.Line, sourceRange.Start.Column, propertyName, variableName)
		v.logger.EmitError(message)
		return false
	}
}