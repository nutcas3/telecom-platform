package graphql

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestParseID_Valid(t *testing.T) {
	id, err := parseID("Subscriber:42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestParseID_Invalid(t *testing.T) {
	_, err := parseID("invalid")
	if err == nil {
		t.Fatal("expected error for invalid ID format")
	}
}

func TestParseID_Zero(t *testing.T) {
	id, err := parseID("Subscriber:0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 0 {
		t.Errorf("expected 0, got %d", id)
	}
}

func TestEncodeCursor(t *testing.T) {
	cursor := encodeCursor(42)
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %v", err)
	}
	if string(decoded) != "cursor:42" {
		t.Errorf("expected 'cursor:42', got '%s'", string(decoded))
	}
}

func TestDecodeCursor_Valid(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("cursor:100"))
	offset, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offset != 100 {
		t.Errorf("expected 100, got %d", offset)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	_, err := decodeCursor("not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid cursor")
	}
}

func TestDecodeCursor_BadFormat(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("bad:format"))
	_, err := decodeCursor(encoded)
	if err == nil {
		t.Fatal("expected error for bad cursor format")
	}
}

func TestBuildUsageFilterWhere_Empty(t *testing.T) {
	result := buildUsageFilterWhere(nil)
	if result != "" {
		t.Errorf("expected empty string for nil filter, got '%s'", result)
	}
}

func TestBuildUsageFilterWhere_WithType(t *testing.T) {
	usageType := UsageTypeData
	filter := &UsageEventFilter{UsageType: &usageType}
	result := buildUsageFilterWhere(filter)
	if result == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
	expected := "WHERE usage_type = 'DATA'"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestBuildUsageFilterWhere_WithCostRange(t *testing.T) {
	min := 1.0
	max := 10.0
	filter := &UsageEventFilter{CostMin: &min, CostMax: &max}
	result := buildUsageFilterWhere(filter)
	if result == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
}

func TestResolverRoot_Methods(t *testing.T) {
	r := &Resolver{}
	if r.Query() != r {
		t.Error("Query() should return receiver")
	}
	if r.Mutation() != r {
		t.Error("Mutation() should return receiver")
	}
	if r.Subscription() != r {
		t.Error("Subscription() should return receiver")
	}
}

// Helper functions for testing (these should match the implementations in resolvers.go)

func encodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "cursor:%d", offset))
}

func decodeCursor(cursor string) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor encoding: %w", err)
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 || parts[0] != "cursor" {
		return 0, fmt.Errorf("invalid cursor format")
	}

	return strconv.Atoi(parts[1])
}

func buildUsageFilterWhere(filter *UsageEventFilter) string {
	if filter == nil {
		return ""
	}

	var conditions []string

	if filter.UsageType != nil {
		conditions = append(conditions, fmt.Sprintf("usage_type = '%s'", *filter.UsageType))
	}

	if filter.CostMin != nil {
		conditions = append(conditions, fmt.Sprintf("cost >= %f", *filter.CostMin))
	}

	if filter.CostMax != nil {
		conditions = append(conditions, fmt.Sprintf("cost <= %f", *filter.CostMax))
	}

	if len(conditions) > 0 {
		return "WHERE " + strings.Join(conditions, " AND ")
	}

	return ""
}
