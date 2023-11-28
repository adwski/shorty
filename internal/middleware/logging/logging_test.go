package logging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	fakeResponseDuration = 10 * time.Millisecond
)

type stub struct {
	status int
	body   string
}

func (s *stub) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	<-time.After(fakeResponseDuration)
	w.WriteHeader(s.status)
	_, _ = w.Write([]byte(s.body))
}

func newLogger(w io.Writer) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(w), zapcore.DebugLevel))
}

type LogMessage struct {
	ID       string   `json:"id"`
	Level    string   `json:"level"`
	Msg      string   `json:"msg"`
	Method   string   `json:"method"`
	URI      string   `json:"uri"`
	Size     int      `json:"size"`
	Status   int      `json:"status"`
	Duration Duration `json:"duration"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) Std() time.Duration {
	return d.Duration
}

func (d *Duration) String() string {
	return d.Duration.String()
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(b, &v); err != nil {
		return
	}
	switch value := v.(type) {
	case string:
		d.Duration, err = time.ParseDuration(value)
	default:
		err = errors.New("invalid duration")
	}
	return
}

func TestMiddleware(t *testing.T) {
	type args struct {
		status int
		method string
		path   string
		body   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "log request",
			args: args{
				status: http.StatusOK,
				method: http.MethodGet,
				path:   "/",
				body:   "qweqwe",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(make([]byte, 50))
			logger := newLogger(buf)
			buf.WriteRune('\x01')

			mw := New(&Config{Logger: logger})
			s := &stub{
				status: tt.args.status,
				body:   tt.args.body,
			}
			mw.Chain(s)

			r := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			w := httptest.NewRecorder()
			mw.ServeHTTP(w, r)

			resp := w.Result()

			body, err := io.ReadAll(resp.Body)
			require.Nil(t, err)
			require.Nil(t, resp.Body.Close())

			assert.Equal(t, tt.args.status, resp.StatusCode)
			assert.Equal(t, tt.args.body, string(body))

			logOut := bufio.NewReader(buf)

			_, _ = logOut.ReadString('\x01')
			line1, err1 := logOut.ReadBytes('\n')
			require.Nil(t, err1)

			logFields := &LogMessage{}
			errj1 := json.Unmarshal(line1[:len(line1)-1], logFields)
			require.Nil(t, errj1)

			assert.Equal(t, "request", logFields.Msg)
			assert.Equal(t, tt.args.method, logFields.Method)
			assert.Equal(t, tt.args.path, logFields.URI)

			line2, err2 := logOut.ReadBytes('\n')
			require.Nil(t, err2)

			logFields2 := &LogMessage{}
			errj2 := json.Unmarshal(line2[:len(line2)-1], logFields2)
			require.Nil(t, errj2)

			assert.Equal(t, "response", logFields2.Msg)
			assert.Equal(t, len(tt.args.body), logFields2.Size)
			assert.Equal(t, tt.args.status, logFields2.Status)

			assert.True(t, logFields2.Duration.Std() > fakeResponseDuration,
				fmt.Sprintf("response duration %v should be greater than %v",
					logFields2.Duration, fakeResponseDuration))
		})
	}
}
