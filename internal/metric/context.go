package metric

import "context"

func ContextWithStream(ctx context.Context, stream string) context.Context {
	return context.WithValue(ctx, "stream", stream)
}

func StreamFromContext(ctx context.Context) string {
	stream, ok := ctx.Value("stream").(string)
	if !ok {
		stream = "unknown"
	}
	return stream
}
