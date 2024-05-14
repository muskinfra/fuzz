package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/bxcodec/faker/v3"
	"github.com/tidwall/gjson"
)


func printJSON(data gjson.Result) {
	fmt.Println(data.String())
}


func GenerateFakeDataJSON(schemaJSON string) (string, error) {
	// Parse the JSON schema
	var schema map[string]string
	err := json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		return "", err
	}

	// Modify schema to map "integer" to "int"
	for key, value := range schema {
		if value == "integer" {
			schema[key] = "int"
		}
	}
	fmt.Println("print",schema)
	// Generate fake data based on the schema
	fakeData := make(map[string]interface{})
	fmt.Println("value",fakeData)
	for key, value := range schema {
		switch value {
		case "string":
			fakeData[key] = faker.Word()
		case "int":
			intValue1, err := faker.RandomInt(100)
			intValue := intValue1[0]
			fmt.Print("valueee",intValue)
			if err != nil {
				return "", err
			}
			fakeData[key] = intValue
			
		// Add cases for other types as needed
		default:
			return "", fmt.Errorf("unsupported data type: %s", value)
		}
	}

	// Convert fakeData to JSON
	fmt.Print("key",fakeData)
	fakeDataJSON, err := json.Marshal(fakeData)
	if err != nil {
		return "", err
	}

	return string(fakeDataJSON), nil
}



func triggerAPI(apiURL, methodStr, payload string) (int, error) {
	// Create a new HTTP request
	req, err := http.NewRequest(methodStr, apiURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Parse the JSON response into a map
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return 0, err
	}
	fmt.Println("Response Body:", string(body))
	// Extract the id field from the response
	id, ok := responseData["id"].(float64)
	if !ok {
		return 0, errors.New("unable to extract id from response body")
	}

	// Convert the id to an integer and return
	return int(id), nil
}





func parseDefinitionObject(ref string) (string, string) {
	// Split the string by "/"
	parts := strings.Split(ref, "/")

	// Extract the last part
	definition := parts[len(parts)-2]
	objectName := parts[len(parts)-1]

	return definition, objectName
}
var payload string
var id int
func main() {
	jsonData, err := ioutil.ReadFile("swagger.json")
	if err != nil {
		fmt.Println("Error reading Swagger JSON file:", err)
		return
	}
	swaggerJSON := string(jsonData)
	definitions := gjson.Get(swaggerJSON, "definitions")
	definitions.ForEach(func(defName, defDetails gjson.Result) bool {
		// defNameStr := defName.String()
		// fmt.Printf("Name: %s\n", defNameStr)
	
		// Extract properties
		properties := defDetails.Get("properties")
		propMap := make(map[string]interface{})
		properties.ForEach(func(propName, propDetails gjson.Result) bool {
			propNameStr := propName.String()
			propType := propDetails.Get("type").String()
			propMap[propNameStr] = propType
			return true
		})
	
		// Convert propMap to JSON
		propJSON, err := json.MarshalIndent(propMap, "", "    ")
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return true
		}
	    payload =string(propJSON)
		return true
	})
	alreadyPrinted := make(map[string]bool)

	paths := gjson.Get(swaggerJSON, "paths")
	paths.ForEach(func(path, methods gjson.Result) bool {
		pathStr := path.String()
		methods.ForEach(func(method, details gjson.Result) bool {
			methodStr := method.String()

			parameters := details.Get("parameters")
			parameters.ForEach(func(_, param gjson.Result) bool {
				paramName := param.Get("name").String()
				paramSchemaRef := param.Get("schema.$ref").String()
				key := fmt.Sprintf("%s-%s-%s-%s", pathStr, methodStr, paramName, paramSchemaRef)
				if !alreadyPrinted[key] {
					fmt.Printf("Path: %s, Method: %s, Parameter Name: %s, Parameter Schema Ref: %s\n", pathStr, methodStr, paramName, paramSchemaRef)
					alreadyPrinted[key] = true
				}
				// definition, objectName := parseDefinitionObject(paramSchemaRef)
				// fmt.Println(definition,objectName)
				methodStr := strings.ToUpper(method.String())
				
				if paramName == "body" { // Check if parameter is body
					apiURL := "http://localhost:4000" + pathStr
					
					if err != nil {
						fmt.Println("Error generating payload:", err)
						return true
					}
					fmt.Println(payload, apiURL, apiURL)
					fakePayload, err := GenerateFakeDataJSON(payload)
					if err != nil {
						fmt.Println("Error generating payload:", err)
					}
					payload = fakePayload
					fmt.Println("payload", payload)
					id , err = triggerAPI(apiURL, methodStr, payload)
					fmt.Println("ID:", id)
					if err != nil {
						fmt.Println("Error triggering API:", err)
					} 
					apiURL = "http://localhost:4000/coverage" 
					// After successful POST request, store the user ID
					// Trigger the GET request
					_, err = triggerAPI(apiURL,"GET", "")
					if err != nil {
						fmt.Println("Error triggering GET API:", err)
					}
				}
					fmt.Println("ID:", id)
					apiURL := strings.ReplaceAll(pathStr, "{id}", strconv.Itoa(id))

					
					apiURL = "http://localhost:4000" + apiURL
					// After successful POST request, store the user ID
					// Trigger the GET request
					_, err = triggerAPI(apiURL,methodStr, "")
					if err != nil {
						fmt.Println("Error triggering GET API:", err)
					}
					apiURL = "http://localhost:4000/coverage" 
					// After successful POST request, store the user ID
					// Trigger the GET request
					_, err = triggerAPI(apiURL,"GET", "")
					if err != nil {
						fmt.Println("Error triggering GET API:", err)
					}     
				
				
				return true
			})

			responses := details.Get("responses")
			responses.ForEach(func(status, response gjson.Result) bool {
				statusStr := status.String()
				responseSchemaRef := response.Get("schema.$ref").String()
				key := fmt.Sprintf("%s-%s-%s", pathStr, methodStr, statusStr)
				if !alreadyPrinted[key] {
					fmt.Printf("Path: %s, Method: %s, Response Status: %s, Response Schema Ref: %s\n", pathStr, methodStr, statusStr, responseSchemaRef)
					alreadyPrinted[key] = true
				}

	

				return true
			})

			return true
		})
		return true
	})
	apiURL := "http://localhost:4000/exit" 
	// After successful POST request, store the user ID
	// Trigger the GET request
	_, err = triggerAPI(apiURL,"GET", "")
	if err != nil {
		fmt.Println("Error triggering GET API:", err)
	}     

    
}
