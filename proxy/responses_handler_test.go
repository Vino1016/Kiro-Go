package proxy

import (
	"context"
	"encoding/json"
	"io"
	"kiro-go/config"
	accountpool "kiro-go/pool"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResponsesParseStringInput(t *testing.T) {
	raw := json.RawMessage(`"hello world"`)
	msgs, err := parseResponsesInput(raw)
	if err != nil {
		t.Fatalf("parse string input: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != "user" {
		t.Fatalf("expected user role, got %q", msgs[0].Role)
	}
	if got, _ := msgs[0].Content.(string); got != "hello world" {
		t.Fatalf("expected hello world, got %v", msgs[0].Content)
	}
}

func TestResponsesParseArrayInput(t *testing.T) {
	raw := json.RawMessage(`[
		{"type":"message","role":"user","content":[{"type":"input_text","text":"first"}]},
		{"type":"input_text","text":"loose part"},
		{"type":"function_call_output","call_id":"call_1","output":"42"}
	]`)
	msgs, err := parseResponsesInput(raw)
	if err != nil {
		t.Fatalf("parse array input: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d (msgs=%+v)", len(msgs), msgs)
	}
	if msgs[0].Role != "user" {
		t.Fatalf("expected first message user, got %q", msgs[0].Role)
	}
	if got, _ := msgs[0].Content.(string); got != "first" {
		t.Fatalf("expected first text, got %v", msgs[0].Content)
	}
	if msgs[2].Role != "tool" || msgs[2].ToolCallID != "call_1" {
		t.Fatalf("expected tool result with call_id call_1, got %+v", msgs[2])
	}
	if got, _ := msgs[2].Content.(string); got != "42" {
		t.Fatalf("expected tool output 42, got %v", msgs[2].Content)
	}
}

func TestResponsesStoreAndLoad(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "config.json")
	if err := config.Init(cfgFile); err != nil {
		t.Fatalf("config.Init: %v", err)
	}

	resp := &ResponsesObject{
		ID:        "resp_unit_test_001",
		Object:    "response",
		CreatedAt: time.Now().Unix(),
		Status:    "completed",
		Model:     "claude-sonnet-4.5",
		Output: []ResponseOutputItem{{
			ID:   "msg_x",
			Type: "message",
			Role: "assistant",
			Content: []ResponseContentPart{{
				Type: "output_text",
				Text: "stored hello",
			}},
		}},
		StoredInput: json.RawMessage(`"hi"`),
	}

	if err := saveResponse(resp); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := loadResponse(resp.ID)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.ID != resp.ID || loaded.Model != resp.Model {
		t.Fatalf("loaded mismatch: %+v", loaded)
	}
	if len(loaded.Output) != 1 || loaded.Output[0].Content[0].Text != "stored hello" {
		t.Fatalf("loaded output mismatch: %+v", loaded.Output)
	}
	if string(loaded.StoredInput) != `"hi"` {
		t.Fatalf("stored input mismatch: %s", string(loaded.StoredInput))
	}

	if _, err := loadResponse("does_not_exist"); err == nil {
		t.Fatalf("expected load error for missing id")
	}
}

func TestResponsesNonStreamPreservesUpstreamIncompleteStatus(t *testing.T) {
	h, cleanup := setupResponsesTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(awsEventStreamFrame(t, "assistantResponseEvent", map[string]interface{}{
			"content": "partial answer",
		}))
		_, _ = w.Write(awsEventStreamFrame(t, "metadataEvent", map[string]interface{}{
			"stopReason": "MAX_TOKENS",
		}))
	}))
	defer server.Close()
	defer swapKiroEndpointsForTest(t, server)()

	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{
		"model":"claude-sonnet-4.5",
		"input":"hello"
	}`))
	rec := httptest.NewRecorder()
	h.handleOpenAIResponses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var response ResponsesObject
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	if response.Status != "incomplete" {
		t.Fatalf("status=%q, want incomplete", response.Status)
	}
	if response.IncompleteDetails == nil || response.IncompleteDetails.Reason != "max_output_tokens" {
		t.Fatalf("incomplete_details=%#v, want max_output_tokens", response.IncompleteDetails)
	}
}

func TestMapResponsesCompletion(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason string
		status string
		detail string
	}{
		{name: "max tokens", reason: "MAX_TOKENS", status: "incomplete", detail: "max_output_tokens"},
		{name: "max output tokens alias", reason: "max_output_tokens", status: "incomplete", detail: "max_output_tokens"},
		{name: "length alias", reason: "length", status: "incomplete", detail: "max_output_tokens"},
		{name: "context limit", reason: "CONTEXT_WINDOW_EXCEEDED", status: "incomplete", detail: "max_output_tokens"},
		{name: "model context limit alias", reason: "model_context_window_exceeded", status: "incomplete", detail: "max_output_tokens"},
		{name: "content filter", reason: "CONTENT_FILTERED", status: "incomplete", detail: "content_filter"},
		{name: "refusal alias", reason: "refusal", status: "incomplete", detail: "content_filter"},
		{name: "guardrail alias", reason: "guardrail_intervened", status: "incomplete", detail: "content_filter"},
		// END_TURN is the literal the Kiro IDE bundle compares against.
		{name: "end turn", reason: "END_TURN", status: "completed"},
		// Absent metadataEvent: reported as completed here on purpose. Detecting
		// that as truncation is the stream integrity layer's job, not this map's.
		{name: "empty reports completed", reason: "", status: "completed"},
		{name: "normal completion", reason: "COMPLETE", status: "completed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, detail := mapResponsesCompletion(tc.reason)
			if status != tc.status || detail != tc.detail {
				t.Fatalf("mapResponsesCompletion(%q) = (%q, %q), want (%q, %q)", tc.reason, status, detail, tc.status, tc.detail)
			}
		})
	}
}

func TestResponsesPreviousResponseIDExpands(t *testing.T) {
	prev := &ResponsesObject{
		ID:          "resp_prev",
		StoredInput: json.RawMessage(`"earlier user"`),
		Output: []ResponseOutputItem{
			{
				Type: "message",
				Role: "assistant",
				Content: []ResponseContentPart{{
					Type: "output_text",
					Text: "earlier assistant reply",
				}},
			},
			{
				Type:      "function_call",
				CallID:    "call_prev",
				Name:      "lookup",
				Arguments: `{"q":"x"}`,
			},
		},
	}

	expanded := expandPreviousResponseHistory(prev)
	if len(expanded) != 3 {
		t.Fatalf("expected 3 messages from history, got %d (%+v)", len(expanded), expanded)
	}
	if expanded[0].Role != "user" {
		t.Fatalf("expected first message to be user, got %+v", expanded[0])
	}
	if expanded[1].Role != "assistant" {
		t.Fatalf("expected second message to be assistant, got %+v", expanded[1])
	}
	if expanded[2].Role != "assistant" || len(expanded[2].ToolCalls) != 1 {
		t.Fatalf("expected third to be assistant with tool_calls, got %+v", expanded[2])
	}
	if expanded[2].ToolCalls[0].ID != "call_prev" {
		t.Fatalf("expected tool call id call_prev, got %+v", expanded[2].ToolCalls[0])
	}
}

// A → B → C: when expanding history starting from C, all of A's and B's
// inputs/outputs must appear before C's. Previously only C's direct parent
// (B) was emitted, dropping A entirely.
func TestResponsesPreviousResponseIDExpandsFullChain(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "config.json")
	if err := config.Init(cfgFile); err != nil {
		t.Fatalf("config.Init: %v", err)
	}

	a := &ResponsesObject{
		ID:           "resp_a",
		Object:       "response",
		Status:       "completed",
		Model:        "claude-sonnet-4.5",
		StoredInput:  json.RawMessage(`"turn A user"`),
		StoredAt:     time.Now().Unix(),
		Instructions: "be terse",
		Output: []ResponseOutputItem{{
			Type: "message", Role: "assistant",
			Content: []ResponseContentPart{{Type: "output_text", Text: "turn A assistant"}},
		}},
	}
	b := &ResponsesObject{
		ID:                 "resp_b",
		Object:             "response",
		Status:             "completed",
		Model:              "claude-sonnet-4.5",
		StoredInput:        json.RawMessage(`"turn B user"`),
		StoredAt:           time.Now().Unix(),
		PreviousResponseID: a.ID,
		Output: []ResponseOutputItem{{
			Type: "message", Role: "assistant",
			Content: []ResponseContentPart{{Type: "output_text", Text: "turn B assistant"}},
		}},
	}
	if err := saveResponse(a); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := saveResponse(b); err != nil {
		t.Fatalf("save b: %v", err)
	}

	expanded := expandPreviousResponseHistory(b)

	var transcript []string
	for _, m := range expanded {
		role := m.Role
		text, _ := m.Content.(string)
		transcript = append(transcript, role+":"+text)
	}
	got := strings.Join(transcript, "|")
	want := "system:be terse|user:turn A user|assistant:turn A assistant|user:turn B user|assistant:turn B assistant"
	if got != want {
		t.Fatalf("chain order mismatch:\n got=%s\nwant=%s", got, want)
	}
}

// New instructions sent on a continuation request must take effect, even when
// previous_response_id is set. The bug: the old code only attached
// req.Instructions when previous_response_id was empty, silently dropping
// updated system prompts on follow-up turns.
func TestResponsesContinuationKeepsNewInstructions(t *testing.T) {
	h, cleanup := setupResponsesTestHandler(t)
	defer cleanup()

	prev := &ResponsesObject{
		ID:          "resp_for_continuation",
		Object:      "response",
		Status:      "completed",
		Model:       "claude-sonnet-4.5",
		StoredInput: json.RawMessage(`"first user message"`),
		StoredAt:    time.Now().Unix(),
		Output: []ResponseOutputItem{{
			Type: "message", Role: "assistant",
			Content: []ResponseContentPart{{Type: "output_text", Text: "first reply"}},
		}},
	}
	if err := saveResponse(prev); err != nil {
		t.Fatalf("save prev: %v", err)
	}

	var capturedSystem string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		capturedSystem = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(awsEventStreamFrame(t, "assistantResponseEvent", map[string]interface{}{
			"content": "second reply",
		}))
		_, _ = w.Write(awsEventStreamFrame(t, "metadataEvent", map[string]interface{}{
			"stopReason": "end_turn",
		}))
	}))
	defer server.Close()
	defer swapKiroEndpointsForTest(t, server)()

	body := strings.NewReader(`{
		"model":"claude-sonnet-4.5",
		"input":"second user turn",
		"previous_response_id":"resp_for_continuation",
		"instructions":"speak only French",
		"store":false
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", body)
	rec := httptest.NewRecorder()
	h.handleOpenAIResponses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(capturedSystem, "speak only French") {
		t.Fatalf("expected new instructions to reach upstream, payload=%s", capturedSystem)
	}
}

