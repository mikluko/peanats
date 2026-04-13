package prom

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xmock/peanatsmock"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMiddleware_AckCounting(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*peanatsmock.MsgJetstream)
		ackOperation    func(context.Context, peanats.Ackable) error
		expectedAckType string
		expectError     bool
	}{
		{
			name: "Ack operation increments ack counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				// Subject is called multiple times for metrics - use a more flexible expectation
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().Ack(context.Background()).Return(nil).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.Ack(ctx)
			},
			expectedAckType: "ack",
			expectError:     false,
		},
		{
			name: "Nak operation increments nak counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().Nak(context.Background()).Return(nil).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.Nak(ctx)
			},
			expectedAckType: "nak",
			expectError:     false,
		},
		{
			name: "Term operation increments term counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().Term(context.Background()).Return(nil).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.Term(ctx)
			},
			expectedAckType: "term",
			expectError:     false,
		},
		{
			name: "TermWithReason operation increments term counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().TermWithReason(context.Background(), "test reason").Return(nil).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.TermWithReason(ctx, "test reason")
			},
			expectedAckType: "term",
			expectError:     false,
		},
		{
			name: "InProgress operation increments in_progress counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().InProgress(context.Background()).Return(nil).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.InProgress(ctx)
			},
			expectedAckType: "in_progress",
			expectError:     false,
		},
		{
			name: "NackWithDelay operation increments nak counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().NackWithDelay(context.Background(), 5*time.Second).Return(nil).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.NackWithDelay(ctx, 5*time.Second)
			},
			expectedAckType: "nak",
			expectError:     false,
		},
		{
			name: "Ack operation with error still increments counter",
			setupMock: func(mock *peanatsmock.MsgJetstream) {
				mock.EXPECT().Subject().Return("test.subject").Maybe()
				mock.EXPECT().Data().Return([]byte("test data")).Maybe()
				mock.EXPECT().Header().Return(peanats.Header{}).Maybe()
				mock.EXPECT().Ack(context.Background()).Return(errors.New("ack error")).Once()
			},
			ackOperation: func(ctx context.Context, ackable peanats.Ackable) error {
				return ackable.Ack(ctx)
			},
			expectedAckType: "ack",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new registry for each test
			registry := prometheus.NewRegistry()

			// Create mocks
			mockMsg := peanatsmock.NewMsgJetstream(t)
			tt.setupMock(mockMsg)

			// Create middleware with custom registry
			middleware := Middleware(MiddlewareRegisterer(registry))

			// Create a simple handler that performs the ack operation
			handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				if ackable, ok := msg.(peanats.Ackable); ok {
					err := tt.ackOperation(ctx, ackable)
					if tt.expectError {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
				return nil
			})

			// Wrap handler with middleware
			wrappedHandler := middleware(handler)

			// Execute handler
			err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
			require.NoError(t, err)

			// Verify metrics using GatherAndCompare
			expected := `
# HELP peanats_acked_total Total number of message acknowledgments
# TYPE peanats_acked_total counter
peanats_acked_total{subject="test.subject",type="` + tt.expectedAckType + `"} 1
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="test.subject"} 1
`
			compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_acked_total", "peanats_processed_total")
			assert.NoError(t, compareErr, "Metrics should match expected values")
		})
	}
}

func TestPrometheusMiddleware_NonAckableMessage(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create a non-ackable message mock
	mockMsg := peanatsmock.NewMsg(t)
	mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("test data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Create middleware with custom registry
	middleware := Middleware(MiddlewareRegisterer(registry))

	// Create a simple handler
	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		// Non-ackable message shouldn't have ack methods
		_, isAckable := msg.(peanats.Ackable)
		assert.False(t, isAckable, "Non-ackable message should not implement Ackable")
		return nil
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Execute handler
	err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
	require.NoError(t, err)

	// Verify only processing counter was incremented, no ack counters
	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="test.subject"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr, "Only processed counter should be incremented for non-ackable messages")
}

func TestPrometheusMiddleware_HandlerError(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create an ackable message mock
	mockMsg := peanatsmock.NewMsgJetstream(t)
	mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("test data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Create middleware with custom registry
	middleware := Middleware(MiddlewareRegisterer(registry))

	// Create a handler that returns an error
	expectedError := errors.New("handler error")
	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		return expectedError
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Execute handler
	err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
	assert.Equal(t, expectedError, err)

	// Verify processing counter was incremented with error status
	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="error",subject="test.subject"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr, "Processed counter should be incremented with error status")
}

