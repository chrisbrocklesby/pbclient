package pbclient

import (
	"context"
	"errors"
	"net"
	"strings"
)

func joinQuery(q string) string {
	if q == "" {
		return ""
	}
	if q[0] == '?' {
		return q[1:]
	}
	return q
}

func optParam(p []string) string {
	if len(p) == 0 {
		return ""
	}
	return joinQuery(p[0])
}

func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout() || ne.Temporary()
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "unexpected eof") ||
		strings.Contains(msg, "realtime disconnected")
}
