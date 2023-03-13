package plugins

import (
	"encoding/json"
	"net/http"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/google/uuid"
)

func init() {
	err := plugin.RegisterPlugin(&MyRewriteRequest{})
	if err != nil {
		log.Fatalf("failed to register plugin MyRewriteRequest: %s", err)
	}
}

// it to the upstream.
type MyRewriteRequest struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type MyRewriteRequestConf struct {
	Tag string `json:"tag"`
}

func (p *MyRewriteRequest) Name() string {
	return "my-rewrite-request"
}

func (p *MyRewriteRequest) ParseConf(in []byte) (interface{}, error) {
	conf := MyRewriteRequestConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}


func (p *MyRewriteRequest) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	
	r.Header().Set("requestid", uuid.New().String())

	tag := conf.(MyRewriteRequestConf).Tag
	if len(tag) > 0 {
		// 注意，如果调用了w.Write，会直接给客户端回复。
		// 因为w.Write写入就表示 response确定了，因此无需继续执行其他插件，也不会请求到upstream的服务器，
		_, err := w.Write([]byte(tag))
		if err != nil {
			log.Errorf("failed to write: %s", err)
		}

		w.Header().Set("requestid", uuid.New().String())
	}

}
