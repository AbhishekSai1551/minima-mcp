package minima

import (
	"testing"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		addr string
		err  bool
	}{
		{"0xabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", false},
		{"Mxabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", false},
		{"", true},
		{"short", true},
		{"0xGGGG", true},
	}

	for _, tt := range tests {
		err := ValidateAddress(tt.addr)
		if tt.err && err == nil {
			t.Errorf("ValidateAddress(%q) expected error, got nil", tt.addr)
		}
		if !tt.err && err != nil {
			t.Errorf("ValidateAddress(%q) unexpected error: %v", tt.addr, err)
		}
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		amount string
		err    bool
	}{
		{"10", false},
		{"0", false},
		{"0.5", false},
		{"1000000", false},
		{"", true},
		{"-1", true},
		{"abc", true},
	}

	for _, tt := range tests {
		err := ValidateAmount(tt.amount)
		if tt.err && err == nil {
			t.Errorf("ValidateAmount(%q) expected error, got nil", tt.amount)
		}
		if !tt.err && err != nil {
			t.Errorf("ValidateAmount(%q) unexpected error: %v", tt.amount, err)
		}
	}
}

func TestValidateScript(t *testing.T) {
	tests := []struct {
		script string
		err    bool
	}{
		{"function test() { return 1; }", false},
		{"", true},
		{"function test() { runtime.exec('rm -rf /') }", true},
		{"System.exit(0)", true},
	}

	for _, tt := range tests {
		err := ValidateScript(tt.script)
		if tt.err && err == nil {
			t.Errorf("ValidateScript() expected error, got nil")
		}
		if !tt.err && err != nil {
			t.Errorf("ValidateScript() unexpected error: %v", err)
		}
	}
}

func TestValidateContractID(t *testing.T) {
	tests := []struct {
		id  string
		err bool
	}{
		{"abc123def456", false},
		{"", true},
		{"a", true},
		{"valid_contract_id_123", false},
		{"invalid!@#", true},
	}

	for _, tt := range tests {
		err := ValidateContractID(tt.id)
		if tt.err && err == nil {
			t.Errorf("ValidateContractID(%q) expected error, got nil", tt.id)
		}
		if !tt.err && err != nil {
			t.Errorf("ValidateContractID(%q) unexpected error: %v", tt.id, err)
		}
	}
}

func TestValidateTokenName(t *testing.T) {
	tests := []struct {
		name string
		err  bool
	}{
		{"MyToken", false},
		{"AB", false},
		{"", true},
		{"a", true},
		{"Token-With Spaces_123", false},
		{"bad!token", true},
	}

	for _, tt := range tests {
		err := ValidateTokenName(tt.name)
		if tt.err && err == nil {
			t.Errorf("ValidateTokenName(%q) expected error, got nil", tt.name)
		}
		if !tt.err && err != nil {
			t.Errorf("ValidateTokenName(%q) unexpected error: %v", tt.name, err)
		}
	}
}

func TestValidateFunctionName(t *testing.T) {
	tests := []struct {
		name string
		err  bool
	}{
		{"transfer", false},
		{"_init", false},
		{"", true},
		{"123bad", true},
		{"valid_name_123", false},
		{"has space", true},
	}

	for _, tt := range tests {
		err := ValidateFunctionName(tt.name)
		if tt.err && err == nil {
			t.Errorf("ValidateFunctionName(%q) expected error, got nil", tt.name)
		}
		if !tt.err && err != nil {
			t.Errorf("ValidateFunctionName(%q) unexpected error: %v", tt.name, err)
		}
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{int64(100), "100"},
		{true, "true"},
		{nil, ""},
	}

	for _, tt := range tests {
		result := toString(tt.input)
		if result != tt.expected {
			t.Errorf("toString(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToInt64(t *testing.T) {
	if toInt64(float64(42)) != 42 {
		t.Error("toInt64(float64(42)) != 42")
	}
	if toInt64("100") != 100 {
		t.Error("toInt64(\"100\") != 100")
	}
	if toInt64("invalid") != 0 {
		t.Error("toInt64(\"invalid\") != 0")
	}
}

func TestToBool(t *testing.T) {
	if !toBool(true) {
		t.Error("toBool(true) != true")
	}
	if toBool(false) {
		t.Error("toBool(false) != false")
	}
	if !toBool("true") {
		t.Error("toBool(\"true\") != true")
	}
	if toBool("false") {
		t.Error("toBool(\"false\") != false")
	}
}
