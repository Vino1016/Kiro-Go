package proxy

import (
	"encoding/json"
	"testing"
)

func TestDecodeResponsesTopLevelToolsFlattensNamespaces(t *testing.T) {
	body := []byte(`{
		"model":"gpt-5.6-sol",
		"tools":[
			{"type":"function","name":"exec_command","parameters":{"type":"object"}},
			{"type":"namespace","name":"mcp__canva","tools":[
				{"type":"function","name":"_search","description":"Canva search","parameters":{"type":"object"}}
			]},
			{"type":"namespace","name":"mcp__github","tools":[
				{"type":"function","name":"_search","description":"GitHub search","parameters":{"type":"object"}}
			]}
		]
	}`)

	tools, err := decodeResponsesTopLevelTools(body)
	if err != nil {
		t.Fatalf("decode top-level tools: %v", err)
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 flattened tools, got %d (%+v)", len(tools), tools)
	}
	if tools[0].Function.Name != "exec_command" || tools[0].Namespace != "" {
		t.Fatalf("unexpected direct tool: %+v", tools[0])
	}
	if tools[1].Function.Name != "_search" || tools[1].Namespace != "mcp__canva" {
		t.Fatalf("unexpected first namespaced tool: %+v", tools[1])
	}
	if tools[2].Function.Name != "_search" || tools[2].Namespace != "mcp__github" {
		t.Fatalf("unexpected second namespaced tool: %+v", tools[2])
	}
}

func TestNamespacedToolBindingsDisambiguateAndRestoreIdentity(t *testing.T) {
	canva := OpenAITool{Type: "function", Namespace: "mcp__canva"}
	canva.Function.Name = "_search"
	canva.Function.Description = "Canva search"
	canva.Function.Parameters = map[string]interface{}{"type": "object"}

	github := OpenAITool{Type: "function", Namespace: "mcp__github"}
	github.Function.Name = "_search"
	github.Function.Description = "GitHub search"
	github.Function.Parameters = map[string]interface{}{"type": "object"}

	tools := []OpenAITool{canva, github}
	wrapped := convertOpenAITools(tools)
	if len(wrapped) != 2 {
		t.Fatalf("expected 2 converted tools, got %d", len(wrapped))
	}
	canvaAlias := wrapped[0].ToolSpecification.Name
	githubAlias := wrapped[1].ToolSpecification.Name
	if canvaAlias == githubAlias {
		t.Fatalf("duplicate response names must have unique Kiro aliases: %q", canvaAlias)
	}
	if len(canvaAlias) > 64 || len(githubAlias) > 64 {
		t.Fatalf("Kiro aliases exceed 64 characters: %q %q", canvaAlias, githubAlias)
	}

	req := &ResponsesRequest{Tools: tools}
	obj := buildResponsesObject("resp_1", "gpt-5.6-sol", "", []KiroToolUse{
		{ToolUseID: "call_canva", Name: canvaAlias, Input: map[string]interface{}{"query": "deck"}},
		{ToolUseID: "call_github", Name: githubAlias, Input: map[string]interface{}{"query": "repo"}},
	}, 0, 0, req, "tool_use")

	if len(obj.Output) != 2 {
		t.Fatalf("expected 2 response tool calls, got %+v", obj.Output)
	}
	if obj.Output[0].Name != "_search" || obj.Output[0].Namespace != "mcp__canva" {
		t.Fatalf("Canva identity was not restored: %+v", obj.Output[0])
	}
	if obj.Output[1].Name != "_search" || obj.Output[1].Namespace != "mcp__github" {
		t.Fatalf("GitHub identity was not restored: %+v", obj.Output[1])
	}
}

func TestRewriteResponsesHistoryToolNameUsesKiroAlias(t *testing.T) {
	tool := OpenAITool{Type: "function", Namespace: "mcp__codegraph"}
	tool.Function.Name = "search_graph"
	tool.Function.Parameters = map[string]interface{}{"type": "object"}

	input := json.RawMessage(`[
		{"type":"function_call","call_id":"call_1","namespace":"mcp__codegraph","name":"search_graph","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_1","output":"ok"}
	]`)
	messages, err := parseResponsesInput(input)
	if err != nil {
		t.Fatalf("parse response history: %v", err)
	}
	if len(messages) != 2 || len(messages[0].ToolCalls) != 1 {
		t.Fatalf("unexpected parsed history: %+v", messages)
	}

	rewriteResponsesToolCallNames(messages, []OpenAITool{tool})
	want := convertOpenAITools([]OpenAITool{tool})[0].ToolSpecification.Name
	if got := messages[0].ToolCalls[0].Function.Name; got != want {
		t.Fatalf("expected Kiro alias %q, got %q", want, got)
	}
}
