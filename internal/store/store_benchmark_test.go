package store

import (
	"fmt"
	"testing"
)

func BenchmarkStoreSize(b *testing.B) {
	st := New()
	for i := range 100_000 {
		if err := st.Set(fmt.Sprintf("key:%d", i), "value"); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_ = st.Size()
	}
}

func BenchmarkStoreSetGet(b *testing.B) {
	st := New()
	b.ResetTimer()
	for i := range b.N {
		key := fmt.Sprintf("key:%d", i)
		if err := st.Set(key, "value"); err != nil {
			b.Fatal(err)
		}
		if _, ok := st.Get(key); !ok {
			b.Fatal("missing key")
		}
	}
}
