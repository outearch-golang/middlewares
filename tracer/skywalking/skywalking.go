package sky

import (
	"bytes"
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"outreach.cooperation.push.data/libs/common/error_wrapper"
	g "outreach.cooperation.push.data/libs/sky"
	"time"
)

func SetUp() error_wrapper.HandlerFunc {

	return func(ctx *gin.Context) error {

		var (
			rootSpan      go2sky.Span
			tracerContext context.Context
		)

		/* 获取上层服务的链路span */
		rootSpan, tracerContext, err := g.GSky.CreateEntrySpan(ctx.Request.Context(), ctx.Request.URL.Path, func() (string, error) {
			fmt.Println(ctx.Request.URL.Path, ctx.Request.Header.Get(propagation.Header))
			return ctx.Request.Header.Get(propagation.Header), nil
		})

		if err != nil {
			ctx.Next()
			return err
		}

		rootSpan.Tag(go2sky.TagHTTPMethod, ctx.Request.Method)
		rootSpan.Tag(go2sky.TagURL, ctx.Request.Host+ctx.Request.URL.Path)

		body, _ := ioutil.ReadAll(ctx.Request.Body)
		rootSpan.Log(time.Now(), string(body))
		rootSpan.SetSpanLayer(common.SpanLayer_Http)

		ctx.Request = ctx.Request.WithContext(tracerContext)

		id := go2sky.TraceID(tracerContext)

		ctx.Writer.Header().Set("X-Trace-Id", id)
		ctx.Set("tracer", id)

		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		ctx.Next()

		rootSpan.End()

		return nil
	}
}
