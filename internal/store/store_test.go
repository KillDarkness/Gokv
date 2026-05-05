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

func TestStoreExpireAndTTL(t *testing.T) {
	st := New()

	if _, exists, hasTTL := st.TTL("missing"); exists || hasTTL {
		t.Fatal("TTL() reported metadata for missing key")
	}

	st.Set("name", "kill")
	if _, exists, hasTTL := st.TTL("name"); !exists || hasTTL {
		t.Fatalf("TTL() exists = %v, hasTTL = %v; want true, false", exists, hasTTL)
	}

	if ok := st.Expire("name", 10_000_000_000); !ok {
		t.Fatal("Expire() = false; want true")
	}
	if ttl, exists, hasTTL := st.TTL("name"); !exists || !hasTTL || ttl <= 0 {
		t.Fatalf("TTL() = %v, %v, %v; want positive, true, true", ttl, exists, hasTTL)
	}

	if ok := st.Expire("name", 0); !ok {
		t.Fatal("Expire(0) = false; want true")
	}
	if _, ok := st.Get("name"); ok {
		t.Fatal("Get() found expired key")
	}
}
