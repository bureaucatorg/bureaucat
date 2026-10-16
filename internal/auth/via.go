package auth

import "context"

// Via records which channel a request came through; it is stored on activity log entries.
const (
	ViaWeb       = "web"
	ViaAssistant = "assistant"
	ViaPAT       = "pat"
)

type viaKey struct{}

// WithVia returns a context carrying the request channel.
func WithVia(ctx context.Context, via string) context.Context {
	return context.WithValue(ctx, viaKey{}, via)
}

// ViaFromContext returns the request channel, defaulting to ViaWeb.
func ViaFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(viaKey{}).(string); ok && v != "" {
		return v
	}
	return ViaWeb
}
