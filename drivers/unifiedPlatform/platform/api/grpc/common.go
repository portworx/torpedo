package grpc

import (
	"context"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"google.golang.org/grpc/metadata"
)

const (
	pxAccountIDKey = "px-account-id"
)

var (
	credentials *Credentials
)

// WithAccountIDMetaCtx returns the context with accountID added in metadata
func WithAccountIDMetaCtx(ctx context.Context, accountID string) context.Context {
	var md metadata.MD
	if accountID != "" {
		md = metadata.Pairs(pxAccountIDKey, accountID)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
