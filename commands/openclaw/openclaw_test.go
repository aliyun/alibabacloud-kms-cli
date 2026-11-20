package openclaw

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestWriteErrorResponse(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writeErrorResponse("global", "test error message")

	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	var resp ExecProviderResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ProtocolVersion != protocolVersion {
		t.Errorf("ProtocolVersion = %d, want %d", resp.ProtocolVersion, protocolVersion)
	}
	if len(resp.Values) != 0 {
		t.Errorf("Values should be empty, got %v", resp.Values)
	}
	if resp.Errors["global"] == nil {
		t.Fatal("expected error for key 'global'")
	}
	if resp.Errors["global"].Message != "test error message" {
		t.Errorf("Error message = %s, want 'test error message'", resp.Errors["global"].Message)
	}
}

func TestExecProviderRequest_Unmarshal(t *testing.T) {
	input := `{"protocolVersion":1,"provider":"kms","ids":["secret1","secret2"]}`
	var req ExecProviderRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.ProtocolVersion != 1 {
		t.Errorf("ProtocolVersion = %d, want 1", req.ProtocolVersion)
	}
	if req.Provider != "kms" {
		t.Errorf("Provider = %s, want 'kms'", req.Provider)
	}
	if len(req.IDs) != 2 || req.IDs[0] != "secret1" || req.IDs[1] != "secret2" {
		t.Errorf("IDs = %v, want [secret1 secret2]", req.IDs)
	}
}

func TestExecProviderResponse_Marshal(t *testing.T) {
	resp := ExecProviderResponse{
		ProtocolVersion: 1,
		Values: map[string]string{
			"secret1": "value1",
		},
		Errors: map[string]*ErrorInfo{
			"secret2": {Message: "not found"},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"protocolVersion":1,"values":{"secret1":"value1"},"errors":{"secret2":{"message":"not found"}}}`
	if string(data) != expected {
		t.Errorf("marshaled JSON mismatch\ngot:      %s\nexpected: %s", string(data), expected)
	}
}

func TestExecProviderResponse_EmptyErrorsOmitted(t *testing.T) {
	resp := ExecProviderResponse{
		ProtocolVersion: 1,
		Values: map[string]string{
			"secret1": "value1",
		},
		Errors: map[string]*ErrorInfo{},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// errors is empty, should be omitted due to `omitempty`
	if bytes.Contains(data, []byte("errors")) {
		t.Errorf("empty errors should be omitted, got: %s", string(data))
	}
}
