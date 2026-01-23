package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmalldedede/agentbox/internal/engine"
)

func TestParseJSONLOutput_BasicMessage(t *testing.T) {
	adapter := New()
	output := `{"type":"system","subtype":"init","session_id":"abc-123-def","tools":[],"mcp_servers":[]}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello! How can I help you?"}]}}
{"type":"result","subtype":"success","session_id":"abc-123-def","usage":{"input_tokens":100,"output_tokens":10}}`

	result, err := adapter.ParseJSONLOutput(output, false)
	require.NoError(t, err)

	assert.Equal(t, "Hello! How can I help you?", result.Message)
	assert.Equal(t, "abc-123-def", result.ThreadID)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 100, result.Usage.InputTokens)
	assert.Equal(t, 10, result.Usage.OutputTokens)
	assert.Empty(t, result.Error)
}

func TestParseJSONLOutput_MultipleTextBlocks(t *testing.T) {
	adapter := New()
	output := `{"type":"system","subtype":"init","session_id":"sess-multi"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Part 1"},{"type":"text","text":"Part 2"}]}}
{"type":"result","subtype":"success","session_id":"sess-multi"}`

	result, err := adapter.ParseJSONLOutput(output, false)
	require.NoError(t, err)

	assert.Equal(t, "Part 1\nPart 2", result.Message)
	assert.Equal(t, "sess-multi", result.ThreadID)
}

func TestParseJSONLOutput_ErrorResult(t *testing.T) {
	adapter := New()
	output := `{"type":"system","subtype":"init","session_id":"sess-err"}
{"type":"result","subtype":"error","session_id":"sess-err","error":"max turns exceeded"}`

	result, err := adapter.ParseJSONLOutput(output, false)
	require.NoError(t, err)

	assert.Equal(t, "sess-err", result.ThreadID)
	assert.Equal(t, "max turns exceeded", result.Error)
}

func TestParseJSONLOutput_WithDockerHeaders(t *testing.T) {
	// Docker stream multiplexing adds 8-byte headers before each frame.
	// The parser uses strings.Index(line, "{") to skip these.
	adapter := New()
	output := "\x01\x00\x00\x00\x00\x00\x00\x50" +
		`{"type":"system","subtype":"init","session_id":"docker-sess"}` + "\n" +
		"\x01\x00\x00\x00\x00\x00\x00\x40" +
		`{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Works!"}]}}` + "\n" +
		"\x01\x00\x00\x00\x00\x00\x00\x30" +
		`{"type":"result","subtype":"success","session_id":"docker-sess"}`

	result, err := adapter.ParseJSONLOutput(output, false)
	require.NoError(t, err)

	assert.Equal(t, "Works!", result.Message)
	assert.Equal(t, "docker-sess", result.ThreadID)
}

func TestParseJSONLOutput_PlainTextFallback(t *testing.T) {
	adapter := New()
	output := "Just plain text output without any JSON"

	result, err := adapter.ParseJSONLOutput(output, false)
	require.NoError(t, err)

	assert.Equal(t, "Just plain text output without any JSON", result.Message)
	assert.Empty(t, result.ThreadID)
}

func TestParseJSONLOutput_WithEvents(t *testing.T) {
	adapter := New()
	output := `{"type":"system","subtype":"init","session_id":"evt-sess"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","session_id":"evt-sess","usage":{"input_tokens":50,"output_tokens":5}}`

	result, err := adapter.ParseJSONLOutput(output, true)
	require.NoError(t, err)

	assert.Equal(t, "Hello", result.Message)
	assert.Equal(t, "evt-sess", result.ThreadID)
	assert.Len(t, result.Events, 3)
	assert.Equal(t, "system", result.Events[0].Type)
	assert.Equal(t, "assistant", result.Events[1].Type)
	assert.Equal(t, "result", result.Events[2].Type)
}

func TestPrepareExec_Resume(t *testing.T) {
	adapter := New()
	opts := &engine.ExecOptions{
		Prompt:   "What did I ask before?",
		ThreadID: "sess-12345",
		MaxTurns: 5,
	}

	args := adapter.PrepareExec(opts)

	assert.Contains(t, args, "--resume")
	assert.Contains(t, args, "sess-12345")
	assert.Contains(t, args, "--verbose")
	assert.Contains(t, args, "--output-format")
	assert.Contains(t, args, "stream-json")
	assert.Contains(t, args, "--max-turns")
}

func TestPrepareExec_NoResume(t *testing.T) {
	adapter := New()
	opts := &engine.ExecOptions{
		Prompt: "Hello",
	}

	args := adapter.PrepareExec(opts)

	assert.NotContains(t, args, "--resume")
	assert.Contains(t, args, "--verbose")
	assert.Contains(t, args, "stream-json")
}

func TestParseJSONLOutput_ErrorMaxTurns(t *testing.T) {
	adapter := New()
	output := `{"type":"system","subtype":"init","session_id":"sess-max"}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Working on it..."}]}}
{"type":"result","subtype":"error_max_turns","session_id":"sess-max","error":"reached maximum turns","usage":{"input_tokens":200,"output_tokens":50}}`

	result, err := adapter.ParseJSONLOutput(output, false)
	require.NoError(t, err)

	assert.Equal(t, "Working on it...", result.Message)
	assert.Equal(t, "sess-max", result.ThreadID)
	assert.Equal(t, "reached maximum turns", result.Error)
	assert.NotNil(t, result.Usage)
	assert.Equal(t, 200, result.Usage.InputTokens)
	assert.Equal(t, 50, result.Usage.OutputTokens)
}
