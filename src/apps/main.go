package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/natefinch/pie"
	"net/http"
	"net/rpc/jsonrpc"
	"net/url"
	"os"
	"regexp"
)

var (
	name = "apps"
	srv  pie.Server
)

type PlugRequest struct {
	Body     string
	Header   http.Header
	Form     url.Values
	PostForm url.Values
	Url      string
	Method   string
	HeadVals map[string]string
	Status   int
}

type ReturnMsg struct {
	Method string
	Err    string
	Plugin string
	Email  string
}

type api struct{}

type GetApplicationsListReply struct {
	Applications []Connection
}

func GetList(args PlugRequest, reply *PlugRequest, name string) {
	reply.HeadVals = make(map[string]string, 1)
	reply.HeadVals["Content-Type"] = "application/json; charset=UTF-8"
	reply.Status = 200
	connections := ListApplications(reply)
	rsp, err := json.Marshal(connections)
	if err != nil {
		reply.Status = 500
		log.Error("Marshalling of connections for all users failed: ", err)
	}
	reply.Body = string(rsp)

}

func GetListForCurrentUser(args PlugRequest, reply *PlugRequest, sam string) {

	reply.HeadVals = make(map[string]string, 1)
	reply.HeadVals["Content-Type"] = "application/json; charset=UTF-8"
	reply.Status = 200
	connections := ListApplicationsForSamAccount(sam, reply)

	rsp, err := json.Marshal(connections)
	if err != nil {
		reply.Status = 500
		log.WithFields(log.Fields{
			"SAMAccount": sam,
		}).Error("Marshalling of connections for current user failed: ", err)

	}
	reply.Body = string(rsp)
}

func UnpublishApplication(args PlugRequest, reply *PlugRequest, name string) {
	reply.HeadVals = make(map[string]string, 1)
	reply.HeadVals["Content-Type"] = "application/json; charset=UTF-8"
	reply.Status = 200
	if name != "" {
		UnpublishApp(name)
	} else {
		reply.Status = 500
		log.Error("No Application name to unpublish")
	}
}

var tab = []struct {
	Url    string
	Method string
	f      func(PlugRequest, *PlugRequest, string)
}{
	{`^\/api\/apps\/{0,1}$`, "GET", GetList},
	{`^\/api\/apps\/(?P<id>[^\/]+)\/{0,1}$`, "DELETE", UnpublishApplication},
	{`^\/api\/apps\/(?P<id>[^\/]+)\/{0,1}$`, "GET", GetListForCurrentUser},
}

func (api) Receive(args PlugRequest, reply *PlugRequest) error {
	for _, val := range tab {
		re := regexp.MustCompile(val.Url)
		match := re.MatchString(args.Url)
		if val.Method == args.Method && match {
			if len(re.FindStringSubmatch(args.Url)) == 2 {
				val.f(args, reply, re.FindStringSubmatch(args.Url)[1])
			} else {
				val.f(args, reply, "")
			}
		}
	}
	return nil
}

func (api) Plug(args interface{}, reply *bool) error {
	*reply = true
	return nil
}

func (api) Check(args interface{}, reply *bool) error {
	*reply = true
	return nil
}

func (api) Unplug(args interface{}, reply *bool) error {
	defer os.Exit(0)
	*reply = true
	return nil
}

func main() {
	var err error

	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	srv = pie.NewProvider()

	if err = srv.RegisterName(name, api{}); err != nil {
		log.Fatal("Failed to register %s: %s", name, err)
	}

	initConf()

	srv.ServeCodec(jsonrpc.NewServerCodec)
}
