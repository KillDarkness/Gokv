package store

import "testing"

func TestStoreSetGetDeleteExists(t *testing.T) {
	st := New()

	st.Set("name", "kill")
	if got, ok := st.Get("name"); !ok || got != "kill" {
		t.Fatalf("Get() = %q, %v; want kill, true", got, ok)
	}

	if got := st.Exists("name", "missing"); got != 1 {
		t.Fatalf("Exists() = %d; want 1", got)
	}

	if got := st.Delete("name", "missing"); got != 1 {
		t.Fatalf("Delete() = %d; want 1", got)
	}

	if _, ok := st.Get("name"); ok {
		t.Fatal("Get() found deleted key")
	}
}
