package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/streamingfast/dstore"
	"github.com/streamingfast/substreams/block"
	pbsubstreams "github.com/streamingfast/substreams/pb/sf/substreams/v1"
	"github.com/streamingfast/substreams/state"
	"github.com/stretchr/testify/require"
)

type NotifierFunc func()

func (n NotifierFunc) Notify(builder string, blockNum uint64) {
	n()
}

func TestSquash(t *testing.T) {
	ctx := context.Background()

	writeCount := 0
	var infoBytes []byte

	store := dstore.NewMockStore(nil)
	store.WriteObjectFunc = func(ctx context.Context, base string, f io.Reader) error {
		if base == state.InfoFileName() {
			infoBytes, _ = io.ReadAll(f)
			return nil
		}
		writeCount++
		return nil
	}

	store.OpenObjectFunc = func(ctx context.Context, name string) (out io.ReadCloser, err error) {
		if name == state.InfoFileName() {
			if infoBytes == nil {
				return nil, dstore.ErrNotFound
			}
			return io.NopCloser(bytes.NewReader(infoBytes)), nil
		}
		return nil, fmt.Errorf("no")
	}

	squashable := &Squashable{
		builder: testStateBuilder(store),
		ranges:  []*block.Range{},
	}

	notificationsSent := 0
	notifierFunc := NotifierFunc(func() {
		notificationsSent++
	})

	err := squash(ctx, squashable, &block.Range{StartBlock: 20_000, ExclusiveEndBlock: 30_000}, notifierFunc)
	require.Nil(t, err)
	require.Equal(t, 0, writeCount)

	err = squash(ctx, squashable, &block.Range{StartBlock: 70_000, ExclusiveEndBlock: 80_000}, notifierFunc)
	require.Nil(t, err)
	require.Equal(t, 0, writeCount)

	err = squash(ctx, squashable, &block.Range{StartBlock: 10_000, ExclusiveEndBlock: 20_000}, notifierFunc)
	require.Nil(t, err)

	require.Equal(t, 2, writeCount) //both [10_000,20_000) and [20_000,30_000) will be merged and written
	require.Equal(t, 2, notificationsSent)
}

func testStateBuilder(store dstore.Store) *state.Builder {
	return &state.Builder{
		Name:             "testBuilder",
		SaveInterval:     10_000,
		ModuleStartBlock: 0,
		Store:            store,
		ModuleHash:       "abc",
		KV:               map[string][]byte{},
		PartialMode:      false,
		BlockRange: &block.Range{
			StartBlock:        0,
			ExclusiveEndBlock: 10_000,
		},
		UpdatePolicy: pbsubstreams.Module_KindStore_UPDATE_POLICY_SET,
		ValueType:    state.OutputValueTypeString,
	}
}