func setupResponsesTestHandler(t *testing.T) (*Handler, func()) {
	t.Helper()
	cfgFile := filepath.Join(t.TempDir(), "config.json")
	if err := config.Init(cfgFile); err != nil {
		t.Fatalf("config.Init: %v", err)
	}
	if err := config.AddAccount(config.Account{
		ID:          "test-account",
		Enabled:     true,
		AccessToken: "token-test",
		ProfileArn:  "arn:aws:codewhisperer:profile/test",
	}); err != nil {
		t.Fatalf("add account: %v", err)
	}
	if err := config.UpdatePreferredEndpoint("kiro"); err != nil {
		t.Fatalf("set endpoint: %v", err)
	}
	if err := config.UpdateEndpointFallback(false); err != nil {
		t.Fatalf("disable fallback: %v", err)
	}
	p := accountpool.GetPool()
	p.Reload()
	h := &Handler{
		pool:        p,
		promptCache: newPromptCacheTracker(defaultPromptCacheTTL),
	}
	cleanup := func() {}
	return h, cleanup
}

func swapKiroEndpointsForTest(t *testing.T, server *httptest.Server) func() {
	t.Helper()
	oldEndpoints := kiroEndpoints
	kiroEndpoints = []kiroEndpoint{{
		URL:    server.URL,
		Origin: "AI_EDITOR",
		Name:   "test",
	}}
	oldClient := kiroHttpStore.Load()
	kiroHttpStore.Store(&http.Client{Timeout: time.Second, Transport: &http.Transport{}})
	return func() {
		kiroEndpoints = oldEndpoints
		kiroHttpStore.Store(oldClient)
	}
}

