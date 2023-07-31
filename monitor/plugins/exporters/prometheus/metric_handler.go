package prometheus

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/trace"
	agentlog "github.com/oceanbase/obagent/log"
	"github.com/oceanbase/obagent/monitor/engine"
	"github.com/oceanbase/obagent/monitor/message"
)

type httpRoute struct {
	routePath string
}

func (rt *httpRoute) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := context.WithValue(context.Background(), agentlog.TraceIdKey{}, request.Header.Get(trace.TraceIdHeader))
	curctx := context.WithValue(ctx, agentlog.StartTimeKey, time.Now())
	defer log.WithContext(curctx).WithField("url", request.RequestURI).Debug("pull metrics end")

	reader, err := getPipelieGroupReader(ctx, rt.routePath)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to readFromPipelieGroup")
		return
	}
	writer.Header().Set("Content-Type", request.Header.Get("Accept"))
	if _, err = reader.WriteTo(writer); err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to write http response from buffer")
	}
}

// getPipelieGroupReader reader with bytes of data, such as prometheus, alarms
func getPipelieGroupReader(ctx context.Context, uri string) (*bytes.Buffer, error) {
	pipelineCache, err := engine.GetRouteManager().GetPipelineGroup(uri)
	if err != nil {
		log.WithContext(ctx).WithField("routePath", uri).WithError(err).Error("get pipeline group failed")
		return nil, errors.Errorf("get pipeline group of %s failed", uri)
	}
	if pipelineCache.Len() == 0 {
		log.WithContext(ctx).WithField("routePath", uri).Error("pull pipeline route manager pipeline cache is zero")
		return nil, errors.Errorf("get pipeline group of %s result is nil", uri)
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	msgs := make([]*message.Message, 0, 512)

	for e := pipelineCache.Front(); e != nil; e = e.Next() {
		cache, ok := e.Value.(map[string]*message.Message)
		if !ok {
			log.WithContext(ctx).WithField("routePath", uri).Error("routeMap list data is not correct")
			continue
		}
		for _, v := range cache {
			msgs = append(msgs, v)
		}

	}

	msgBuffer, err := transformMetric(ctx, msgs)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("transform message failed")
	}
	buffer.Write(msgBuffer.Bytes())

	return buffer, nil
}

func transformMetric(ctx context.Context, metrics []*message.Message) (*bytes.Buffer, error) {
	collector := message.NewCollector(nil)
	collector.Fam = message.CreateMetricFamily(metrics)
	registry := prometheus.NewRegistry()
	err := registry.Register(collector)
	if err != nil {
		return nil, errors.Wrap(err, "exporter prometheus register collector")
	}
	metricFamilies, err := registry.Gather()
	if err != nil {
		return nil, errors.Wrap(err, "exporter prometheus registry gather")
	}

	//filter labels, delete labels which value is empty
	for _, family := range metricFamilies {
		for _, metric := range family.Metric {
			labels := metric.GetLabel()
			var newLabels = make([]*io_prometheus_client.LabelPair, 0)
			for _, label := range labels {
				newLabels = append(newLabels, label)
			}
			metric.Label = newLabels
		}
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	encoder := expfmt.NewEncoder(buffer, expfmt.FmtText)
	for _, metricFamily := range metricFamilies {
		err := encoder.Encode(metricFamily)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("exporter encode metricFamily failed")
			continue
		}
	}
	return buffer, nil
}
