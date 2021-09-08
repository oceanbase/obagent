// Copyright (c) 2021 OceanBase
// obagent is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//
// http://license.coscl.org.cn/MulanPSL2
//
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package engine

import (
	"bytes"
	"container/list"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	prom "github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/api/route"
	"github.com/oceanbase/obagent/stat"
)

//RouteManager responsible for managing the pipeline corresponding to the url
type RouteManager struct {
	routeMap map[string]*list.List
	rwMutex  sync.RWMutex
}

var routeManager *RouteManager
var routeManagerOnce sync.Once

//GetRouteManager get route manager singleton
func GetRouteManager() *RouteManager {
	routeManagerOnce.Do(func() {
		routeManager = &RouteManager{
			routeMap: make(map[string]*list.List, 16),
			rwMutex:  sync.RWMutex{},
		}
	})
	return routeManager
}

//getPipelineGroup get the pipeline group corresponding to the route
func (r *RouteManager) getPipelineGroup(route string) (*list.List, error) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()

	l, exist := r.routeMap[route]
	if !exist {
		return nil, errors.New("route path is not exist")
	}
	copyList := list.New()
	copyList.PushBackList(l)
	return copyList, nil
}

//addPipelineFromPipelineGroup add pipeline instance to the route
func (r *RouteManager) addPipelineFromPipelineGroup(route string, pipeline *PipelineInstance) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	_, exist := r.routeMap[route]
	if !exist {
		r.routeMap[route] = list.New()
	}
	r.routeMap[route].PushBack(pipeline)
}

//delPipelineFromPipelineGroup delete pipeline instance to the route
func (r *RouteManager) delPipelineFromPipelineGroup(route string, pipeline *PipelineInstance) error {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	var element *list.Element
	l, exist := r.routeMap[route]
	if !exist {
		return errors.New("route path is not exist")
	}
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == pipeline {
			element = e
			break
		}
	}

	if element != nil {
		l.Remove(element)
	} else {
		return errors.New("pipeline is not exist")
	}
	return nil
}

//registerHTTPRoute register http route
func (r *RouteManager) registerHTTPRoute(exposeURL string) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	if _, exist := r.routeMap[exposeURL]; !exist {
		rt := &httpRoute{
			routePath: exposeURL,
		}

		var pullHandlerFunction = func(h http.Handler) http.Handler {
			return rt
		}

		route.RegisterPipelineRoute(GetMonitorAgentServer().Server.Router, exposeURL, pullHandlerFunction)
		r.routeMap[exposeURL] = list.New()
	}
}

type httpRoute struct {
	routePath string
}

func (rt *httpRoute) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	rt.pull(writer, request)
}

//pull mode, parallel pull involving multiple pipelines.
//After concurrent pull, the various pipelines are aggregated and written back.
func (rt *httpRoute) pull(writer http.ResponseWriter, _ *http.Request) {
	routeManager := GetRouteManager()
	pipelineGroup, err := routeManager.getPipelineGroup(rt.routePath)
	if err != nil {

		log.WithError(err).Error("pull pipeline route manager get pipeline group failed")

		if _, err := writer.Write([]byte("The current path is not registered successfully.")); err != nil {

			log.WithError(err).Error("http response write failed")

			return
		}

		return
	}

	if pipelineGroup.Len() == 0 {

		log.Warn("pull pipeline route manager pipeline group len is zero")

		if _, err := writer.Write([]byte("The current path does not have an executable pipeline.")); err != nil {

			log.WithError(err).Error("http response write failed")

			return
		}

		return
	}

	var waitGroup sync.WaitGroup
	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	var mutex sync.Mutex
	for e := pipelineGroup.Front(); e != nil; e = e.Next() {
		pipelineInstance := e.Value.(*PipelineInstance)

		waitGroup.Add(1)
		go rt.pipelinePull(&waitGroup, pipelineInstance, buffer, &mutex)

	}
	waitGroup.Wait()

	if _, err = buffer.WriteTo(writer); err != nil {
		log.WithError(err).Error("failed to write http response from buffer")
	}
}

//pipelinePull perform a pipeline pull.
// ┌────────┐       ┌────────────┐       ┌────────────┐       ┌────────────┐
// │ Input1 │------>│ Processor1 │------>│ Processor2 │------>│ Processor3 │---┐
// └────────┘       └────────────┘       └────────────┘       └────────────┘   │
// ┌────────┐       ┌────────────┐       ┌────────────┐       ┌────────────┐   │    ┌──────────┐
// │ Input2 │------>│ Processor1 │------>│ Processor2 │------>│ Processor3 │---┼--->│ exporter │
// └────────┘       └────────────┘       └────────────┘       └────────────┘   │    └──────────┘
// ┌────────┐       ┌────────────┐       ┌────────────┐       ┌────────────┐   │
// │ Input3 │------>│ Processor1 │------>│ Processor2 │------>│ Processor3 │---┘
// └────────┘       └────────────┘       └────────────┘       └────────────┘
func (rt *httpRoute) pipelinePull(waitGroup *sync.WaitGroup, p *PipelineInstance, buffer *bytes.Buffer, mutex *sync.Mutex) {
	defer waitGroup.Done()

	tStart := time.Now()
	metrics := p.parallelCompute()
	if metrics == nil || len(metrics) == 0 {
		log.Warnf("pull pipeline parallel compute result metrics is nil")
		return
	}

	newBuffer, err := p.pipeline.ExporterInstance.Export(metrics)
	if err != nil {
		log.WithError(err).Errorf("pull pipeline exporter export metrics failed")
		return
	}

	mutex.Lock()
	_, err = buffer.ReadFrom(newBuffer)
	mutex.Unlock()
	if err != nil {
		log.WithError(err).Error("pull pipeline read into buffer failed", err.Error())
		return
	}

	elapsedTimeSeconds := time.Now().Sub(tStart).Seconds()
	stat.MonAgentPipelineExecuteTotal.With(prom.Labels{"name": p.name, "status": "Successful"}).Inc()
	stat.MonAgentPipelineExecuteSecondsTotal.With(prom.Labels{"name": p.name, "status": "Successful"}).Add(elapsedTimeSeconds)

}