func TestResponsesNonStreamRoundTrip(t *testing.T) {
	h, cleanup := setupResponsesTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(awsEventStreamFrame(t, "assistantResponseEvent", map[string]interface{}{
			"content": "responses non-stream OK",
		}))
		_, _ = w.Write(awsEventStreamFrame(t, "metadataEvent", map[string]interface{}{
			"stopReason": "end_turn",
		}))
	}))
	defer server.Close()
	defer swapKiroEndpointsForTest(t, server)()

	body := strings.NewReader(`{"model":"claude-sonnet-4.5","input":"hi from test"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", body)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	h.handleOpenAIResponses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp ResponsesObject
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v body=%s", err, rec.Body.String())
	}
	if resp.Object != "response" {
		t.Fatalf("expected object=response, got %q", resp.Object)
	}
	if resp.Status != "completed" {
		t.Fatalf("expected status=completed, got %q", resp.Status)
	}
	if len(resp.Output) == 0 {
		t.Fatalf("expected output items, got none")
	}
	if resp.Output[0].Type != "message" || len(resp.Output[0].Content) == 0 {
		t.Fatalf("expected message with content, got %+v", resp.Output[0])
	}
	if resp.Output[0].Content[0].Text != "responses non-stream OK" {
		t.Fatalf("unexpected text: %q", resp.Output[0].Content[0].Text)
	}

	loaded, err := loadResponse(resp.ID)
	if err != nil {
		t.Fatalf("loadResponse: %v", err)
	}
	if loaded.ID != resp.ID {
		t.Fatalf("stored response id mismatch")
	}
}

func TestResponsesStreamPreservesUpstreamIncompleteStatus(t *testing.T) {
	h, cleanup := setupResponsesTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(awsEventStreamFrame(t, "assistantResponseEvent", map[string]interface{}{
			"content": "partial answer",
		}))
		_, _ = w.Write(awsEventStreamFrame(t, "metadataEvent", map[string]interface{}{
			"stopReason": "MAX_TOKENS",
		}))
	}))
	defer server.Close()
	defer swapKiroEndpointsForTest(t, server)()

	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{
		"model":"claude-sonnet-4.5",
		"input":"hello",
		"stream":true,
		"store":false
	}`))
	rec := httptest.NewRecorder()
	h.handleOpenAIResponses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"status":"incomplete"`) || !strings.Contains(rec.Body.String(), `"reason":"max_output_tokens"`) {
		t.Fatalf("expected incomplete response.completed event, got %s", rec.Body.String())
	}
}

func TestResponsesStreamSSE(t *testing.T) {
	h, cleanup := setupResponsesTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(awsEventStreamFrame(t, "assistantResponseEvent", map[string]interface{}{
			"content": "stream chunk",
		}))
		_, _ = w.Write(awsEventStreamFrame(t, "metadataEvent", map[string]interface{}{
			"stopReason": "end_turn",
		}))
	}))
	defer server.Close()
	defer swapKiroEndpointsForTest(t, server)()

	body := strings.NewReader(`{"model":"claude-sonnet-4.5","input":"stream please","stream":true,"store":false}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", body)
	rec := httptest.NewRecorder()

	h.handleOpenAIResponses(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	bodyBytes, _ := io.ReadAll(rec.Body)
	bodyStr := string(bodyBytes)

	for _, evt := range []string{"event: response.created", "event: response.output_text.delta", "event: response.completed"} {
		if !strings.Contains(bodyStr, evt) {
			t.Fatalf("missing event %q in stream body:\n%s", evt, bodyStr)
		}
	}
	if !strings.Contains(bodyStr, "stream chunk") {
		t.Fatalf("expected stream content delta, got:\n%s", bodyStr)
	}
}

