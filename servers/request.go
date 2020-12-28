package servers

import (
	"encoding/json"
	"erpc/pb"
)

type Request struct {
	*pb.RequestContext                        //rpc请求参数
	input              map[string]interface{} //参数
}

func (r *Request) GetService() string {
	//todo
	return ""
}

func (r *Request) GetMethod() string {
	//todo
	return ""
}

func (r *Request) GetForm() map[string]interface{} {
	if r.input == nil {
		json.Unmarshal([]byte(r.RequestContext.Input), r.input)
	}
	return r.input
}
