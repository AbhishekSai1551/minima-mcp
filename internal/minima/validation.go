package minima

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	addressRegex    = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$|^Mx[0-9a-fA-F]{64}$`)
	txIDRegex       = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$`)
	contractIDRegex = regexp.MustCompile(`^[0-9a-zA-Z_]{8,128}$`)
	tokenIDRegex    = regexp.MustCompile(`^0x[0-9a-fA-F]{64}$|^[0-9a-zA-Z_\-]{4,64}$`)
	hexRegex        = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)
	functionRegex   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)
	tokenNameRegex  = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]{2,32}$`)
	miniDAPPIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-.]{4,128}$`)
)

func ValidateAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if len(addr) > 256 {
		return fmt.Errorf("address too long: max 256 chars")
	}
	if !addressRegex.MatchString(addr) && !isValidMinimaAddress(addr) {
		return fmt.Errorf("invalid address format: %s", truncate(addr, 32))
	}
	return nil
}

func ValidateTxID(txID string) error {
	if txID == "" {
		return fmt.Errorf("transaction id cannot be empty")
	}
	if len(txID) > 256 {
		return fmt.Errorf("transaction id too long: max 256 chars")
	}
	if !txIDRegex.MatchString(txID) && !isValidHexID(txID) {
		return fmt.Errorf("invalid transaction id format")
	}
	return nil
}

func ValidateAmount(amount string) error {
	if amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}
	n, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: must be numeric")
	}
	if n < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	if n > 1e18 {
		return fmt.Errorf("amount exceeds maximum (1e18)")
	}
	return nil
}

func ValidateScript(script string) error {
	if script == "" {
		return fmt.Errorf("contract script cannot be empty")
	}
	if utf8.RuneCountInString(script) > 65536 {
		return fmt.Errorf("contract script too long: max 65536 chars")
	}
	lower := strings.ToLower(script)
	dangerous := []string{"runtime.exec", "processbuilder", "system.exit", "delete(", "drop table"}
	for _, d := range dangerous {
		if strings.Contains(lower, d) {
			return fmt.Errorf("contract script contains forbidden pattern")
		}
	}
	return nil
}

func ValidateContractID(id string) error {
	if id == "" {
		return fmt.Errorf("contract id cannot be empty")
	}
	if !contractIDRegex.MatchString(id) {
		return fmt.Errorf("invalid contract id format")
	}
	return nil
}

func ValidateFunctionName(name string) error {
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if !functionRegex.MatchString(name) {
		return fmt.Errorf("invalid function name format: must match %s", functionRegex.String())
	}
	return nil
}

func ValidateTokenName(name string) error {
	if name == "" {
		return fmt.Errorf("token name cannot be empty")
	}
	if !tokenNameRegex.MatchString(name) {
		return fmt.Errorf("invalid token name: must be 2-32 chars, alphanumeric/underscore/hyphen/space")
	}
	return nil
}

func ValidateTokenID(id string) error {
	if id == "" {
		return fmt.Errorf("token id cannot be empty")
	}
	if !tokenIDRegex.MatchString(id) {
		return fmt.Errorf("invalid token id format")
	}
	return nil
}

func ValidatePrivateKey(key string) error {
	if key == "" {
		return fmt.Errorf("private key cannot be empty")
	}
	if len(key) < 32 {
		return fmt.Errorf("private key too short")
	}
	if len(key) > 512 {
		return fmt.Errorf("private key too long")
	}
	return nil
}

func ValidatePublicKey(key string) error {
	if key == "" {
		return fmt.Errorf("public key cannot be empty")
	}
	if len(key) < 32 {
		return fmt.Errorf("public key too short")
	}
	if len(key) > 512 {
		return fmt.Errorf("public key too long")
	}
	return nil
}

func ValidateMessage(msg string) error {
	if utf8.RuneCountInString(msg) > 65536 {
		return fmt.Errorf("message too long: max 65536 chars")
	}
	return nil
}

func ValidateMiniDAPPID(id string) error {
	if id == "" {
		return fmt.Errorf("minidapp id cannot be empty")
	}
	if !miniDAPPIDRegex.MatchString(id) {
		return fmt.Errorf("invalid minidapp id format")
	}
	return nil
}

func isValidMinimaAddress(addr string) bool {
	if len(addr) < 10 || len(addr) > 128 {
		return false
	}
	return strings.HasPrefix(addr, "Mx") || strings.HasPrefix(addr, "0x") || hexRegex.MatchString(addr)
}

func isValidHexID(id string) bool {
	return hexRegex.MatchString(id) || regexp.MustCompile(`^[0-9a-fA-F]{64,128}$`).MatchString(id)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(val, 10)
	case int:
		return strconv.Itoa(val)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

func toInt(v int64) int {
	return int(v)
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false
		}
		return b
	case float64:
		return val != 0
	case int64:
		return val != 0
	case int:
		return val != 0
	default:
		return false
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
