package testutil

import "testing"

func TestMockDataStore_SetCurrentSuite_UpdatesCurrentSuite(t *testing.T) {
	mock := &MockDataStore{}

	if err := mock.SetCurrentSuite("suite-a"); err != nil {
		t.Fatalf("SetCurrentSuite returned error: %v", err)
	}

	if mock.CurrentSuite != "suite-a" {
		t.Fatalf("expected CurrentSuite to be updated, got %q", mock.CurrentSuite)
	}

	if got := mock.GetCurrentSuiteName(); got != "suite-a" {
		t.Fatalf("expected GetCurrentSuiteName to return updated suite, got %q", got)
	}
}
