package qlog

import (
	"context"

	"github.com/icha-senpai/note/third_party/forks/github/quic-go/quic-go"
	"github.com/icha-senpai/note/third_party/forks/github/quic-go/quic-go/qlog"
	"github.com/icha-senpai/note/third_party/forks/github/quic-go/quic-go/qlogwriter"
)

const EventSchema = "urn:ietf:params:qlog:events:http3-12"

func DefaultConnectionTracer(ctx context.Context, isClient bool, connID quic.ConnectionID) qlogwriter.Trace {
	return qlog.DefaultConnectionTracerWithSchemas(ctx, isClient, connID, []string{qlog.EventSchema, EventSchema})
}
