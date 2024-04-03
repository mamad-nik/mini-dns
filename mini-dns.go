package minidns

type Dn struct {
	Domain string
	IP     chan<- string
	Err    chan<- error
}
