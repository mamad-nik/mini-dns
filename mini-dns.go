package minidns

type Request struct {
	ReqType  string
	Requset  string
	Response chan string
	Err      chan error
}

type MultiRequest struct {
	ReqType  string
	Requset  string
	Response chan map[string]string
	Err      chan error
}
