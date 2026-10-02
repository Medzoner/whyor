package structs

type Config struct {
	Addr string
	Port int
}

type Logger struct{}

type Deps struct {
	Cfg    *Config
	Log    *Logger
	Skip   string `whyor:"-"`
	hidden int
}

func NewConfig() *Config { return &Config{Addr: "localhost", Port: 80} }
func NewLogger() *Logger { return &Logger{} }

type Server struct{ Addr string }

func NewServer(addr string, port int, d *Deps) *Server {
	return &Server{Addr: addr + ":" + string(rune('0'+port/10%10)) + d.Skip}
}
