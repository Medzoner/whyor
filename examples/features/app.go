package features

import "fmt"

type Handler interface{ Name() string }

type Users struct{}
type Orders struct{}

func (*Users) Name() string  { return "users" }
func (*Orders) Name() string { return "orders" }

func NewUsers() *Users   { return &Users{} }
func NewOrders() *Orders { return &Orders{} }

type Router struct{ Handlers []Handler }

func NewRouter(hs []Handler) *Router { return &Router{hs} }

type Cache struct{ log *[]string }

func (c *Cache) Close() error { *c.log = append(*c.log, "cache closed"); return nil }

func NewCache(log *[]string) *Cache { return &Cache{log} }

type Server struct {
	Router *Router
	Cache  *Cache
	H      Handler
}

func NewServer(r *Router, c *Cache, h Handler) *Server { return &Server{r, c, h} }

func (s *Server) String() string { return fmt.Sprint(len(s.Router.Handlers), " handlers") }
