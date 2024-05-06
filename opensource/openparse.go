package main

import (
	"fmt"
	"io/ioutil"

	"github.com/tidwall/gjson"
)


func printJSON(data gjson.Result) {
	fmt.Println(data.String())
}
func printProperties(def gjson.Result) {
	// Check if definition is an object
	if def.Get("type").String() == "object" {
		properties := def.Get("properties")
		properties.ForEach(func(key, value gjson.Result) bool {
			fmt.Printf("\t%s: %s\n", key.String(), value.Get("type").String())

			// Check if property has sub-properties
			if value.Get("type").String() == "object" {
				printProperties(value)
			}

			return true
		})
	}
}
func main() {
	// Read Swagger JSON file
	
	jsonData, err := ioutil.ReadFile("swagger.json")
	if err != nil {
		fmt.Println("Error reading Swagger JSON file:", err)
		return
	}
	swaggerJSON := string(jsonData)

// Map to keep track of already printed combinations
alreadyPrinted := make(map[string]bool)

// Extract paths
paths := gjson.Get(swaggerJSON, "paths")
paths.ForEach(func(path, methods gjson.Result) bool {
	pathStr := path.String()
	methods.ForEach(func(method, details gjson.Result) bool {
		methodStr := method.String()

		// Extract parameters
		parameters := details.Get("parameters")
		parameters.ForEach(func(_, param gjson.Result) bool {
			paramName := param.Get("name").String()
			paramSchemaRef := param.Get("schema.$ref").String()
			key := fmt.Sprintf("%s-%s-%s-%s", pathStr, methodStr, paramName, paramSchemaRef)
			if !alreadyPrinted[key] {
				fmt.Printf("Path: %s, Method: %s, Parameter Name: %s, Parameter Schema Ref: %s\n", pathStr, methodStr, paramName, paramSchemaRef)
				alreadyPrinted[key] = true
			}
			return true
		})

		// Extract response schema ref
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


	definitions := gjson.Get(swaggerJSON, "definitions")
	definitions.ForEach(func(name, def gjson.Result) bool {
		fmt.Println("Name:", name.String())
		// Print definition
		printJSON(def)
		return true // keep iterating
	})




	
	
}
