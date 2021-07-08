package sinister

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

var ErrInvalidParam = errors.New("sinister: invalid param")

type Handler func(*HC)

type router struct {
	routes          []*route
	node            *node
	pool            *sync.Pool
	NotFoundHandler Handler
	logger          *zap.Logger
}

type route struct {
	path    string
	rawPath string
	method  string
	handler Handler
	params  []string
}

func (r *router) setNotFoundHandler() {
	r.NotFoundHandler = func(ctx *HC) {
		ctx.MIME(ApplicationJSON)
		ctx.JSONS(404, "Not found")
	}
}

func newRouter(logger *zap.Logger) *router {
	r := &router{
		routes: nil,
		node:   nil,
		pool: &sync.Pool{
			New: func() interface{} { return newHC() },
		},
		logger: logger,
	}
	r.setNotFoundHandler()
	return r
}

func findParam(params []*Param, param string) string {
	for _, p := range params {
		if p.Name == param {
			return p.Value
		}
	}
	return ""
}

func newRoute(path, rawPath, method string, h Handler, params []string) *route {
	return &route{
		path:    path,
		rawPath: rawPath,
		method:  method,
		handler: h,
		params:  params,
	}
}

func setParams(params []string, values []string) []*Param {
	if len(params) == 0 || len(params) != len(values) {
		return nil
	}
	paramsOut := make([]*Param, len(params))
	param := &Param{}
	for i, v := range params {
		param = &Param{Name: v, Value: values[i]}
		paramsOut[i] = param
	}
	return paramsOut
}

func (router *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	formattedPath, params, valid := validateRequestPath(r.URL.Path, r.Method)
	fmt.Println(formattedPath, params, valid)
	if valid {
		route := findNode(router.node, formattedPath)
		ctxLogger := router.logger.With(zap.String("plm", "test"))
		lib := router.pool.Get().(*HC)
		lib.reset()
		if route != nil && isMatch(route.rawPath, formattedPath) {
			fmt.Println("is match")
			urlParams := setParams(route.params, params)
			lib.set(w, r, ctxLogger, urlParams)
			route.handler(lib)
		} else {
			lib.set(w, r, ctxLogger, nil)
			router.NotFoundHandler(lib)
		}
		router.pool.Put(lib)
		fmt.Println("ok")
	}
}
