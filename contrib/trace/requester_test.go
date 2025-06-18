package trace

import (
	"context"
	"errors"
	"testing"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xmock/requestermock"
	"github.com/mikluko/peanats/requester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type testRequest struct {
	Message string
}

type testResponse struct {
	Reply string
}

func TestNewRequester(t *testing.T) {
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	tracer := otel.Tracer("test")

	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithSpanName[testRequest, testResponse]("custom.request"),
		RequesterWithAttributes[testRequest, testResponse](
			attribute.String("test.attr", "value"),
		),
	)

	assert.NotNil(t, req)
}

func TestTracingRequester_Request(t *testing.T) {
	// Setup
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	respData := &testResponse{Reply: "world"}

	// Setup expectations
	mockResp.EXPECT().Header().Return(peanats.Header{"Content-Type": []string{"application/json"}})
	mockResp.EXPECT().Value().Return(respData)

	mockReq.EXPECT().Request(
		mock.AnythingOfType("*context.valueCtx"),
		subject,
		reqData,
		mock.AnythingOfType("[]requester.RequestOption"),
	).Run(func(ctx context.Context, subj string, data *testRequest, opts ...requester.RequestOption) {
		// Verify trace context was injected by checking at least one option exists
		assert.NotEmpty(t, opts)
	}).Return(mockResp, nil)

	// Create tracing requester
	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithSpanName[testRequest, testResponse]("test.request"),
		RequesterWithAttributes[testRequest, testResponse](
			attribute.String("test.type", "request"),
		),
	)

	// Execute request
	resp, err := req.Request(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, respData, resp.Value())

	// Verify span was created
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "test.request", span.Name)
	assert.Equal(t, trace.SpanKindClient, span.SpanKind)
	assert.Equal(t, subject, span.Attributes[0].Value.AsString())
	assert.Equal(t, "request", span.Attributes[1].Value.AsString())
}

func TestTracingRequester_RequestError(t *testing.T) {
	// Setup
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	expectedErr := errors.New("request failed")

	// Setup expectations
	mockReq.EXPECT().Request(
		mock.AnythingOfType("*context.valueCtx"),
		subject,
		reqData,
		mock.AnythingOfType("[]requester.RequestOption"),
	).Return(nil, expectedErr)

	// Create tracing requester
	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
	)

	// Execute request
	resp, err := req.Request(ctx, subject, reqData)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)

	// Verify span was created with error
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "peanats.request", span.Name)
	assert.Len(t, span.Events, 1)
	assert.Equal(t, "exception", span.Events[0].Name)
}

func TestTracingRequester_ResponseReceiver(t *testing.T) {
	// Setup
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockReceiver := requestermock.NewResponseReceiver[testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	respData := &testResponse{Reply: "world"}

	// Setup expectations
	mockReq.EXPECT().ResponseReceiver(
		mock.AnythingOfType("*context.valueCtx"),
		subject,
		reqData,
		mock.AnythingOfType("[]requester.ResponseReceiverOption"),
	).Return(mockReceiver, nil)

	mockResp.EXPECT().Header().Return(peanats.Header{"Content-Type": []string{"application/json"}})
	mockResp.EXPECT().Value().Return(respData)

	mockReceiver.EXPECT().Next(mock.AnythingOfType("*context.emptyCtx")).Return(mockResp, nil)
	mockReceiver.EXPECT().Next(mock.AnythingOfType("*context.emptyCtx")).Return(nil, requester.ErrOver)
	mockReceiver.EXPECT().Stop().Return(nil)

	// Create tracing requester
	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithSpanName[testRequest, testResponse]("test.stream"),
	)

	// Create response receiver
	receiver, err := req.ResponseReceiver(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, receiver)

	// Get first response
	resp, err := receiver.Next(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, respData, resp.Value())

	// Get end of stream
	resp, err = receiver.Next(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, requester.ErrOver))
	assert.Nil(t, resp)

	// Stop receiver
	err = receiver.Stop()
	assert.NoError(t, err)

	// Verify span was created
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "test.stream", span.Name())
	assert.Equal(t, trace.SpanKindClient, span.SpanKind)
}

func TestTracingResponseReceiver_StopWithError(t *testing.T) {
	// Setup
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockReceiver := requestermock.NewResponseReceiver[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	stopErr := errors.New("stop failed")

	// Setup expectations
	mockReq.EXPECT().ResponseReceiver(
		mock.AnythingOfType("*context.valueCtx"),
		subject,
		reqData,
		mock.AnythingOfType("[]requester.ResponseReceiverOption"),
	).Return(mockReceiver, nil)

	mockReceiver.EXPECT().Stop().Return(stopErr)

	// Create tracing requester
	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
	)

	// Create response receiver
	receiver, err := req.ResponseReceiver(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, receiver)

	// Stop receiver with error
	err = receiver.Stop()
	assert.Error(t, err)
	assert.Equal(t, stopErr, err)

	// Verify span was created with error
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Len(t, span.Events, 1)
	assert.Equal(t, "exception", span.Events[0].Name)
}

func TestTracingResponseReceiver_NextWithSkip(t *testing.T) {
	// Setup
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockReceiver := requestermock.NewResponseReceiver[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}

	// Setup expectations
	mockReq.EXPECT().ResponseReceiver(
		mock.AnythingOfType("*context.valueCtx"),
		subject,
		reqData,
		mock.AnythingOfType("[]requester.ResponseReceiverOption"),
	).Return(mockReceiver, nil)

	mockReceiver.EXPECT().Next(mock.AnythingOfType("*context.emptyCtx")).Return(nil, requester.ErrSkip)
	mockReceiver.EXPECT().Stop().Return(nil)

	// Create tracing requester
	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
	)

	// Create response receiver
	receiver, err := req.ResponseReceiver(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, receiver)

	// Get skip response
	resp, err := receiver.Next(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, requester.ErrSkip))
	assert.Nil(t, resp)

	// Stop receiver
	err = receiver.Stop()
	assert.NoError(t, err)

	// Verify span was created without error events
	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Len(t, span.Events, 0) // No error events for ErrSkip
}
