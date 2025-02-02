package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	systemMsg = "You are a bot parsing bank statements imported as a string in CSV format to extract column IDs for each specified column type. You will return the IDs (in 0-index array format) of each desired column. CSV contents:"
)

// var DocumentSchema = generateSchema[document.Document]()

type OpenAi struct {
	client *openai.Client
}

func NewOpenAi(config Config) (*OpenAi, error) {
	if config.ApiKey == "" {
		return nil, errors.New("failed to instantiate OpenAi: API key not provided")
	}

	c := openai.NewClient(option.WithAPIKey(config.ApiKey))
	return &OpenAi{
		client: c,
	}, nil
}

func createColumnIndexSchema(elements DesiredElements) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0, len(elements))

	for name, desc := range elements {
		properties[name] = map[string]interface{}{
			"type":        "number",
			"description": desc,
		}
		required = append(required, name)
	}

	return map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func (oai *OpenAi) FindElements(elements DesiredElements, content string) (map[string]int, error) {

	var toExtract []string

	for name := range elements {
		toExtract = append(toExtract, name)

	}
	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemMsg),
		openai.UserMessage("CSV Content:"),
		openai.UserMessage(content),
		openai.UserMessage("Elements to extract:"),
		openai.UserMessage(strings.Join(toExtract, ",")),
	}
	schema := createColumnIndexSchema(elements)
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("indeces"),
		Description: openai.F("Indeces of requested columns"),
		Schema:      openai.F(interface{}(schema)),
		Strict:      openai.Bool(true),
	}

	responseFormat := openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
		openai.ResponseFormatJSONSchemaParam{
			Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
			JSONSchema: openai.F(schemaParam),
		},
	)
	chat, err := oai.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages:       openai.F(msgs),
		ResponseFormat: responseFormat,
		Model:          openai.F(openai.ChatModelGPT4o2024_11_20),
	})
	if err != nil {
		var apierr *openai.Error
		if errors.As(err, &apierr) {
			fmt.Println(string(apierr.DumpRequest(true)))
			fmt.Println(string(apierr.DumpResponse(true)))
		}
		return nil, err
	}

	completionContent := chat.Choices[0].Message.Content
	var indices map[string]int
	if err := json.Unmarshal([]byte(completionContent), &indices); err != nil {
		return nil, fmt.Errorf("failed to parse column indices: %w", err)
	}
	return indices, nil
}
