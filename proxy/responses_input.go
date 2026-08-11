package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseResponsesInput(raw json.RawMessage) ([]OpenAIMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return nil, nil
	}

	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, fmt.Errorf("invalid input string: %w", err)
		}
		if strings.TrimSpace(s) == "" {
			return nil, nil
		}
		return []OpenAIMessage{{Role: "user", Content: s}}, nil
	}

	if trimmed[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, fmt.Errorf("invalid input array: %w", err)
		}
		return convertResponsesInputItems(items)
	}

	if trimmed[0] == '{' {
		return convertResponsesInputItems([]json.RawMessage{raw})
	}

	return nil, fmt.Errorf("unsupported input shape")
}

func convertResponsesInputItems(items []json.RawMessage) ([]OpenAIMessage, error) {
	messages := make([]OpenAIMessage, 0, len(items))
	pendingUserParts := []interface{}{}

	flushPendingUser := func() {
		if len(pendingUserParts) == 0 {
			return
		}
		messages = append(messages, OpenAIMessage{
			Role:    "user",
			Content: pendingUserParts,
		})
		pendingUserParts = nil
	}

	for _, item := range items {
		var obj map[string]interface{}
		if err := json.Unmarshal(item, &obj); err != nil {
			continue
		}

		typ, _ := obj["type"].(string)
		role, _ := obj["role"].(string)

		switch {
		case typ == "message" || (typ == "" && role != ""):
			flushPendingUser()
			msg := buildMessageFromInputItem(obj, role)
			if msg != nil {
				messages = append(messages, *msg)
			}

		case typ == "additional_tools":
			// Tool declarations are extracted separately and are not messages.

		case typ == "agent_message":
			flushPendingUser()
			if text := extractAgentMessageText(obj); text != "" {
				messages = append(messages, OpenAIMessage{Role: "user", Content: text})
			}

		case typ == "function_call_output" || typ == "tool_result" || typ == "custom_tool_call_output":
			flushPendingUser()
			callID, _ := obj["call_id"].(string)
			if callID == "" {
				callID, _ = obj["tool_call_id"].(string)
			}
			out := stringifyArbitrary(obj["output"])
			if out == "" {
				out = stringifyArbitrary(obj["content"])
			}
			messages = append(messages, OpenAIMessage{
				Role:       "tool",
				Content:    out,
				ToolCallID: callID,
			})

		case typ == "function_call":
			flushPendingUser()
			tc := ToolCall{
				ID:   stringField(obj, "call_id", "id"),
				Type: "function",
			}
			tc.Function.Name, _ = obj["name"].(string)
			tc.Function.Arguments = stringifyArbitrary(obj["arguments"])
			messages = appendAssistantToolCall(messages, tc)

		case typ == "custom_tool_call":
			flushPendingUser()
			tc := ToolCall{ID: stringField(obj, "call_id", "id"), Type: "function"}
			tc.Function.Name, _ = obj["name"].(string)
			input := stringifyArbitrary(obj["input"])
			args, _ := json.Marshal(map[string]string{"input": input})
			tc.Function.Arguments = string(args)
			messages = appendAssistantToolCall(messages, tc)

		case typ == "input_text" || typ == "text":
			text, _ := obj["text"].(string)
			if text != "" {
				pendingUserParts = append(pendingUserParts, map[string]interface{}{
					"type": "input_text",
					"text": text,
				})
			}

		case typ == "input_image", typ == "image", typ == "image_url":
			pendingUserParts = append(pendingUserParts, map[string]interface{}(obj))

		case typ == "output_text":
			flushPendingUser()
			text, _ := obj["text"].(string)
			if text != "" {
				messages = append(messages, OpenAIMessage{Role: "assistant", Content: text})
			}

		default:
			if role != "" {
				flushPendingUser()
				msg := buildMessageFromInputItem(obj, role)
				if msg != nil {
					messages = append(messages, *msg)
				}
			}
		}
	}

	flushPendingUser()
	return messages, nil
}

