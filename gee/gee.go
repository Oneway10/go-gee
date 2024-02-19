package gee

import (
	"log"
	"net/http"
	"path"
	"strings"
	"text/template"
)

type HandlerFunc func(*Context)

type (
	RouterGroup struct {
		prefix      string
		middlewares []HandlerFunc // 支持中间件
		engine      *Engine       // 所有组共享一个 engine
	}

	Engine struct {
		*RouterGroup
		router        *router
		groups        []*RouterGroup // 存储所有组
		htmlTemplates *template.Template
		funcMap       template.FuncMap
	}
)

func New() *Engine {
	engine := &Engine{
		router: newRouter(),
	}
	engine.RouterGroup = &RouterGroup{
		engine: engine,
	}
	engine.groups = []*RouterGroup{
		engine.RouterGroup,
	}

	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// 将中间件添加到组
func (group *RouterGroup) Use(middleware ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middleware...)
}

// 新建RouterGroup, 所有 group 共享一个 Engine 实例
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}

	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// 为静态文件提供服务
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)

}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET 请求
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 请求
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, r)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

// 启动 http服务器
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
