package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/spf13/cobra"
)

var (
	urlFlag string
	outFlag string
)

var rootCmd = &cobra.Command{
	Use:   "schemagen",
	Short: "Generate JSON Schema from API response",
	Long:  `Generate JSON Schema from API response for use in validating API responses.`,
	Run: func(cmd *cobra.Command, args []string) {
		if urlFlag == "" {
			fmt.Println("URL is required")
			os.Exit(1)
		}

		response, err := httpRequest(urlFlag)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		jsonParsed, err := gabs.ParseJSON([]byte(response))
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			os.Exit(1)
		}

		schema := generateSchema(jsonParsed)

		// Print the generated schema
		fmt.Println(schema)

		if outFlag != "" {
			// Write schema to the output file
			err := writeToFile(outFlag, schema)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				os.Exit(1)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&urlFlag, "url", "u", "", "URL to the API endpoint")
	rootCmd.MarkFlagRequired("url") // Mark URL flag as required
	rootCmd.Flags().StringVarP(&outFlag, "out", "o", "", "Output file for the generated JSON Schema")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func httpRequest(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("HTTP request failed with status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func generateSchema(data *gabs.Container) string {
	schema := make(map[string]interface{})
	schema["$schema"] = "http://json-schema.org/draft-07/schema#"
	schema["title"] = "Generated schema for Root"

	if data.Data() != nil && isJSONArray(data.Data()) {
		schema["type"] = "array"
		children, _ := data.Children()
		if len(children) > 0 {
			items := generateType(children[0].Data())
			schema["items"] = items
		}
	} else if data.Data() != nil && isJSONMap(data.Data()) {
		schema["type"] = "object"
		objectSchema := generateType(data.Data()).(map[string]interface{})
		for key, value := range objectSchema {
			schema[key] = value
		}
	}

	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	return string(schemaJSON)
}

func generateType(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		properties := make(map[string]interface{})
		for key, value := range v {
			properties[key] = generateType(value)
		}
		objectSchema := map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}
		objectSchema["required"] = getRequiredFields(properties)
		return objectSchema
	case []interface{}:
		if len(v) > 0 {
			return map[string]interface{}{
				"type":  "array",
				"items": generateType(v[0]),
			}
		}
		return map[string]interface{}{"type": "array"}
	default:
		return map[string]interface{}{"type": detectType(v)}
	}
}

func getRequiredFields(properties map[string]interface{}) []string {
	required := make([]string, 0)

	for key := range properties {
		required = append(required, key)
	}

	return required
}

func detectType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64, uint, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case bool:
		return "boolean"
	default:
		return "string" // Fallback to string for unknown types
	}
}

func writeToFile(filename, content string) error {
	file, err := os.Create(fmt.Sprintf("%s.json", filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func isJSONMap(data interface{}) bool {
	_, isMap := data.(map[string]interface{})
	return isMap
}

func isJSONArray(data interface{}) bool {
	_, isArray := data.([]interface{})
	return isArray
}
