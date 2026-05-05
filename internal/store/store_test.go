package store

import (
	"testing"
	"time"
)

func TestStoreSetGetDeleteExists(t *testing.T) {
	st := New()

	if err := st.Set("name", "kill"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
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

	if err := st.Set("name", "kill"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
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

	if err := st.Set("bad", "not-number"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if _, err := st.Increment("bad", 1); err == nil {
		t.Fatal("Increment() error = nil; want error")
	}
}

func TestStoreMSetMGet(t *testing.T) {
	st := New()
	if err := st.MSet(map[string]string{"name": "kill", "lang": "go"}); err != nil {
		t.Fatalf("MSet() error = %v", err)
	}

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
	if err := st.Set("name", "kill"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := st.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if got := st.Size(); got != 2 {
		t.Fatalf("Size() = %d; want 2", got)
	}
	st.FlushDB()
	if got := st.Size(); got != 0 {
		t.Fatalf("Size() = %d; want 0", got)
	}
}

func TestStoreSnapshotAndRestore(t *testing.T) {
	st := New()
	if err := st.Set("name", "kill"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	restored := New()
	restored.Restore(st.Snapshot())

	if got, ok := restored.Get("name"); !ok || got != "kill" {
		t.Fatalf("Get() = %q, %v; want kill, true", got, ok)
	}
}

func TestStoreRuleAppliesTTLByPrefix(t *testing.T) {
	st := New()
	st.SetRule("session:", time.Minute)
	if err := st.Set("session:abc", "token"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if ttl, exists, hasTTL := st.TTL("session:abc"); !exists || !hasTTL || ttl <= 0 {
		t.Fatalf("TTL() = %v, %v, %v; want positive, true, true", ttl, exists, hasTTL)
	}
	if err := st.Set("plain", "value"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if _, exists, hasTTL := st.TTL("plain"); !exists || hasTTL {
		t.Fatalf("TTL(plain) exists = %v, hasTTL = %v; want true, false", exists, hasTTL)
	}
}

func TestStoreEvictsLRUKey(t *testing.T) {
	st := NewWithOptions(Options{MaxKeys: 2, EvictionPolicy: AllKeysLRU})
	if err := st.Set("old", "1"); err != nil {
		t.Fatalf("Set(old) error = %v", err)
	}
	if err := st.Set("new", "2"); err != nil {
		t.Fatalf("Set(new) error = %v", err)
	}
	if err := st.Set("latest", "3"); err != nil {
		t.Fatalf("Set(latest) error = %v", err)
	}
	if _, ok := st.Get("old"); ok {
		t.Fatal("Get(old) found evicted key")
	}
	if got := st.Size(); got != 2 {
		t.Fatalf("Size() = %d; want 2", got)
	}
}

func TestStoreNoEvictionReturnsError(t *testing.T) {
	st := NewWithOptions(Options{MaxKeys: 1, EvictionPolicy: NoEviction})
	if err := st.Set("one", "1"); err != nil {
		t.Fatalf("Set(one) error = %v", err)
	}
	if err := st.Set("two", "2"); err == nil {
		t.Fatal("Set(two) error = nil; want error")
	}
}
