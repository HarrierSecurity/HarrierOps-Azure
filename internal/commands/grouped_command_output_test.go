package commands

import (
	"context"
	"testing"
)

func TestRunGroupedCommandOutputReusesCommandResult(t *testing.T) {
	group := newCommandOutputGroup(chainsFanoutLimit)
	calls := 0
	handler := func(context.Context, Request) (any, error) {
		calls++
		return "shared-output", nil
	}

	first := runGroupedCommandOutput[string](group, context.Background(), Request{}, handler, "shared")
	second := runGroupedCommandOutput[string](group, context.Background(), Request{}, handler, "shared")

	firstValue, err := first.wait()
	if err != nil {
		t.Fatalf("first wait failed: %v", err)
	}
	secondValue, err := second.wait()
	if err != nil {
		t.Fatalf("second wait failed: %v", err)
	}
	if firstValue != "shared-output" || secondValue != "shared-output" {
		t.Fatalf("unexpected values: first=%q second=%q", firstValue, secondValue)
	}
	if calls != 1 {
		t.Fatalf("expected one backing call, got %d", calls)
	}
}
