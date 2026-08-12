package proxy

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
)

const maxKiroToolNameLen = 64

type kiroToolBinding struct {
	Tool     OpenAITool
	KiroName string
}

type responseToolCallIdentity struct {
	Type      string
	Name      string
	Namespace string
}

// decodeResponsesTopLevelTools expands Responses namespace wrappers from the
// top-level tools array. Newer Codex builds use this shape for MCP/plugin tools,
// while some internal tools still arrive through input[].additional_tools.
func decodeResponsesTopLevelTools(body []byte) ([]OpenAITool, error) {
	var envelope struct {
		Tools []json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}

	var tools []OpenAITool
	for _, raw := range envelope.Tools {
		tools = append(tools, decodeResponsesTool(raw)...)
	}
	return tools, nil
}

func mergeResponsesTools(groups ...[]OpenAITool) []OpenAITool {
	seen := make(map[string]bool)
	var merged []OpenAITool
	for _, group := range groups {
		for _, tool := range group {
			identity := responseToolIdentity(tool)
			if seen[identity] {
				continue
			}
			seen[identity] = true
			merged = append(merged, tool)
		}
	}
	return merged
}

func responseToolIdentity(tool OpenAITool) string {
	return tool.Type + "\x00" + tool.Namespace + "\x00" + tool.Function.Name
}

func buildKiroToolBindings(tools []OpenAITool) []kiroToolBinding {
	nameCounts := make(map[string]int)
	for _, tool := range tools {
		if isKiroClientTool(tool) {
			nameCounts[tool.Function.Name]++
		}
	}

	usedNames := make(map[string]bool)
	bindings := make([]kiroToolBinding, 0, len(tools))
	for _, tool := range tools {
		if !isKiroClientTool(tool) {
			continue
		}

		baseName := shortenToolName(tool.Function.Name)
		if nameCounts[tool.Function.Name] > 1 {
			qualifier := tool.Namespace
			if qualifier == "" {
				qualifier = tool.Type
			}
			baseName = qualifier + "__" + tool.Function.Name
		}

		bindings = append(bindings, kiroToolBinding{
			Tool:     tool,
			KiroName: uniqueKiroToolName(baseName, responseToolIdentity(tool), usedNames),
		})
	}
	return bindings
}

func isKiroClientTool(tool OpenAITool) bool {
	return (tool.Type == "function" || tool.Type == "custom") &&
		strings.TrimSpace(tool.Function.Name) != ""
}

func uniqueKiroToolName(baseName, identity string, used map[string]bool) string {
	if len(baseName) <= maxKiroToolNameLen && !used[baseName] {
		used[baseName] = true
		return baseName
	}

	for attempt := 0; ; attempt++ {
		hashInput := identity
		if attempt > 0 {
			hashInput = fmt.Sprintf("%s#%d", identity, attempt)
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(hashInput))
		suffix := fmt.Sprintf("__%08x", h.Sum32())
		prefixLimit := maxKiroToolNameLen - len(suffix)
		prefix := baseName
		if len(prefix) > prefixLimit {
			prefix = prefix[:prefixLimit]
		}
		candidate := prefix + suffix
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}

func kiroToolBindingsByName(tools []OpenAITool) map[string]kiroToolBinding {
	bindings := buildKiroToolBindings(tools)
	byName := make(map[string]kiroToolBinding, len(bindings))
	for _, binding := range bindings {
		byName[binding.KiroName] = binding
	}
	return byName
}

func resolveResponseToolCall(bindings map[string]kiroToolBinding, kiroName string) responseToolCallIdentity {
	if binding, ok := bindings[kiroName]; ok {
		return responseToolCallIdentity{
			Type:      binding.Tool.Type,
			Name:      binding.Tool.Function.Name,
			Namespace: binding.Tool.Namespace,
		}
	}
	return responseToolCallIdentity{Type: "function", Name: kiroName}
}

func rewriteResponsesToolCallNames(messages []OpenAIMessage, tools []OpenAITool) {
	bindings := buildKiroToolBindings(tools)
	for i := range messages {
		for j := range messages[i].ToolCalls {
			call := &messages[i].ToolCalls[j]
			if alias, ok := findKiroToolName(bindings, call.Namespace, call.Function.Name); ok {
				call.Function.Name = alias
			}
		}
	}
}

func findKiroToolName(bindings []kiroToolBinding, namespace, name string) (string, bool) {
	if namespace != "" {
		for _, binding := range bindings {
			if binding.Tool.Namespace == namespace && binding.Tool.Function.Name == name {
				return binding.KiroName, true
			}
		}
		return "", false
	}

	var match string
	for _, binding := range bindings {
		if binding.Tool.Function.Name != name {
			continue
		}
		if match != "" {
			return "", false
		}
		match = binding.KiroName
	}
	return match, match != ""
}