func TestPrometheusMiddleware_InFlightGauge(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create a message mock
	mockMsg := peanatsmock.NewMsg(t)
	mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("test data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Create middleware with custom registry
	middleware := Middleware(MiddlewareRegisterer(registry))

	// Track in-flight gauge during handler execution
	var inFlightDuringExecution int
	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		// Capture in-flight gauge count during execution
		inFlightDuringExecution = testutil.CollectAndCount(registry, "peanats_in_flight")
		return nil
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Execute handler
	err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
	require.NoError(t, err)

	// Verify in-flight was incremented during execution
	assert.Equal(t, 1, inFlightDuringExecution, "In-flight should be 1 during execution")

	// Verify in-flight was decremented after execution (should be 0)
	expected := `
# HELP peanats_in_flight Number of messages currently being processed
# TYPE peanats_in_flight gauge
peanats_in_flight{subject="test.subject"} 0
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="test.subject"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_in_flight", "peanats_processed_total")
	assert.NoError(t, compareErr, "In-flight should be 0 after execution")
}

func TestPrometheusMiddleware_MultipleAckOperations(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create an ackable message mock
	mockMsg := peanatsmock.NewMsgJetstream(t)
	mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("test data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()
	mockMsg.EXPECT().InProgress(context.Background()).Return(nil).Times(3)
	mockMsg.EXPECT().Ack(context.Background()).Return(nil).Once()

	// Create middleware with custom registry
	middleware := Middleware(MiddlewareRegisterer(registry))

	// Create a handler that performs multiple ack operations
	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		if ackable, ok := msg.(peanats.Ackable); ok {
			// Simulate multiple InProgress calls followed by final Ack
			err := ackable.InProgress(ctx)
			assert.NoError(t, err)
			err = ackable.InProgress(ctx)
			assert.NoError(t, err)
			err = ackable.InProgress(ctx)
			assert.NoError(t, err)
			err = ackable.Ack(ctx)
			assert.NoError(t, err)
		}
		return nil
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Execute handler
	err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
	require.NoError(t, err)

	// Verify multiple ack operations were tracked correctly
	expected := `
# HELP peanats_acked_total Total number of message acknowledgments
# TYPE peanats_acked_total counter
peanats_acked_total{subject="test.subject",type="ack"} 1
peanats_acked_total{subject="test.subject",type="in_progress"} 3
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="test.subject"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_acked_total", "peanats_processed_total")
	assert.NoError(t, compareErr, "Multiple ack operations should be tracked correctly")
}

func TestPrometheusMiddleware_CustomNamespaceSubsystem(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create a message mock
	mockMsg := peanatsmock.NewMsg(t)
	mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("test data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Create middleware with custom namespace and subsystem
	middleware := Middleware(
		MiddlewareNamespace("myapp"),
		MiddlewareSubsystem("messaging"),
		MiddlewareRegisterer(registry),
	)

	// Create a simple handler
	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		return nil
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Execute handler
	err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
	require.NoError(t, err)

	// Verify metric names include custom namespace and subsystem
	families, err := registry.Gather()
	require.NoError(t, err)

	expectedNames := map[string]bool{
		"myapp_messaging_processed_total": false,
		"myapp_messaging_latency_seconds": false,
		"myapp_messaging_in_flight":       false,
	}

	for _, family := range families {
		if _, exists := expectedNames[*family.Name]; exists {
			expectedNames[*family.Name] = true
		}
	}

	for name, found := range expectedNames {
		assert.True(t, found, "Expected metric %s to be registered", name)
	}
}

func TestPrometheusMiddleware_MetadatablePassthrough(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Create a JetStream message mock that implements both Ackable and Metadatable
	mockMsg := peanatsmock.NewMsgJetstream(t)
	mockMsg.EXPECT().Subject().Return("test.subject").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("test data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Set up mock metadata
	expectedMetadata := &jetstream.MsgMetadata{
		Sequence: jetstream.SequencePair{
			Stream:   42,
			Consumer: 10,
		},
		NumDelivered: 1,
		NumPending:   5,
		Stream:       "test-stream",
		Consumer:     "test-consumer",
		Domain:       "test-domain",
	}
	mockMsg.EXPECT().Metadata().Return(expectedMetadata, nil).Once()

	// Create middleware with custom registry
	middleware := Middleware(MiddlewareRegisterer(registry))

	// Create a handler that verifies Metadatable interface works through the wrapper
	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		// Verify the message still implements Ackable after wrapping
		ackable, isAckable := msg.(peanats.Ackable)
		assert.True(t, isAckable, "Wrapped message should still implement Ackable")
		assert.NotNil(t, ackable, "Ackable interface should not be nil")

		// Verify the message implements Metadatable after wrapping
		metadatable, isMetadatable := msg.(peanats.Metadatable)
		assert.True(t, isMetadatable, "Wrapped message should implement Metadatable")
		assert.NotNil(t, metadatable, "Metadatable interface should not be nil")

		// Verify we can call Metadata() and get the expected result
		metadata, err := metadatable.Metadata()
		assert.NoError(t, err, "Metadata() should not return error")
		assert.Equal(t, expectedMetadata, metadata, "Metadata should match expected values")

		return nil
	})

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Execute handler
	err := wrappedHandler.HandleMsg(context.Background(), mockMsg)
	require.NoError(t, err)

	// Verify processing metrics are recorded
	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="test.subject"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr, "Processed counter should be incremented")
}

