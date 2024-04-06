package minidns

type Request struct {
	Domain string
	IP     chan string
	Err    chan error
}
