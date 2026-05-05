package command

import (
	"context"
	"strconv"
	"testing"

	"github.com/KillDarkness/gokv/internal/store"
)

func BenchmarkRegistrySet(b *testing.B) {
	registry := NewDefaultRegistry(nil)
	st := store.New()
	ctx := context.Background()

	b.ReportAllocs()
	for i := range b.N {
		registry.Dispatch(ctx, st, nil, []string{"SET", "key:" + strconv.Itoa(i), "value"})
	}
}

func BenchmarkRegistryGet(b *testing.B) {
	registry := NewDefaultRegistry(nil)
	st := store.New()
	ctx := context.Background()
	registry.Dispatch(ctx, st, nil, []string{"SET", "key", "value"})

	b.ReportAllocs()
	for b.Loop() {
		registry.Dispatch(ctx, st, nil, []string{"GET", "key"})
	}
}

func BenchmarkRegistryGetParallel(b *testing.B) {
	registry := NewDefaultRegistry(nil)
	st := store.New()
	ctx := context.Background()
	registry.Dispatch(ctx, st, nil, []string{"SET", "key", "value"})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			registry.Dispatch(ctx, st, nil, []string{"GET", "key"})
		}
	})
}
