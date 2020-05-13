package jaeger

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
	"io/ioutil"
	"outreach.cooperation.push.data/libs/common/error_wrapper"
	"outreach.cooperation.push.data/libs/configs"
)

func SetUp() error_wrapper.HandlerFunc {

	return func(c *gin.Context) error {

		var (
			err  error
			body []byte
		)

		if configs.GConfig.GetBool("trace.open") {

			var parentSpan opentracing.Span

			sender := transport.NewHTTPTransport(configs.GConfig.GetString("trace.host"))
			jTracer, io := jaeger.NewTracer(configs.GConfig.GetString("server.AppName"),
				jaeger.NewConstSampler(true),
				jaeger.NewRemoteReporter(sender, jaeger.ReporterOptions.Logger(jaeger.StdLogger)),
			)

			defer io.Close()

			spCtx, err := jTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))

			body, _ = ioutil.ReadAll(c.Request.Body)

			if err != nil {
				parentSpan = jTracer.StartSpan(
					c.Request.URL.Path,
					opentracing.Tags{
						"入参":                  string(body),
						string(ext.Component): "HTTP",
					},
				)
			} else {
				parentSpan = jTracer.StartSpan(
					c.Request.URL.Path,
					opentracing.ChildOf(spCtx),
					opentracing.Tags{
						"入参":                  string(body),
						string(ext.Component): "HTTP",
					},
				)
			}
			defer parentSpan.Finish()

			//输出链路ID
			sp := parentSpan.Context().(jaeger.SpanContext)
			c.Writer.Header().Set("X-Trace-Id", sp.TraceID().String())

			c.Set("Tracer", jTracer)
			c.Set("ParentSpanContext", parentSpan.Context())
		}

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		c.Next()

		return err
	}
}