func TestResponsesNamespacedToolsRoundTrip(t *testing.T) {
	for _, stream := range []bool{false, true} {
		name := "non-stream"
		if stream {
			name = "stream"
		}
		t.Run(name, func(t *testing.T) {
			h, cleanup := setupResponsesTestHandler(t)
			defer cleanup()

			var capturedToolNames []string
			var payloadErr string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var payload KiroPayload
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					payloadErr = err.Error()
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				ctx := payload.ConversationState.CurrentMessage.UserInputMessage.UserInputMessageContext
				if ctx == nil || len(ctx.Tools) < 2 {
					payloadErr = "expected two namespaced tools in Kiro payload"
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				for _, tool := range ctx.Tools {
					capturedToolNames = append(capturedToolNames, tool.ToolSpecification.Name)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(awsEventStreamFrame(t, "toolUseEvent", map[string]interface{}{
					"toolUseId": "call_canva",
					"name":      capturedToolNames[0],
					"input":     `{"query":"deck"}`,
					"stop":      true,
				}))
				_, _ = w.Write(awsEventStreamFrame(t, "metadataEvent", map[string]interface{}{
					"stopReason": "tool_use",
				}))
			}))
			defer server.Close()
			defer swapKiroEndpointsForTest(t, server)()

			requestBody, err := json.Marshal(map[string]interface{}{
				"model":  "claude-sonnet-4.5",
				"input":  "find a deck",
				"stream": stream,
				"store":  false,
				"tools": []interface{}{
					map[string]interface{}{
						"type": "namespace", "name": "mcp__canva",
						"tools": []interface{}{map[string]interface{}{
							"type": "function", "name": "_search", "description": "Canva search",
							"parameters": map[string]interface{}{"type": "object"},
						}},
					},
					map[string]interface{}{
						"type": "namespace", "name": "mcp__github",
						"tools": []interface{}{map[string]interface{}{
							"type": "function", "name": "_search", "description": "GitHub search",
							"parameters": map[string]interface{}{"type": "object"},
						}},
					},
				},
			})
			if err != nil {
				t.Fatalf("encode request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(string(requestBody)))
			rec := httptest.NewRecorder()
			h.handleOpenAIResponses(rec, req)

			if payloadErr != "" {
				t.Fatalf("invalid Kiro payload: %s", payloadErr)
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
			}
			if len(capturedToolNames) != 2 || capturedToolNames[0] == capturedToolNames[1] {
				t.Fatalf("expected two unique Kiro aliases, got %v", capturedToolNames)
			}
			if capturedToolNames[0] == "_search" || capturedToolNames[1] == "_search" {
				t.Fatalf("duplicate bare tool names were not disambiguated: %v", capturedToolNames)
			}

			if stream {
				body := rec.Body.String()
				if !strings.Contains(body, `"name":"_search"`) || !strings.Contains(body, `"namespace":"mcp__canva"`) {
					t.Fatalf("stream did not restore Codex tool identity:\n%s", body)
				}
				if strings.Contains(body, capturedToolNames[0]) {
					t.Fatalf("stream leaked Kiro alias %q:\n%s", capturedToolNames[0], body)
				}
				return
			}

			var resp ResponsesObject
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
			}
			if len(resp.Output) != 1 || resp.Output[0].Type != "function_call" ||
				resp.Output[0].Name != "_search" || resp.Output[0].Namespace != "mcp__canva" {
				t.Fatalf("non-stream response did not restore Codex tool identity: %+v", resp.Output)
			}
		})
	}
}

