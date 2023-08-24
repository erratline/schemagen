package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/Jeffail/gabs"
	"github.com/stretchr/testify/assert"
)

func TestHttpRequest(t *testing.T) {
	t.Run("Test http request", func(t *testing.T) {
		response, err := httpRequest("https://jsonplaceholder.typicode.com/todos/1")
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
	})
}

func TestWriteToFile(t *testing.T) {
	t.Run("Test write to file", func(t *testing.T) {
		fileName := "test.txt"
		content := "test content"

		err := writeToFile(fileName, content)
		assert.NoError(t, err)

		file, err := os.Open(fileName)
		assert.NoError(t, err)

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, file)
		assert.NoError(t, err)

		assert.Equal(t, content, buf.String())

		err = os.Remove(fileName)
		assert.NoError(t, err)
	})
}

func TestGenerateSchema(t *testing.T) {
	t.Run("Test generate schema", func(t *testing.T) {
		response, err := httpRequest("https://jsonplaceholder.typicode.com/todos/1")
		assert.NoError(t, err)

		jsonParsed, err := gabs.ParseJSON([]byte(response))
		assert.NoError(t, err)

		schema := generateSchema(jsonParsed)
		assert.NotEmpty(t, schema)
	})
}

func TestRootCmd(t *testing.T) {
	t.Run("Test root command", func(t *testing.T) {
		rootCmd.SetArgs([]string{"--url", "https://jsonplaceholder.typicode.com/todos/1"})
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})
}
