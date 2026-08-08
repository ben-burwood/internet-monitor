package database

import "testing"

func TestServerSelectionRoundTrip(t *testing.T) {
	setupTestDB(t)

	// Unset defaults to Automatic (nil ID).
	sel, err := GetServerSelection()
	if err != nil {
		t.Fatalf("GetServerSelection (unset): %v", err)
	}
	if sel.ID != nil {
		t.Errorf("unset selection ID = %v, want nil (automatic)", *sel.ID)
	}

	// Pin a server.
	id := 1234
	if err := SetServerSelection(ServerSelection{ID: &id, Name: "Test Server", Location: "London"}); err != nil {
		t.Fatalf("SetServerSelection: %v", err)
	}
	sel, err = GetServerSelection()
	if err != nil {
		t.Fatalf("GetServerSelection (pinned): %v", err)
	}
	if sel.ID == nil || *sel.ID != id || sel.Name != "Test Server" || sel.Location != "London" {
		t.Errorf("pinned selection = %+v", sel)
	}

	// Overwrite back to Automatic.
	if err := SetServerSelection(ServerSelection{}); err != nil {
		t.Fatalf("SetServerSelection (auto): %v", err)
	}
	sel, err = GetServerSelection()
	if err != nil {
		t.Fatalf("GetServerSelection (auto): %v", err)
	}
	if sel.ID != nil {
		t.Errorf("selection ID after reset = %v, want nil", *sel.ID)
	}
}
