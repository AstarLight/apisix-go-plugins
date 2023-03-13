package plugins

import (
	"encoding/json"
	//"net/http"
	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/google/uuid"
)

func init() {
	err := plugin.RegisterPlugin(&MyRewriteResponse{})
	if err != nil {
		log.Fatalf("failed to register plugin MyRewriteResponse: %s", err)
	}
}

// it to the upstream.
type MyRewriteResponse struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type MyRewriteResponseConf struct {
	Tag string `json:"tag"`
}

func (p *MyRewriteResponse) Name() string {
	return "my-rewrite-response"
}

func (p *MyRewriteResponse) ParseConf(in []byte) (interface{}, error) {
	conf := MyRewriteResponseConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}


func (p *MyRewriteResponse) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	
	w.Header().Set("responseid", uuid.New().String())

	tag := conf.(MyRewriteResponseConf).Tag
	if len(tag) > 0 {
		_, err := w.Write([]byte(tag))
		if err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}
