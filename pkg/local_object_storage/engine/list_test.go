package engine

import (
	"errors"
	"os"
	"sort"
	"testing"

	"github.com/nspcc-dev/neofs-node/pkg/core/object"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestListWithCursor(t *testing.T) {
	s1 := testNewShard(t, 1)
	s2 := testNewShard(t, 2)
	e := testNewEngineWithShards(s1, s2)

	t.Cleanup(func() {
		e.Close()
		os.RemoveAll(t.Name())
	})

	const total = 20

	expected := make([]oid.Address, 0, total)
	got := make([]oid.Address, 0, total)

	for i := 0; i < total; i++ {
		containerID := cidtest.ID()
		obj := generateObjectWithCID(t, containerID)

		var prm PutPrm
		prm.WithObject(obj)

		_, err := e.Put(prm)
		require.NoError(t, err)
		expected = append(expected, object.AddressOf(obj))
	}

	expected = sortAddresses(expected)

	var prm ListWithCursorPrm
	prm.WithCount(1)

	res, err := e.ListWithCursor(prm)
	require.NoError(t, err)
	require.NotEmpty(t, res.AddressList())
	got = append(got, res.AddressList()...)

	for i := 0; i < total-1; i++ {
		prm.WithCursor(res.Cursor())

		res, err = e.ListWithCursor(prm)
		if errors.Is(err, ErrEndOfListing) {
			break
		}
		got = append(got, res.AddressList()...)
	}

	prm.WithCursor(res.Cursor())

	_, err = e.ListWithCursor(prm)
	require.ErrorIs(t, err, ErrEndOfListing)

	got = sortAddresses(got)
	require.Equal(t, expected, got)
}

func sortAddresses(addr []oid.Address) []oid.Address {
	sort.Slice(addr, func(i, j int) bool {
		return addr[i].EncodeToString() < addr[j].EncodeToString()
	})
	return addr
}
