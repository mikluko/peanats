package peanats

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testAdapterArgument struct {
	Arg string `json:"arg"`
}

type testAdapterResult struct {
	Res string `json:"res"`
}

func TestTyped(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		arg := testAdapterArgument{Arg: "Hello, world!"}
		res := testAdapterResult{Res: "Goodbye, world!"}

		p := TypedHandlerFunc[testAdapterArgument, testAdapterResult](
			func(pub TypedPublisher[testAdapterResult], req TypedRequest[testAdapterArgument]) error {
				assert.Equal(t, req.Payload(), &arg)
				return pub.Publish(&res)
			},
		)

		f := Typed[testAdapterArgument, testAdapterResult](&JsonCodec{}, p)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		req := new(requestMock)
		defer req.AssertExpectations(t)

		req.On("Data").Return(must(json.Marshal(&arg)))

		pub.On("Publish", mock.Anything).Run(func(args mock.Arguments) {
			data := args.Get(0).([]byte)
			assert.Equal(t, must(json.Marshal(&res)), data)
		}).Return(nil)

		err := f.Serve(pub, req)
		require.NoError(t, err)
	})
	t.Run("decode error", func(t *testing.T) {
		p := TypedHandlerFunc[testAdapterArgument, testAdapterResult](
			func(pub TypedPublisher[testAdapterResult], req TypedRequest[testAdapterArgument]) error {
				t.Fatal("should not be called")
				return nil
			},
		)

		errDecodeError := errors.New("decode error")

		c := new(codecMock)
		defer c.AssertExpectations(t)

		f := Typed[testAdapterArgument, testAdapterResult](c, p)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		req := new(requestMock)
		defer req.AssertExpectations(t)

		req.On("Data").Return([]byte("Hello, world!"))
		c.On("Decode", []byte("Hello, world!"), mock.Anything).Return(errDecodeError)

		err := f.Serve(pub, req)
		require.Error(t, err)
		require.True(t, errors.Is(err, errDecodeError))

		var errObj *Error
		require.True(t, errors.As(err, &errObj))
		require.Equal(t, http.StatusBadRequest, errObj.Code)
	})
	t.Run("encode error", func(t *testing.T) {
		arg := testAdapterArgument{Arg: "Hello, world!"}
		res := testAdapterResult{Res: "Goodbye, world!"}

		c := new(codecMock)
		defer c.AssertExpectations(t)

		af := TypedHandlerFunc[testAdapterArgument, testAdapterResult](
			func(pub TypedPublisher[testAdapterResult], req TypedRequest[testAdapterArgument]) error {
				return pub.Publish(&res)
			},
		)
		f := Typed[testAdapterArgument, testAdapterResult](c, af)

		pub := new(publisherMock)
		defer pub.AssertExpectations(t)

		req := new(requestMock)
		defer req.AssertExpectations(t)

		req.On("Data").Return(must(json.Marshal(&arg)))

		encodeErr := errors.New("encode error")
		c.On("Decode", mock.Anything, mock.Anything).Return(nil)
		c.On("Encode", mock.Anything).Return(nil, encodeErr)

		err := f.Serve(pub, req)
		require.Error(t, err)
		require.True(t, errors.Is(err, encodeErr))

		var errObj *Error
		require.True(t, errors.As(err, &errObj))
		require.Equal(t, http.StatusInternalServerError, errObj.Code)
	})
}

type codecMock struct {
	mock.Mock
}

func (c *codecMock) Encode(v any) ([]byte, error) {
	args := c.Called(v)
	if err := args.Error(1); err != nil {
		return nil, err
	}
	return args.Get(0).([]byte), nil
}

func (c *codecMock) Decode(data []byte, vPtr any) error {
	args := c.Called(data, vPtr)
	return args.Error(0)
}
