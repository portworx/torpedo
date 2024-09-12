package grpc

import (
	"context"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	commonapis "github.com/pure-px/apis/public/portworx/common/apiv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type PdsGrpc struct {
	ApiClientV2 *grpc.ClientConn
	AccountId   string
}

const (
	pxAccountIDKey = "px-account-id"
)

var (
	credentials *utils.Credentials
)

// WithAccountIDMetaCtx returns the context with accountID added in metadata
func WithAccountIDMetaCtx(ctx context.Context, accountID string) context.Context {
	var md metadata.MD
	if accountID != "" {
		md = metadata.Pairs(pxAccountIDKey, accountID)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func NewPaginationRequest(pageNumber, pageSize int) *commonapis.PageBasedPaginationRequest {
	return &commonapis.PageBasedPaginationRequest{
		PageNumber: int64(pageNumber),
		PageSize:   int64(pageSize),
	}
}