func TestExtractResponsesToolsAdditionalTools(t *testing.T) {
	raw := json.RawMessage(`[
		{
			"type":"additional_tools",
			"role":"developer",
			"tools":[
				{"type":"custom","name":"exec","description":"Run JS"},
				{"type":"function","name":"wait","description":"Wait","parameters":{"type":"object","properties":{"cell_id":{"type":"string"}}}},
				{"type":"namespace","name":"collaboration","tools":[
					{"type":"function","name":"spawn_agent","description":"Spawn","parameters":{"type":"object","properties":{"message":{"type":"string","encrypted":true}}}}
				]}
			]
		},
		{"type":"message","role":"user","content":"hi"}
	]`)

	tools := extractResponsesTools(raw)
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d (%+v)", len(tools), tools)
	}
	byName := make(map[string]OpenAITool, len(tools))
	for _, tool := range tools {
		byName[tool.Function.Name] = tool
	}
	if tool := byName["exec"]; tool.Type != "custom" {
		t.Fatalf("expected custom exec tool, got %+v", tool)
	}
	if tool := byName["spawn_agent"]; tool.Type != "function" || tool.Namespace != "collaboration" {
		t.Fatalf("expected collaboration spawn_agent, got %+v", tool)
	}
}

func TestConvertOpenAIToolsSupportsCustomAndCleansEncrypted(t *testing.T) {
	custom := OpenAITool{Type: "custom"}
	custom.Function.Name = "exec"
	custom.Function.Description = "Run JS"

	function := OpenAITool{Type: "function"}
	function.Function.Name = "spawn_agent"
	function.Function.Parameters = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{"type": "string", "encrypted": true},
		},
	}

	wrapped := convertOpenAITools([]OpenAITool{custom, function})
	if len(wrapped) != 2 {
		t.Fatalf("expected 2 converted tools, got %d", len(wrapped))
	}
	customSchema := wrapped[0].ToolSpecification.InputSchema.JSON.(map[string]interface{})
	customProps := customSchema["properties"].(map[string]interface{})
	if _, ok := customProps["input"]; !ok {
		t.Fatalf("custom tool schema missing input property: %+v", customSchema)
	}
	functionSchema := wrapped[1].ToolSpecification.InputSchema.JSON.(map[string]interface{})
	messageSchema := functionSchema["properties"].(map[string]interface{})["message"].(map[string]interface{})
	if _, ok := messageSchema["encrypted"]; ok {
		t.Fatalf("Kiro-incompatible encrypted keyword was not removed: %+v", messageSchema)
	}
}

