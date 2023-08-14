/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package engine

import (
	"container/list"
	"context"
	"github.com/oceanbase/obagent/errors"
	"net/http"
	"reflect"
	"sync"
)

// PipelineRouteHandler responsible for transferring the pipeline exposeUrl and handler
type PipelineRouteHandler struct {
	Ctx         context.Context
	ExposeUrl   string
	FuncHandler func(http.Handler) http.Handler
}

var PipelineRouteChan = make(chan *PipelineRouteHandler, 10)

// RouteManager responsible for managing the pipeline corresponding to the url
type RouteManager struct {
	routeMap map[string]*list.List
	rwMutex  sync.RWMutex
}

var routeManager *RouteManager
var routeManagerOnce sync.Once

// GetRouteManager get route manager singleton
func GetRouteManager() *RouteManager {

	routeManagerOnce.Do(func() {
		routeManager = &RouteManager{
			routeMap: make(map[string]*list.List, 16),
			rwMutex:  sync.RWMutex{},
		}
	})
	return routeManager
}

// GetPipelineGroup get the pipeline group corresponding to the route
func (r *RouteManager) GetPipelineGroup(route string) (*list.List, error) {
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

// AddPipelineGroup add data to the route
func (r *RouteManager) AddPipelineGroup(route string, data interface{}) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	_, exist := r.routeMap[route]
	if !exist {
		r.routeMap[route] = list.New()
	}
	r.routeMap[route].PushBack(data)
}

// delPipelineFromPipelineGroup delete pipeline instance to the route
func (r *RouteManager) DeletePipelineGroup(route string, data interface{}) error {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()

	var element *list.Element
	l, exist := r.routeMap[route]
	if !exist {
		return errors.New("route path is not exist")
	}
	for e := l.Front(); e != nil; e = e.Next() {
		if reflect.DeepEqual(e.Value, data) {
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

// RegisterHTTPRoute register http route
func (r *RouteManager) RegisterHTTPRoute(ctx context.Context, exposeURL string, handler http.Handler) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	if _, exist := r.routeMap[exposeURL]; !exist {
		handlerFunc := func(h http.Handler) http.Handler {
			return handler
		}
		pipelineRoute := &PipelineRouteHandler{
			Ctx:         ctx,
			ExposeUrl:   exposeURL,
			FuncHandler: handlerFunc,
		}
		PipelineRouteChan <- pipelineRoute
		r.routeMap[exposeURL] = list.New()
	}
}
