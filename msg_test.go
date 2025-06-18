package peanats_test

import (
	"context"
	"testing"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xmock/peanatsmock"
)

func TestChainMsgMiddleware(t *testing.T) {

	t.Run("order", func(t *testing.T) {
		// Test that middleware ordering is reversed so that visually the bottom
		// middleware in the list is the outermost wrapper (first to execute)

		var executionOrder []string

		// Create a base handler that records when it's called
		baseHandler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			executionOrder = append(executionOrder, "handler")
			return nil
		})

		// Create middleware that records execution order
		middleware1 := func(next peanats.MsgHandler) peanats.MsgHandler {
			return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				executionOrder = append(executionOrder, "middleware1-before")
				err := next.HandleMsg(ctx, msg)
				executionOrder = append(executionOrder, "middleware1-after")
				return err
			})
		}

		middleware2 := func(next peanats.MsgHandler) peanats.MsgHandler {
			return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				executionOrder = append(executionOrder, "middleware2-before")
				err := next.HandleMsg(ctx, msg)
				executionOrder = append(executionOrder, "middleware2-after")
				return err
			})
		}

		middleware3 := func(next peanats.MsgHandler) peanats.MsgHandler {
			return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				executionOrder = append(executionOrder, "middleware3-before")
				err := next.HandleMsg(ctx, msg)
				executionOrder = append(executionOrder, "middleware3-after")
				return err
			})
		}

		// Chain middlewares - with reversed ordering, middleware3 should be outermost
		// Visual order: middleware1, middleware2, middleware3 (bottom to top execution)
		wrappedHandler := peanats.ChainMsgMiddleware(baseHandler, middleware1, middleware2, middleware3)

		// Create a test message using mockery-generated mock
		testMsg := peanatsmock.NewMsg(t)
		testMsg.On("Subject").Return("test.subject").Maybe()
		testMsg.On("Data").Return([]byte("test data")).Maybe()
		testMsg.On("Header").Return(make(peanats.Header)).Maybe()

		// Execute the handler
		err := wrappedHandler.HandleMsg(context.Background(), testMsg)
		if err != nil {
			t.Fatalf("Handler execution failed: %v", err)
		}

		// Expected execution order with reversed middleware wrapping:
		// The last middleware in the list (middleware3) should execute first (outermost),
		// then middleware2, then middleware1, then the handler
		expectedOrder := []string{
			"middleware3-before", // outermost middleware executes first
			"middleware2-before",
			"middleware1-before",
			"handler",           // base handler executes
			"middleware1-after", // unwinding in reverse order
			"middleware2-after",
			"middleware3-after", // outermost middleware executes last
		}

		if len(executionOrder) != len(expectedOrder) {
			t.Fatalf("Expected %d execution steps, got %d", len(expectedOrder), len(executionOrder))
		}

		for i, expected := range expectedOrder {
			if executionOrder[i] != expected {
				t.Errorf("Step %d: expected %q, got %q", i, expected, executionOrder[i])
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		// Test edge case: no middleware should return the original handler
		var called bool

		baseHandler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			called = true
			return nil
		})

		wrappedHandler := peanats.ChainMsgMiddleware(baseHandler)

		// Create a test message using mockery-generated mock
		testMsg := peanatsmock.NewMsg(t)
		testMsg.On("Subject").Return("test.subject").Maybe()
		testMsg.On("Data").Return([]byte("test data")).Maybe()
		testMsg.On("Header").Return(make(peanats.Header)).Maybe()

		// Execute the handler to verify it works
		err := wrappedHandler.HandleMsg(context.Background(), testMsg)
		if err != nil {
			t.Fatalf("Handler execution failed: %v", err)
		}

		if !called {
			t.Error("Base handler was not called through wrapped handler")
		}
	})

	t.Run("single", func(t *testing.T) {
		// Test edge case: single middleware
		var called bool

		baseHandler := peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
			called = true
			return nil
		})

		middleware := func(next peanats.MsgHandler) peanats.MsgHandler {
			return peanats.MsgHandlerFunc(func(ctx context.Context, msg peanats.Msg) error {
				return next.HandleMsg(ctx, msg)
			})
		}

		wrappedHandler := peanats.ChainMsgMiddleware(baseHandler, middleware)

		// Create a test message using mockery-generated mock
		testMsg := peanatsmock.NewMsg(t)
		testMsg.On("Subject").Return("test.subject").Maybe()
		testMsg.On("Data").Return([]byte("test data")).Maybe()
		testMsg.On("Header").Return(make(peanats.Header)).Maybe()

		err := wrappedHandler.HandleMsg(context.Background(), testMsg)
		if err != nil {
			t.Fatalf("Handler execution failed: %v", err)
		}

		if !called {
			t.Error("Base handler was not called")
		}
	})
}
