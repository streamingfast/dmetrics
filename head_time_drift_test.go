package dmetrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHeadTimeDrift_SetBlockTimeForward(t *testing.T) {
	h := NewSet().NewHeadTimeDrift("test-set-block-time-forward")

	head := time.Now().Add(-time.Minute)
	h.SetBlockTimeForward(head)
	require.Equal(t, head.UnixNano(), h.lastBlockTimeNanos.Load())

	h.SetBlockTimeForward(head.Add(-10 * time.Minute))
	require.Equal(t, head.UnixNano(), h.lastBlockTimeNanos.Load(), "an older block time must be ignored")

	h.SetBlockTimeForward(head)
	require.Equal(t, head.UnixNano(), h.lastBlockTimeNanos.Load(), "the same block time must be ignored")

	newer := head.Add(30 * time.Second)
	h.SetBlockTimeForward(newer)
	require.Equal(t, newer.UnixNano(), h.lastBlockTimeNanos.Load())
}
