package alias

type Conn struct{}

// DB is an alias of Conn: generated code must keep working with it.
type DB = Conn

type Repo struct{ DB *DB }

func NewDB() *DB             { return &Conn{} }
func NewRepo(db *Conn) *Repo { return &Repo{db} }
