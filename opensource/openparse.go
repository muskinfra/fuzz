package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type EndpointInfo struct {
	Path         string                 `json:"path"`
	Method       string                 `json:"method"`
	RequestBody  map[string]interface{} `json:"requestBody,omitempty"`
	ResponseBody map[string]interface{} `json:"responseBody,omitempty"`
}

var definitions map[string]interface{}
var authToken = ""

func convertMapInterfaceToString(input map[interface{}]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for key, value := range input {
		strKey := fmt.Sprintf("%v", key)
		switch value.(type) {
		case map[interface{}]interface{}:
			output[strKey] = convertMapInterfaceToString(value.(map[interface{}]interface{}))
		case []interface{}:
			output[strKey] = convertSliceInterfaceToString(value.([]interface{}))
		default:
			output[strKey] = value
		}
	}
	return output
}

func convertSliceInterfaceToString(input []interface{}) []interface{} {
	output := make([]interface{}, len(input))
	for i, value := range input {
		switch value.(type) {
		case map[interface{}]interface{}:
			output[i] = convertMapInterfaceToString(value.(map[interface{}]interface{}))
		case []interface{}:
			output[i] = convertSliceInterfaceToString(value.([]interface{}))
		default:
			output[i] = value
		}
	}
	return output
}

func ParseAPIDefinition(yamlData []byte) ([]EndpointInfo, error) {
	var apiSpec map[interface{}]interface{}
	err := yaml.Unmarshal(yamlData, &apiSpec)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML data: %w", err)
	}

	apiSpecConverted := convertMapInterfaceToString(apiSpec)
	// fmt.Printf("Parsed YAML data: %+v\n", apiSpecConverted)

	paths, ok := apiSpecConverted["paths"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("paths section not found in API definition")
	}

	definitions, ok = apiSpecConverted["definitions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("definitions section not found in API definition")
	}

	var endpoints []EndpointInfo
	for path, pathData := range paths {
		pathDetails := pathData.(map[string]interface{})
		for method, methodData := range pathDetails {
			methodInfo := methodData.(map[string]interface{})
			info := EndpointInfo{
				Path:   path,
				Method: method,
			}

			if parameters, ok := methodInfo["parameters"].([]interface{}); ok {
				for _, param := range parameters {
					paramMap := param.(map[string]interface{})
					if paramMap["in"] == "body" {
						info.RequestBody = resolveRefSchema(paramMap["schema"])
					}
				}
			}

			if responses, ok := methodInfo["responses"].(map[string]interface{}); ok {
				for _, response := range responses {
					responseMap := response.(map[string]interface{})
					if schema, ok := responseMap["schema"]; ok {
						info.ResponseBody = resolveRefSchema(schema)
						break
					}
				}
			}

			endpoints = append(endpoints, info)
		}
	}

	return endpoints, nil
}

func resolveRefSchema(data interface{}) map[string]interface{} {
	dataMap := data.(map[string]interface{})
	if ref, ok := dataMap["$ref"].(string); ok {
		refParts := strings.Split(ref, "/")
		definitionName := refParts[len(refParts)-1]
		if definition, ok := definitions[definitionName]; ok {
			return resolveRefSchema(definition)
		}
	}
	if properties, ok := dataMap["properties"].(map[string]interface{}); ok {
		for key, prop := range properties {
			properties[key] = resolveRefSchema(prop)
		}
	}
	return dataMap
}

func generateRandomData(schema map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for key, prop := range properties {
			propMap := prop.(map[string]interface{})
			data[key] = generateRandomValue(propMap)
		}
	}
	return data
}

func generateRandomValue(schema map[string]interface{}) interface{} {
	switch schema["type"] {
	case "string":
		return randomString()
	case "integer":
		return rand.Intn(10)
	case "boolean":
		return rand.Intn(2) == 1
	case "object":
		return generateRandomData(schema)
	case "array":
		itemSchema := schema["items"].(map[string]interface{})
		return []interface{}{generateRandomValue(itemSchema)}
	default:
		return nil
	}
}

func randomString() string {
	return uuid.New().String()
}

func triggerAPI(endpoint EndpointInfo) int {
	url := "http://localhost:4000" + endpoint.Path
	jsonData := generateRandomData(endpoint.RequestBody)
	requestBody, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return 0
	}

	req, err := http.NewRequest(strings.ToUpper(endpoint.Method), url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("User-Agent", "Go-Client")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error triggering API:", err)
		return 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return 0
	}

	fmt.Printf("Response for %s %s:\n%s\n", endpoint.Method, endpoint.Path, body)

	var responseMap map[string]interface{}
	json.Unmarshal(body, &responseMap)
	if id, ok := responseMap["id"].(float64); ok {
		return int(id)
	}

	return 0
}

func printCoverage(authToken string) {
	coverageURL := "http://localhost:4000/coverage"

	req, err := http.NewRequest("GET", coverageURL, nil)
	if err != nil {
		fmt.Println("Error creating coverage request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error getting coverage:", err)
		return
	}
	defer resp.Body.Close()

	coverageBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading coverage response body:", err)
		return
	}

	fmt.Println("Coverage Response:")
	fmt.Println(string(coverageBody))
}

func main() {
	yamlData, err := ioutil.ReadFile("swagger.yaml")
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return
	}

	endpointInfos, err := ParseAPIDefinition(yamlData)
	if err != nil {
		fmt.Println("Error parsing API definition:", err)
		return
	}

	var postEndpoint, putEndpoint, getEndpoint, deleteEndpoint EndpointInfo
	for _, info := range endpointInfos {
		if strings.ToUpper(info.Method) == "POST" {
			postEndpoint = info
		} else if strings.ToUpper(info.Method) == "PUT" {
			putEndpoint = info
		} else if strings.ToUpper(info.Method) == "GET" {
			getEndpoint = info
		} else if strings.ToUpper(info.Method) == "DELETE" {
			deleteEndpoint = info
		}
	}

	printCoverage(authToken)

	// Trigger POST to create resource and get the ID
	postID := triggerAPI(postEndpoint)
	if postID == 0 {
		fmt.Println("Failed to create resource, ID not found in response.")
		return
	}

	// Trigger PUT to update resource
	putEndpoint.Path = strings.ReplaceAll(putEndpoint.Path, "{id}", fmt.Sprintf("%d", postID))
	triggerAPI(putEndpoint)

	// Trigger GET to retrieve resource
	getEndpoint.Path = strings.ReplaceAll(getEndpoint.Path, "{id}", fmt.Sprintf("%d", postID))
	triggerAPI(getEndpoint)

	// Trigger DELETE to remove resource
	deleteEndpoint.Path = strings.ReplaceAll(deleteEndpoint.Path, "{id}", fmt.Sprintf("%d", postID))
	triggerAPI(deleteEndpoint)

	printCoverage(authToken)
}


func printSchema(schema map[string]interface{}) {
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(schemaJSON))
}
