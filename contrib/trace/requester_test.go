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
	mockResp.EXPECT().Header().Return(peanats.Header{"Content-Type": []string{"application/json"}}).Maybe()
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

	// Find the attributes we expect
	var subjectAttr, typeAttr attribute.KeyValue
	for _, attr := range span.Attributes {
		if attr.Key == "nats.subject" {
			subjectAttr = attr
		}
		if attr.Key == "test.type" {
			typeAttr = attr
		}
	}
	assert.Equal(t, subject, subjectAttr.Value.AsString())
	assert.Equal(t, "request", typeAttr.Value.AsString())
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

	mockResp.EXPECT().Header().Return(peanats.Header{"Content-Type": []string{"application/json"}}).Maybe()
	mockResp.EXPECT().Value().Return(respData)

	mockReceiver.EXPECT().Next(mock.Anything).Return(mockResp, nil).Once()
	mockReceiver.EXPECT().Next(mock.Anything).Return(nil, requester.ErrOver).Once()
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
	assert.Equal(t, "test.stream", span.Name)
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

	mockReceiver.EXPECT().Next(mock.Anything).Return(nil, requester.ErrSkip)
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

func findEventByName(events []sdktrace.Event, name string) *sdktrace.Event {
	for i := range events {
		if events[i].Name == name {
			return &events[i]
		}
	}
	return nil
}

func findEventAttr(event *sdktrace.Event, key attribute.Key) *attribute.KeyValue {
	if event == nil {
		return nil
	}
	for i := range event.Attributes {
		if event.Attributes[i].Key == key {
			return &event.Attributes[i]
		}
	}
	return nil
}

func TestTracingRequester_RequestWithEventHeaders(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	respData := &testResponse{Reply: "world"}

	mockResp.EXPECT().Header().Return(peanats.Header{"X-Custom": []string{"resp-val"}}).Maybe()
	mockResp.EXPECT().Value().Return(respData).Maybe()

	mockReq.EXPECT().Request(
		mock.Anything, subject, reqData, mock.Anything,
	).Return(mockResp, nil)

	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithEventHeaders[testRequest, testResponse](),
	)

	resp, err := req.Request(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]

	// Request event should have traceparent header from propagation
	reqEvt := findEventByName(span.Events, "nats.request")
	assert.NotNil(t, reqEvt, "expected nats.request event")
	tp_attr := findEventAttr(reqEvt, "nats.header.Traceparent")
	assert.NotNil(t, tp_attr, "expected Traceparent header attribute")

	// Response event should have the X-Custom header
	respEvt := findEventByName(span.Events, "nats.response")
	assert.NotNil(t, respEvt, "expected nats.response event")
	customAttr := findEventAttr(respEvt, "nats.header.X-Custom")
	assert.NotNil(t, customAttr, "expected X-Custom header attribute")
	assert.Equal(t, "resp-val", customAttr.Value.AsString())
}

func TestTracingRequester_RequestWithEventData(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	respData := &testResponse{Reply: "world"}

	mockResp.EXPECT().Header().Return(nil).Maybe()
	mockResp.EXPECT().Value().Return(respData).Maybe()

	mockReq.EXPECT().Request(
		mock.Anything, subject, reqData, mock.Anything,
	).Return(mockResp, nil)

	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithEventData[testRequest, testResponse](0), // no truncation
	)

	resp, err := req.Request(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]

	// Request event should contain data
	reqEvt := findEventByName(span.Events, "nats.request")
	assert.NotNil(t, reqEvt)
	dataAttr := findEventAttr(reqEvt, "nats.data")
	assert.NotNil(t, dataAttr)
	assert.Contains(t, dataAttr.Value.AsString(), `"Message":"hello"`)
	lenAttr := findEventAttr(reqEvt, "nats.data_length")
	assert.NotNil(t, lenAttr)
	truncAttr := findEventAttr(reqEvt, "nats.data_truncated")
	assert.NotNil(t, truncAttr)
	assert.False(t, truncAttr.Value.AsBool())

	// Response event should contain data
	respEvt := findEventByName(span.Events, "nats.response")
	assert.NotNil(t, respEvt)
	respDataAttr := findEventAttr(respEvt, "nats.data")
	assert.NotNil(t, respDataAttr)
	assert.Contains(t, respDataAttr.Value.AsString(), `"Reply":"world"`)
}

