package cambaz

import (
	"flag"
)

type mStrSlice []string

func (m *mStrSlice) String() string {
	return ""
}

func (m *mStrSlice) Set(v string) error {
	*m = append(*m, v)
	return nil
}

var (
	bind       = flag.String("b", "", "Bind ip:port to listen")
	forward    = flag.String("f", "", "Connect to ip:port to forward incoming packets")
	poolSize   = flag.String("p", "", "Pool size for connections. If not specified, cambaz creates new a connection for each incoming.")
	alternates mStrSlice
)

func init() {
	flag.Var(&alternates, "a", "List of alternates for forwarding connection")
}