func TestSubjectDepth(t *testing.T) {
	tests := []struct {
		name    string
		depth   int
		subject string
		want    string
	}{
		{"depth 1", 1, "a.b.c.d", "a"},
		{"depth 2", 2, "a.b.c.d", "a.b"},
		{"depth 3", 3, "a.b.c.d", "a.b.c"},
		{"depth equal to tokens", 4, "a.b.c.d", "a.b.c.d"},
		{"depth exceeds tokens", 10, "a.b.c.d", "a.b.c.d"},
		{"single token", 2, "a", "a"},
		{"empty subject", 2, "", ""},
		{"zero depth collapses", 0, "a.b.c", ""},
		{"negative depth collapses", -1, "a.b.c", ""},
		{"real jetstream subject", 5, "up.eu.monitoring.v1beta7.checkexecutionrequested.tenant-42.entity-abc", "up.eu.monitoring.v1beta7.checkexecutionrequested"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubjectDepth(tt.depth)(tt.subject)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubjectConstant(t *testing.T) {
	assert.Equal(t, "fixed", SubjectConstant("fixed")("a.b.c"))
	assert.Equal(t, "", SubjectConstant("")("a.b.c"))
	assert.Equal(t, "", SubjectConstant("")(""))
}

func TestPrometheusMiddleware_SubjectMapper_CollapsesCardinality(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Two messages with different full subjects that share a common prefix.
	msg1 := peanatsmock.NewMsg(t)
	msg1.EXPECT().Subject().Return("up.eu.monitoring.v1.entity-abc").Maybe()
	msg1.EXPECT().Data().Return([]byte("data")).Maybe()
	msg1.EXPECT().Header().Return(peanats.Header{}).Maybe()

	msg2 := peanatsmock.NewMsg(t)
	msg2.EXPECT().Subject().Return("up.eu.monitoring.v1.entity-xyz").Maybe()
	msg2.EXPECT().Data().Return([]byte("data")).Maybe()
	msg2.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Truncate to 4 tokens so both messages collapse to the same label value.
	middleware := Middleware(
		MiddlewareRegisterer(registry),
		MiddlewareSubjectMapper(SubjectDepth(4)),
	)

	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		return nil
	})
	wrapped := middleware(handler)

	require.NoError(t, wrapped.HandleMsg(context.Background(), msg1))
	require.NoError(t, wrapped.HandleMsg(context.Background(), msg2))

	// Both messages should be counted under the truncated subject.
	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="up.eu.monitoring.v1"} 2
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr)
}

