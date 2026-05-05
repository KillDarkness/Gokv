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

func TestStoreIncrement(t *testing.T) {
	st := New()

	if got, err := st.Increment("counter", 1); err != nil || got != 1 {
		t.Fatalf("Increment() = %d, %v; want 1, nil", got, err)
	}
	if got, err := st.Increment("counter", -1); err != nil || got != 0 {
		t.Fatalf("Increment() = %d, %v; want 0, nil", got, err)
	}

	st.Set("bad", "not-number")
	if _, err := st.Increment("bad", 1); err == nil {
		t.Fatal("Increment() error = nil; want error")
	}
}

func TestStoreMSetMGet(t *testing.T) {
	st := New()
	st.MSet(map[string]string{"name": "kill", "lang": "go"})

	got := st.MGet("name", "missing", "lang")
	if len(got) != 3 {
		t.Fatalf("MGet() length = %d; want 3", len(got))
	}
	if !got[0].OK || got[0].Value != "kill" {
		t.Fatalf("MGet(name) = %#v; want kill", got[0])
	}
	if got[1].OK {
		t.Fatalf("MGet(missing) = %#v; want missing", got[1])
	}
	if !got[2].OK || got[2].Value != "go" {
		t.Fatalf("MGet(lang) = %#v; want go", got[2])
	}
}

func TestStoreFlushDBAndSize(t *testing.T) {
	st := New()
	st.Set("name", "kill")
	st.Set("lang", "go")

	if got := st.Size(); got != 2 {
		t.Fatalf("Size() = %d; want 2", got)
	}
	st.FlushDB()
	if got := st.Size(); got != 0 {
		t.Fatalf("Size() = %d; want 0", got)
	}
}
