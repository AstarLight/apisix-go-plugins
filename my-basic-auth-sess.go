package plugins

import (
	"encoding/json"
	"net/http"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/google/uuid"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"time"
)

var (
	ctx     = context.Background()
	RedisDb *redis.Client
)

func init() {
	err := plugin.RegisterPlugin(&MyBasicAuthSess{})
	if err != nil {
		log.Fatalf("failed to register plugin MyBasicAuthSess: %s", err)
	}
}

// it to the upstream.
type MyBasicAuthSess struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

type MyBasicAuthSessConf struct {
	RedisAddr string `json:"redis_addr"`
	RedisDB int `json:"redis_db"`
	RedisPasswd string `json:"redis_passwd"`
}

func (p *MyBasicAuthSess) Name() string {
	return "my-basic-auth-sess"
}

func initRedis(conf interface{}) error {
	RedisAddr := conf.(MyBasicAuthSessConf).RedisAddr
	RedisDb = redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})

	_, err := RedisDb.Ping(ctx).Result()
	if err != nil {
		log.Infof("connect redis fail, err:%v", err)
		RedisDb = nil
		return err
	} else {
		log.Infof("connect redis succ")
	}

	return nil
}

func (p *MyBasicAuthSess) ParseConf(in []byte) (interface{}, error) {
	conf := MyBasicAuthSessConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}

func GetSession(username string) string {
	if len(username) <= 0 {
		return ""
	}
	redisSess, err := RedisDb.Get(ctx, username).Result() // 先取header的seesionid，然后去redis查看是否存在该session
	if err != nil {
		log.Infof("GetSession redis fail username=%s", username)
		return ""
	}

	return redisSess

}

func CheckSession(username, sessionid string) bool {
	if len(username) <= 0 {
		return false
	}
	redisSess, err := RedisDb.Get(ctx, username).Result() // 先取header的seesionid，然后去redis查看是否存在该session
	if err != nil {
		return false
	}

	return redisSess == sessionid

}

func CheckPasswd(username, password string) bool {
	if len(username) <= 0 || len(password) <= 0 {
		return false
	}

	key := username + "&" + password
	
	_, err := RedisDb.Get(ctx, key).Result() // 去redis检查是否有这个账密
	if err != nil {
		return false
	}

	return true
}


func (p *MyBasicAuthSess) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	
	if RedisDb == nil {
		err := initRedis(conf)
		if err != nil {
			log.Infof("initRedis fail")
			return
		}
	}

	r.RespHeader().Set("rtag", "this set when request")

	sessionid := r.Header().Get("sessionid")
	username := r.Header().Get("username")
	traceid := uuid.New().String()

	if !CheckSession(username, sessionid) {
		// 不存在该session，检查账密
		password := r.Header().Get("password")

		if !CheckPasswd(username, password) {
			log.Infof("fail username passwd check")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// 账密检查完成，开始生成session，并写入redis，15分钟的session过期时间
		newSessionId := uuid.New().String()
		err := RedisDb.Set(ctx, username, newSessionId, 15*60*time.Second).Err()
		if err != nil {
			log.Infof("set session fail")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Header().Set("traceid", traceid)
		r.Header().Set("password", "")
		r.Header().Set("username", username)

	} else {
		// 存在该session
		r.Header().Set("traceid", traceid)
		r.Header().Set("password", "")
		r.Header().Set("username", username)
	}
	
}

func (p *MyBasicAuthSess) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	username := w.Header().Get("username")
	sessionid := GetSession(username)
	log.Infof("ResponseFilter %s", sessionid)
	// 把session写回header
	w.Header().Set("sessionid", sessionid)
}