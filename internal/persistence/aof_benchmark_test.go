package persistence

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkAOFAppendSyncAlways(b *testing.B) {
	aof, err := NewAOF(true, filepath.Join(b.TempDir(), "appendonly.aof"), string(FsyncAlways))
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		if err := aof.Append(ctx, []string{"SET", "key", strconv.Itoa(i)}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAOFAppendAsyncEverySec(b *testing.B) {
	aof, err := NewAOF(true, filepath.Join(b.TempDir(), "appendonly.aof"), string(FsyncEverySec))
	if err != nil {
		b.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := aof.StartWriter(ctx, 8192)

	b.ResetTimer()
	for i := range b.N {
		if err := aof.Append(ctx, []string{"SET", "key", strconv.Itoa(i)}); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	cancel()
	<-done
}