func TestTracingRequester_RequestWithEventDataTruncation(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	respData := &testResponse{Reply: "world"}

	mockResp.EXPECT().Header().Return(nil).Maybe()
	mockResp.EXPECT().Value().Return(respData).Maybe()

	mockReq.EXPECT().Request(
		mock.Anything, subject, reqData, mock.Anything,
	).Return(mockResp, nil)

	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithEventData[testRequest, testResponse](5), // truncate to 5 chars
	)

	resp, err := req.Request(ctx, subject, reqData)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]

	reqEvt := findEventByName(span.Events, "nats.request")
	assert.NotNil(t, reqEvt)
	dataAttr := findEventAttr(reqEvt, "nats.data")
	assert.NotNil(t, dataAttr)
	assert.Len(t, dataAttr.Value.AsString(), 5)
	truncAttr := findEventAttr(reqEvt, "nats.data_truncated")
	assert.NotNil(t, truncAttr)
	assert.True(t, truncAttr.Value.AsBool())
}

func TestTracingRequester_ResponseReceiverWithEvents(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockReceiver := requestermock.NewResponseReceiver[testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "stream-req"}
	respData := &testResponse{Reply: "stream-resp"}

	mockReq.EXPECT().ResponseReceiver(
		mock.Anything, subject, reqData, mock.Anything,
	).Return(mockReceiver, nil)

	mockResp.EXPECT().Header().Return(peanats.Header{"X-Stream": []string{"1"}}).Maybe()
	mockResp.EXPECT().Value().Return(respData).Maybe()

	mockReceiver.EXPECT().Next(mock.Anything).Return(mockResp, nil).Once()
	mockReceiver.EXPECT().Next(mock.Anything).Return(nil, requester.ErrOver).Once()
	mockReceiver.EXPECT().Stop().Return(nil)

	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
		RequesterWithEventHeaders[testRequest, testResponse](),
		RequesterWithEventData[testRequest, testResponse](0),
	)

	receiver, err := req.ResponseReceiver(ctx, subject, reqData)
	assert.NoError(t, err)

	// First response
	resp, err := receiver.Next(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// End of stream
	_, err = receiver.Next(context.Background())
	assert.ErrorIs(t, err, requester.ErrOver)

	err = receiver.Stop()
	assert.NoError(t, err)

	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	span := spans[0]

	// Should have request event (from ResponseReceiver call) and response event (from Next call)
	reqEvt := findEventByName(span.Events, "nats.request")
	assert.NotNil(t, reqEvt, "expected nats.request event")
	// Request event should have data
	reqDataAttr := findEventAttr(reqEvt, "nats.data")
	assert.NotNil(t, reqDataAttr)
	assert.Contains(t, reqDataAttr.Value.AsString(), `"Message":"stream-req"`)

	respEvt := findEventByName(span.Events, "nats.response")
	assert.NotNil(t, respEvt, "expected nats.response event")
	// Response event should have headers
	streamAttr := findEventAttr(respEvt, "nats.header.X-Stream")
	assert.NotNil(t, streamAttr)
	assert.Equal(t, "1", streamAttr.Value.AsString())
	// Response event should have data
	respDataAttr := findEventAttr(respEvt, "nats.data")
	assert.NotNil(t, respDataAttr)
	assert.Contains(t, respDataAttr.Value.AsString(), `"Reply":"stream-resp"`)
}

func TestTracingRequester_NoEventsWithoutOptions(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tracer := tp.Tracer("test")
	mockReq := requestermock.NewRequester[testRequest, testResponse](t)
	mockResp := requestermock.NewResponse[testResponse](t)

	ctx := context.Background()
	subject := "test.subject"
	reqData := &testRequest{Message: "hello"}
	respData := &testResponse{Reply: "world"}

	mockResp.EXPECT().Header().Return(peanats.Header{"Content-Type": []string{"application/json"}}).Maybe()
	mockResp.EXPECT().Value().Return(respData).Maybe()

	mockReq.EXPECT().Request(
		mock.Anything, subject, reqData, mock.Anything,
	).Return(mockResp, nil)

	// No event options — should produce no events
	req := NewRequester(mockReq,
		RequesterWithTracer[testRequest, testResponse](tracer),
	)

	_, err := req.Request(ctx, subject, reqData)
	assert.NoError(t, err)

	tp.ForceFlush(ctx)
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Empty(t, spans[0].Events, "no events should be emitted without event options")
}
