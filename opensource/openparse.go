package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type EndpointInfo struct {
	Path         string                 `json:"path"`
	Method       string                 `json:"method"`
	Tags         []string               `json:"tags"`
	RequestBody  map[string]interface{} `json:"requestBody,omitempty"`
	ResponseBody map[string]interface{} `json:"responseBody,omitempty"`
}

var definitions map[string]interface{}
var authToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjoidHJ1ZSIsImxvZ2luX2lkIjoiMTExMTExMTExMSIsImxvZ2luX3R5cGUiOiJwaG9uZW5vIiwicmVxdWVzdF91c2VyX2lkIjoxLCJ0ZW5hbnRfaWQiOiIzNDQ0Y2JjNC0wMTA0LTU5YjUtYjU5MS00ZmUzOTY0NmNiNTEifQ.vGwblwb1yHqIweLB4M6lwyQd7rXI3lInFRx9mKqGIjo"

// Convert map[interface{}]interface{} to map[string]interface{}
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
			tags := []string{}
			if tagList, ok := methodInfo["tags"].([]interface{}); ok {
				for _, tag := range tagList {
					tags = append(tags, tag.(string))
				}
			}
			info := EndpointInfo{
				Path:   path,
				Method: method,
				Tags:   tags,
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

func generateRandomData(schema map[string]interface{}, gid string) map[string]interface{} {
	data := make(map[string]interface{})
	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		for key, prop := range properties {
			propMap := prop.(map[string]interface{})
			data[key] = generateRandomValue(key, propMap, gid)
		}
	}
	return data
}

func generateRandomValue(fieldName string, schema map[string]interface{}, gid string) interface{} {
	switch fieldName {
	case "operator_type":
		return getRandomOperatorType()
	case "policy_type":
		return getRandomPolicyType()
	case "attributes_list":
		return getRandomList()
	case "gid":
		return gid // Use the gid extracted from the POST response
	}

	switch schema["type"] {
	case "string":
		switch fieldName {
		case "gst":
			return generateRandomGST()
		case "pan":
			return generateRandomPAN()
		case "phone":
			return generateRandomPhone()
		}
		if schema["format"] == "uuid" {
			return uuid.New().String()
		}
		if schema["enum"] != nil {
			enumValues := schema["enum"].([]interface{})
			return enumValues[rand.Intn(len(enumValues))]
		}
		return randomString()
	case "integer":
		return rand.Intn(100)
	case "boolean":
		return rand.Intn(2) == 1
	case "object":
		return generateRandomData(schema, gid)
	case "array":
		itemSchema := schema["items"].(map[string]interface{})
		return []interface{}{generateRandomValue(fieldName, itemSchema, gid)}
	default:
		return nil
	}
}

func generateRandomGST() string {
	stateCode := rand.Intn(36) + 1
	if stateCode < 10 {
		return fmt.Sprintf("0%dAAAAA%04dA1Z%d", stateCode, rand.Intn(10000), rand.Intn(9))
	}
	return fmt.Sprintf("%dAAAAA%04dA1Z%d", stateCode, rand.Intn(10000), rand.Intn(9))
}

func generateRandomPAN() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digitset = "0123456789"
	return fmt.Sprintf("%s%s%s%s", string(charset[rand.Intn(len(charset))]),
		string(charset[rand.Intn(len(charset))]), string(charset[rand.Intn(len(charset))]),
		string(charset[rand.Intn(len(charset))]) + string(digitset[rand.Intn(len(digitset))]),
	)
}

func generateRandomPhone() string {
	return fmt.Sprintf("9%09d", rand.Intn(1000000000))
}

func getRandomOperatorType() string {
	operators := []string{"AND", "OR","",uuid.New().String()}
	return operators[rand.Intn(len(operators))]
}

func getRandomPolicyType() string {
	policyTypes := []string{"identity", "", uuid.New().String()}
	return policyTypes[rand.Intn(len(policyTypes))]
}

func randomString() string {
	return uuid.New().String()
}

func getRandomList() []string {
	rand.Seed(time.Now().UnixNano())

	// Predefined policy types
	policyTypes := []string{"gst", "pan", "phone", "email", "aadhaar", "employee_code", "gst_location"}

	// Generate a random UUID string
	randomUUID := uuid.New().String()

	// Randomly decide which list to return
	switch rand.Intn(3) {
	case 0:
		return policyTypes
	case 1:
		return []string{randomUUID}
	default:
		return []string{} // Empty list
	}
}

func triggerAPI(endpoint EndpointInfo, gid string) (map[string]interface{}, int) {
	url := "http://localhost:9034" + endpoint.Path
	jsonData := generateRandomData(endpoint.RequestBody, gid)
	requestBody, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return nil, 0
	}

	fmt.Printf("Request Body for %s %s:\n%s\n", endpoint.Method, endpoint.Path, requestBody)

	req, err := http.NewRequest(strings.ToUpper(endpoint.Method), url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return nil, 0
	}

	req.Header.Set("Authorization",  authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return nil, 0
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, resp.StatusCode
	}

	fmt.Printf("Response for %s %s:\n%s\n", endpoint.Method, endpoint.Path, body)
	fmt.Printf("%s %s Status Code: %d\n", endpoint.Method, endpoint.Path, resp.StatusCode)
	fmt.Printf("%s %s Duration: %v\n", endpoint.Method, endpoint.Path, duration)

	if strings.ToUpper(endpoint.Method) == "POST" && strings.Contains(endpoint.Path, "create-identity") {
		var responseBody map[string]interface{}
		if err := json.Unmarshal(body, &responseBody); err != nil {
			fmt.Println("Error unmarshalling response body:", err)
			return nil, resp.StatusCode
		}
		if data, ok := responseBody["data"].(map[string]interface{}); ok {
			if extractedGid, ok := data["gid"].(string); ok {
				fmt.Printf("Extracted gid from first POST: %s\n", extractedGid)
				return responseBody, resp.StatusCode
			}
		}
	}

	return nil, resp.StatusCode
}


func triggerCoverage() {
	url := "http://localhost:9034/coverage"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating HTTP request for /coverage:", err)
		return
	}

	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request for /coverage:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body for /coverage:", err)
		return
	}

	fmt.Printf("Response for GET /coverage:\n%s\n", body)
}

func main() {
	yamlData, err := ioutil.ReadFile("swagger.yaml")
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return
	}

	endpoints, err := ParseAPIDefinition(yamlData)
	if err != nil {
		fmt.Println("Error parsing API definition:", err)
		return
	}

	var firstPostGid string

	for {
		for _, endpoint := range endpoints {
			triggerCoverage()
			responseBody, statusCode := triggerAPI(endpoint, firstPostGid)
			triggerCoverage()

			if statusCode == http.StatusOK && firstPostGid == "" && strings.ToUpper(endpoint.Method) == "POST" && strings.Contains(endpoint.Path, "create-identity") {
				if data, ok := responseBody["data"].(map[string]interface{}); ok {
					if gid, ok := data["gid"].(string); ok {
						firstPostGid = gid
					}
				}
			}

			// Delay before the next request
			time.Sleep(1 * time.Second)
		}
	}
}
