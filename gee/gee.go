package gee

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

// Engine 实现 ServeHTTP的接口
type Engine struct {
	router map[string]HandlerFunc
}

// gee.Engine 的构造函数
func New() *Engine {
	return &Engine{
		router: make(map[string]HandlerFunc),
	}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	engine.router[key] = handler
}

// 添加 GET 请求方法
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// 添加 POST 请求方法
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + "-" + r.URL.Path
	if handler, ok := engine.router[key]; ok {
		handler(w, r)
	} else {
		fmt.Fprintf(w, "404 Not Found: %s\n", r.URL)
	}
}

// 定义启动 http 服务器的方法
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}