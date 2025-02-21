package jetconsumer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrAck(t *testing.T) {
	err := fmt.Errorf("%w: %s", ErrAck, "test")
	require.NotEqual(t, err, ErrAck)
	require.ErrorIs(t, err, ErrAck)
}