func appendAssistantToolCall(messages []OpenAIMessage, tc ToolCall) []OpenAIMessage {
	if n := len(messages); n > 0 &&
		messages[n-1].Role == "assistant" &&
		len(messages[n-1].ToolCalls) > 0 &&
		strings.TrimSpace(extractOpenAIMessageText(messages[n-1].Content)) == "" {
		messages[n-1].ToolCalls = append(messages[n-1].ToolCalls, tc)
		return messages
	}
	return append(messages, OpenAIMessage{
		Role:      "assistant",
		Content:   "",
		ToolCalls: []ToolCall{tc},
	})
}

// extractResponsesTools finds Codex tool declarations embedded in Responses
// input items instead of the top-level tools field.
func extractResponsesTools(raw json.RawMessage) []OpenAITool {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed[0] != '[' {
		return nil
	}

	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}

	var tools []OpenAITool
	for _, item := range items {
		var block struct {
			Type  string            `json:"type"`
			Tools []json.RawMessage `json:"tools"`
		}
		if err := json.Unmarshal(item, &block); err != nil || block.Type != "additional_tools" {
			continue
		}
		for _, tool := range block.Tools {
			tools = append(tools, decodeResponsesTool(tool)...)
		}
	}
	return tools
}

func decodeResponsesTool(raw json.RawMessage) []OpenAITool {
	var head struct {
		Type  string            `json:"type"`
		Name  string            `json:"name"`
		Tools []json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil
	}
	if head.Type == "namespace" {
		var tools []OpenAITool
		for _, nested := range head.Tools {
			for _, tool := range decodeResponsesTool(nested) {
				if tool.Namespace == "" {
					tool.Namespace = head.Name
				}
				tools = append(tools, tool)
			}
		}
		return tools
	}

	var tool OpenAITool
	if err := json.Unmarshal(raw, &tool); err != nil {
		return nil
	}
	return []OpenAITool{tool}
}

func buildMessageFromInputItem(obj map[string]interface{}, role string) *OpenAIMessage {
	if role == "" {
		role = "user"
	}

	if content, ok := obj["content"]; ok {
		switch v := content.(type) {
		case string:
			return &OpenAIMessage{Role: role, Content: v}
		case []interface{}:
			parts := make([]interface{}, 0, len(v))
			textOnly := strings.Builder{}
			anyNonText := false
			for _, p := range v {
				part, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				ptype, _ := part["type"].(string)
				switch ptype {
				case "input_text", "text":
					if t, ok := part["text"].(string); ok {
						textOnly.WriteString(t)
						parts = append(parts, map[string]interface{}{"type": "input_text", "text": t})
					}
				case "output_text":
					if t, ok := part["text"].(string); ok {
						textOnly.WriteString(t)
						parts = append(parts, map[string]interface{}{"type": "input_text", "text": t})
					}
				case "input_image", "image", "image_url":
					anyNonText = true
					parts = append(parts, part)
				default:
					if t, ok := part["text"].(string); ok && t != "" {
						textOnly.WriteString(t)
						parts = append(parts, map[string]interface{}{"type": "input_text", "text": t})
					}
				}
			}
			if !anyNonText {
				return &OpenAIMessage{Role: role, Content: textOnly.String()}
			}
			return &OpenAIMessage{Role: role, Content: parts}
		case map[string]interface{}:
			return buildMessageFromInputItem(v, role)
		}
	}

	if text, ok := obj["text"].(string); ok && text != "" {
		return &OpenAIMessage{Role: role, Content: text}
	}

	return nil
}

func extractAgentMessageText(obj map[string]interface{}) string {
	content, _ := obj["content"].([]interface{})
	var parts []string
	for _, value := range content {
		part, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		switch part["type"] {
		case "input_text":
			if text, ok := part["text"].(string); ok && text != "" {
				parts = append(parts, text)
			}
		case "encrypted_content":
			// For non-ChatGPT providers Codex does not negotiate encryption;
			// this field contains the plaintext agent task.
			if text, ok := part["encrypted_content"].(string); ok && text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n")
}

func stringifyArbitrary(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func stringField(obj map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