func TestPrometheusMiddleware_SubjectMapper_ConstantEmpty(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Two messages with completely unrelated subjects.
	msg1 := peanatsmock.NewMsg(t)
	msg1.EXPECT().Subject().Return("foo.bar").Maybe()
	msg1.EXPECT().Data().Return([]byte("data")).Maybe()
	msg1.EXPECT().Header().Return(peanats.Header{}).Maybe()

	msg2 := peanatsmock.NewMsg(t)
	msg2.EXPECT().Subject().Return("baz.qux.quux").Maybe()
	msg2.EXPECT().Data().Return([]byte("data")).Maybe()
	msg2.EXPECT().Header().Return(peanats.Header{}).Maybe()

	// Collapse every subject to an empty label — effectively disables the
	// subject dimension for cardinality purposes.
	middleware := Middleware(
		MiddlewareRegisterer(registry),
		MiddlewareSubjectMapper(SubjectConstant("")),
	)

	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		return nil
	})
	wrapped := middleware(handler)

	require.NoError(t, wrapped.HandleMsg(context.Background(), msg1))
	require.NoError(t, wrapped.HandleMsg(context.Background(), msg2))

	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject=""} 2
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr)
}

func TestPrometheusMiddleware_DefaultSubjectMapper(t *testing.T) {
	// The default middleware (no MiddlewareSubjectMapper option) should apply
	// SubjectDepth(DefaultSubjectDepth) to bound cardinality out of the box.
	registry := prometheus.NewRegistry()

	// Subject with dynamic tail tokens beyond the default depth.
	mockMsg := peanatsmock.NewMsg(t)
	mockMsg.EXPECT().Subject().Return("orders.v1.created.tenant-42.entity-abc").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	middleware := Middleware(MiddlewareRegisterer(registry))

	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		return nil
	})
	wrapped := middleware(handler)
	require.NoError(t, wrapped.HandleMsg(context.Background(), mockMsg))

	// Expect the subject to be truncated to the first DefaultSubjectDepth tokens.
	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="orders.v1.created"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr)
}

func TestPrometheusMiddleware_ExplicitNilMapperDisablesDefault(t *testing.T) {
	// Passing MiddlewareSubjectMapper(nil) explicitly should disable the
	// default and yield pass-through behavior (raw subject).
	registry := prometheus.NewRegistry()

	mockMsg := peanatsmock.NewMsg(t)
	mockMsg.EXPECT().Subject().Return("orders.v1.created.tenant-42.entity-abc").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()

	middleware := Middleware(
		MiddlewareRegisterer(registry),
		MiddlewareSubjectMapper(nil),
	)

	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		return nil
	})
	wrapped := middleware(handler)
	require.NoError(t, wrapped.HandleMsg(context.Background(), mockMsg))

	expected := `
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="orders.v1.created.tenant-42.entity-abc"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_processed_total")
	assert.NoError(t, compareErr)
}

func TestPrometheusMiddleware_SubjectMapper_AppliedToAckCounters(t *testing.T) {
	// Create a new registry for this test
	registry := prometheus.NewRegistry()

	// Ackable message with a dynamic subject segment.
	mockMsg := peanatsmock.NewMsgJetstream(t)
	mockMsg.EXPECT().Subject().Return("a.b.c.entity-42").Maybe()
	mockMsg.EXPECT().Data().Return([]byte("data")).Maybe()
	mockMsg.EXPECT().Header().Return(peanats.Header{}).Maybe()
	mockMsg.EXPECT().Ack(context.Background()).Return(nil).Once()

	// Depth-2 truncation.
	middleware := Middleware(
		MiddlewareRegisterer(registry),
		MiddlewareSubjectMapper(SubjectDepth(2)),
	)

	handler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
		ackable, ok := msg.(peanats.Ackable)
		require.True(t, ok)
		return ackable.Ack(ctx)
	})
	wrapped := middleware(handler)
	require.NoError(t, wrapped.HandleMsg(context.Background(), mockMsg))

	// Both processed_total and acked_total should carry the truncated subject,
	// not the raw one.
	expected := `
# HELP peanats_acked_total Total number of message acknowledgments
# TYPE peanats_acked_total counter
peanats_acked_total{subject="a.b",type="ack"} 1
# HELP peanats_processed_total Total number of messages processed
# TYPE peanats_processed_total counter
peanats_processed_total{status="success",subject="a.b"} 1
`
	compareErr := testutil.GatherAndCompare(registry, strings.NewReader(expected), "peanats_acked_total", "peanats_processed_total")
	assert.NoError(t, compareErr)
}
