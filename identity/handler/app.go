package handler

import (
	"context"
	"github.com/sparrow-community/protos/identity"
)

type AppServiceHandler struct {
}

func (a AppServiceHandler) Register(ctx context.Context, request *identity.AppRegisterRequest, response *identity.AppRegisterResponse) error {
	//TODO implement me
	panic("implement me")
}

func (a AppServiceHandler) PageResources(ctx context.Context, request *identity.AppPageResourceRequest, response *identity.AppPageResourceResponse) error {
	//TODO implement me
	panic("implement me")
}

func (a AppServiceHandler) ApiResources(ctx context.Context, request *identity.AppApiResourceRequest, response *identity.AppApiResourceResponse) error {
	//TODO implement me
	panic("implement me")
}