func TestResponsesParseCustomToolCallAndAgentMessage(t *testing.T) {
	raw := json.RawMessage(`[
		{"type":"agent_message","author":"/root","recipient":"/root/worker","content":[
			{"type":"input_text","text":"Message Type: NEW_TASK"},
			{"type":"encrypted_content","encrypted_content":"reply with exactly: PONG"}
		]},
		{"type":"custom_tool_call","call_id":"call_1","name":"exec","input":"console.log(1)"},
		{"type":"custom_tool_call_output","call_id":"call_1","output":"1"}
	]`)

	msgs, err := parseResponsesInput(raw)
	if err != nil {
		t.Fatalf("parse Responses input: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d (%+v)", len(msgs), msgs)
	}
	if text, _ := msgs[0].Content.(string); !strings.Contains(text, "reply with exactly: PONG") {
		t.Fatalf("agent task was dropped: %q", text)
	}
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) != 1 {
		t.Fatalf("expected assistant custom tool call, got %+v", msgs[1])
	}
	var args map[string]string
	if err := json.Unmarshal([]byte(msgs[1].ToolCalls[0].Function.Arguments), &args); err != nil {
		t.Fatalf("decode custom tool arguments: %v", err)
	}
	if args["input"] != "console.log(1)" {
		t.Fatalf("unexpected custom tool input: %+v", args)
	}
	if msgs[2].Role != "tool" || msgs[2].ToolCallID != "call_1" || msgs[2].Content != "1" {
		t.Fatalf("unexpected custom tool output: %+v", msgs[2])
	}
}

func TestBuildResponsesObjectPreservesCustomTypeAndNamespace(t *testing.T) {
	custom := OpenAITool{Type: "custom"}
	custom.Function.Name = "exec"
	namespaced := OpenAITool{Type: "function", Namespace: "collaboration"}
	namespaced.Function.Name = "spawn_agent"
	req := &ResponsesRequest{Tools: []OpenAITool{custom, namespaced}}

	obj := buildResponsesObject("resp_1", "claude-sonnet-4.5", "", []KiroToolUse{
		{ToolUseID: "call_exec", Name: "exec", Input: map[string]interface{}{"input": "console.log(1)"}},
		{ToolUseID: "call_spawn", Name: "spawn_agent", Input: map[string]interface{}{"message": "hi"}},
	}, 0, 0, req, "tool_use")

	if len(obj.Output) != 2 {
		t.Fatalf("expected 2 output items, got %+v", obj.Output)
	}
	if obj.Output[0].Type != "custom_tool_call" || obj.Output[0].Input != "console.log(1)" {
		t.Fatalf("unexpected custom tool output: %+v", obj.Output[0])
	}
	if obj.Output[1].Type != "function_call" || obj.Output[1].Namespace != "collaboration" {
		t.Fatalf("unexpected namespaced function output: %+v", obj.Output[1])
	}
}
