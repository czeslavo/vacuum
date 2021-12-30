package utils

import (
	"encoding/json"
	"fmt"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

const (
	// OpenApi3 is used by all OpenAPI 3+ docs
	OpenApi3 = "openapi"

	// OpenApi2 is used by all OpenAPI 2 docs, formerly known as swagger.
	OpenApi2 = "swagger"

	// AsyncApi is used by akk AsyncAPI docs, all versions.
	AsyncApi = "asyncapi"
)

// FindNodes will find a node based on JSONPath.
func FindNodes(yamlData []byte, jsonPath string) ([]*yaml.Node, error) {
	jsonPath = FixContext(jsonPath)

	var node yaml.Node
	yaml.Unmarshal(yamlData, &node)

	path, err := yamlpath.NewPath(jsonPath)
	if err != nil {
		return nil, err
	} else {
		results, err := path.Find(&node)
		if err != nil {
			return nil, err
		}
		return results, nil
	}
}

// ConvertInterfaceIntoStringMap will convert an unknown input into a string map.
func ConvertInterfaceIntoStringMap(context interface{}) map[string]string {
	converted := make(map[string]string)
	if context != nil {
		if v, ok := context.(map[string]interface{}); ok {
			for k, n := range v {
				if s, okB := n.(string); okB {
					converted[k] = s
				}
			}
		}
		if v, ok := context.(map[string]string); ok {
			for k, n := range v {
				converted[k] = n
			}
		}
		if v, ok := context.(map[string]string); ok {
			for k, n := range v {
				converted[k] = n
			}
		}
	}
	return converted
}

// ConvertInterfaceToStringArray will convert an unknown input map type into a string array/slice
func ConvertInterfaceToStringArray(raw interface{}) []string {
	if vals, ok := raw.(map[string]interface{}); ok {
		var s []string
		for _, v := range vals {
			if v, ok := v.([]interface{}); ok {
				for _, q := range v {
					s = append(s, fmt.Sprint(q))
				}
			}
		}
		return s
	}
	if vals, ok := raw.(map[string][]string); ok {
		var s []string
		for _, v := range vals {
			s = append(s, v...)
		}
		return s
	}
	return nil
}

// ConvertInterfaceArrayToStringArray will convert an unknown interface array type, into a string slice
func ConvertInterfaceArrayToStringArray(raw interface{}) []string {
	if vals, ok := raw.([]interface{}); ok {
		s := make([]string, len(vals))
		for i, v := range vals {
			s[i] = fmt.Sprint(v)
		}
		return s
	}
	if vals, ok := raw.([]string); ok {
		return vals
	}
	return nil
}

// ExtractValueFromInterfaceMap pulls out an unknown value from a map using a string key
func ExtractValueFromInterfaceMap(name string, raw interface{}) interface{} {

	if propMap, ok := raw.(map[string]interface{}); ok {
		if props, ok := propMap[name].([]interface{}); ok {
			return props
		}
	}
	if propMap, ok := raw.(map[string][]string); ok {
		return propMap[name]
	}
	return nil
}

// FindFirstKeyNode will locate the first key and value yaml.Node based on a key.
func FindFirstKeyNode(key string, nodes []*yaml.Node) (keyNode *yaml.Node, valueNode *yaml.Node) {

	for i, v := range nodes {
		if key != "" && key == v.Value {
			return v, nodes[i+1] // next node is what we need.
		}
		if len(v.Content) > 0 {
			x, y := FindFirstKeyNode(key, v.Content)
			if x != nil && y != nil {
				return x, y
			}
		}
	}
	return nil, nil
}

// FindKeyNode is a non-recursive search of an  yaml.Node Content for a child node with a key.
// Returns the key and value
func FindKeyNode(key string, nodes []*yaml.Node) (keyNode *yaml.Node, valueNode *yaml.Node) {

	for i, v := range nodes {
		if key == v.Value {
			return v, nodes[i+1] // next node is what we need.
		}
		for x, j := range v.Content {
			if key == j.Value {
				return v, v.Content[x+1] // next node is what we need.
			}
		}
	}
	return nil, nil
}

// IsNodeMap checks if the node is a map type
func IsNodeMap(node *yaml.Node) bool {
	return node.Tag == "!!map"
}

// IsNodeArray checks if a node is an array type
func IsNodeArray(node *yaml.Node) bool {
	return node.Tag == "!!seq"
}

// IsNodeStringValue checks if a node is a string value
func IsNodeStringValue(node *yaml.Node) bool {
	return node.Tag == "!!str"
}

// IsNodeIntValue will check if a node is an int value
func IsNodeIntValue(node *yaml.Node) bool {
	return node.Tag == "!!int"
}

// IsNodeFloatValue will check is a node is a float value.
func IsNodeFloatValue(node *yaml.Node) bool {
	return node.Tag == "!!float"
}

// IsNodeBoolValue will check is a node is a bool
func IsNodeBoolValue(node *yaml.Node) bool {
	return node.Tag == "!!bool"
}

// FixContext will clean up a JSONpath string to be correctly traversable.
func FixContext(context string) string {

	tokens := strings.Split(context, ".")
	var cleaned = []string{}
	for i, t := range tokens {

		if v, err := strconv.Atoi(t); err == nil {

			if v < 200 { // codes start here
				if cleaned[i-1] != "" {
					cleaned[i-1] += fmt.Sprintf("[%v]", t)
				}
			} else {
				cleaned = append(cleaned, t)
			}
			continue
		}
		cleaned = append(cleaned, strings.ReplaceAll(t, "(root)", "$"))

	}
	return strings.Join(cleaned, ".")
}

// IsJSON will tell you if a string is JSON or not.
func IsJSON(testString string) bool {
	if testString == "" {
		return false
	}
	runes := []rune(strings.TrimSpace(testString))
	if runes[0] == '{' && runes[len(runes)-1] == '}' {
		return true
	}
	return false
}

// IsYAML will tell you if a string is YAML or not.
func IsYAML(testString string) bool {
	if testString == "" {
		return false
	}
	if IsJSON(testString) {
		return false
	}
	var n interface{}
	err := yaml.Unmarshal([]byte(testString), &n)
	if err != nil {
		return false
	}
	_, err = yaml.Marshal(n)
	return err == nil
}

// ConvertYAMLtoJSON will do exactly what you think it will. It will deserialize YAML into serialized JSON.
func ConvertYAMLtoJSON(yamlData []byte) ([]byte, error) {
	var decodedYaml map[string]interface{}
	err := yaml.Unmarshal(yamlData, &decodedYaml)
	if err != nil {
		return nil, err
	}
	jsonData, err := json.Marshal(decodedYaml)
	if err != nil {
		return nil, err
	}
	return jsonData, nil

}

//func parseVersionTypeData(d interface{}) string {
//	switch d.(type) {
//	case int:
//		return strconv.Itoa(d.(int))
//	case float64:
//		return strconv.FormatFloat(d.(float64), 'f', 2, 32)
//	case bool:
//		if d.(bool) {
//			return "true"
//		}
//		return "false"
//	case []string:
//		return "multiple versions detected"
//	}
//	return fmt.Sprintf("%v", d)
//}
