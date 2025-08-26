package testhelpers

import (
	"reflect"
	"strings"
	"testing"
	"time"
	
	"github.com/pluqqy/pluqqy-cli/pkg/models"
)

// AssertComponentEqual checks if two component items are equal
func AssertComponentEqual(t *testing.T, expected, actual ComponentItem) {
	t.Helper()
	
	if expected.Name != actual.Name {
		t.Errorf("Component name mismatch: expected %q, got %q", expected.Name, actual.Name)
	}
	if expected.Path != actual.Path {
		t.Errorf("Component path mismatch: expected %q, got %q", expected.Path, actual.Path)
	}
	if expected.CompType != actual.CompType {
		t.Errorf("Component type mismatch: expected %q, got %q", expected.CompType, actual.CompType)
	}
	if expected.TokenCount != actual.TokenCount {
		t.Errorf("Component token count mismatch: expected %d, got %d", expected.TokenCount, actual.TokenCount)
	}
	if expected.UsageCount != actual.UsageCount {
		t.Errorf("Component usage count mismatch: expected %d, got %d", expected.UsageCount, actual.UsageCount)
	}
	if expected.IsArchived != actual.IsArchived {
		t.Errorf("Component archived status mismatch: expected %v, got %v", expected.IsArchived, actual.IsArchived)
	}
	if !reflect.DeepEqual(expected.Tags, actual.Tags) {
		t.Errorf("Component tags mismatch: expected %v, got %v", expected.Tags, actual.Tags)
	}
}


// AssertPipelineEqual checks if two pipeline models are equal
func AssertPipelineEqual(t *testing.T, expected, actual *models.Pipeline) {
	t.Helper()
	
	if expected.Name != actual.Name {
		t.Errorf("Pipeline name mismatch: expected %q, got %q", expected.Name, actual.Name)
	}
	if expected.Path != actual.Path {
		t.Errorf("Pipeline path mismatch: expected %q, got %q", expected.Path, actual.Path)
	}
	if !reflect.DeepEqual(expected.Tags, actual.Tags) {
		t.Errorf("Pipeline tags mismatch: expected %v, got %v", expected.Tags, actual.Tags)
	}
	if !reflect.DeepEqual(expected.Components, actual.Components) {
		t.Errorf("Pipeline components mismatch: expected %v, got %v", expected.Components, actual.Components)
	}
}

// AssertPipelineItemEqual checks if two pipeline items are equal
func AssertPipelineItemEqual(t *testing.T, expected, actual PipelineItem) {
	t.Helper()
	
	if expected.Name != actual.Name {
		t.Errorf("Pipeline name mismatch: expected %q, got %q", expected.Name, actual.Name)
	}
	if expected.Path != actual.Path {
		t.Errorf("Pipeline path mismatch: expected %q, got %q", expected.Path, actual.Path)
	}
	if expected.TokenCount != actual.TokenCount {
		t.Errorf("Pipeline token count mismatch: expected %d, got %d", expected.TokenCount, actual.TokenCount)
	}
	if expected.IsArchived != actual.IsArchived {
		t.Errorf("Pipeline archived status mismatch: expected %v, got %v", expected.IsArchived, actual.IsArchived)
	}
	if !reflect.DeepEqual(expected.Tags, actual.Tags) {
		t.Errorf("Pipeline tags mismatch: expected %v, got %v", expected.Tags, actual.Tags)
	}
}


// AssertViewContains checks if a view contains expected text
func AssertViewContains(t *testing.T, view, expected string) {
	t.Helper()
	
	if !strings.Contains(view, expected) {
		t.Errorf("View does not contain expected text: %q\nView:\n%s", expected, view)
	}
}

// AssertViewNotContains checks if a view does not contain certain text
func AssertViewNotContains(t *testing.T, view, unexpected string) {
	t.Helper()
	
	if strings.Contains(view, unexpected) {
		t.Errorf("View unexpectedly contains text: %q\nView:\n%s", unexpected, view)
	}
}

// AssertNoError checks that an error is nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertError checks that an error is not nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}
}

// AssertErrorContains checks that an error contains expected text
func AssertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()
	
	if err == nil {
		t.Fatal("Expected error, but got nil")
	}
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Error message does not contain expected text: %q\nGot: %v", expected, err)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Values not equal:\nExpected: %v\nActual:   %v", expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	
	if reflect.DeepEqual(expected, actual) {
		t.Errorf("Values should not be equal: %v", expected)
	}
}

// AssertTrue checks if a condition is true
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	
	if !condition {
		t.Error(message)
	}
}

// AssertFalse checks if a condition is false
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	
	if condition {
		t.Error(message)
	}
}

// AssertSliceEqual checks if two slices are equal
func AssertSliceEqual[T comparable](t *testing.T, expected, actual []T) {
	t.Helper()
	
	if len(expected) != len(actual) {
		t.Errorf("Slice length mismatch: expected %d, got %d", len(expected), len(actual))
		return
	}
	
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("Slice element mismatch at index %d: expected %v, got %v", i, expected[i], actual[i])
		}
	}
}

// AssertMapEqual checks if two maps are equal
func AssertMapEqual[K comparable, V comparable](t *testing.T, expected, actual map[K]V) {
	t.Helper()
	
	if len(expected) != len(actual) {
		t.Errorf("Map length mismatch: expected %d, got %d", len(expected), len(actual))
		return
	}
	
	for k, v := range expected {
		actualV, ok := actual[k]
		if !ok {
			t.Errorf("Key %v not found in actual map", k)
			continue
		}
		if v != actualV {
			t.Errorf("Value mismatch for key %v: expected %v, got %v", k, v, actualV)
		}
	}
}

// AssertWithinDuration checks if two times are within a given duration
func AssertWithinDuration(t *testing.T, expected, actual time.Time, delta time.Duration) {
	t.Helper()
	
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	if diff > delta {
		t.Errorf("Times not within duration: expected %v, got %v (delta: %v)", expected, actual, diff)
	}
}

// WaitForCondition waits for a condition with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
	t.Helper()
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Condition not met within timeout: %s", msg)
}