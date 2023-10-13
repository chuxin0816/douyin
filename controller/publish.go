package controller

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type PublishController struct{}

func NewPublishController() *PublishController {
	return &PublishController{}
}

func (pc *PublishController) Publish(c context.Context, ctx *app.RequestContext) {

}
