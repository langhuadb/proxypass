package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg:=flag.String("f","config.json","Config file")
	flag.Parse()
	VarSignal:=make(chan os.Signal,1)
	signal.Notify(VarSignal,syscall.SIGINT,syscall.SIGKILL)
	configBytes,err:=ioutil.ReadFile(*cfg)
	if err != nil {
		panic(err)
	}
	var config []Config
	err = json.Unmarshal(configBytes,&config)
	if err != nil {
		panic(err)
	}
	go DoServer(config)
	<-VarSignal
	log.Println("Bye~")
}

//config配置
type Config struct {
	Listen uint16
	Upstream []string
}

//DoServer 启动服务器
func DoServer(config []Config){
	var handle = func(cfg Config){
		//负载均衡封装
		var getUpstream func() string
		var fid=-1
		if len(cfg.Upstream)>1{
			fmt.Println(fid)
			getUpstream = func() string {
				fid++
				if fid>=len(cfg.Upstream){
					fid=0
				}
				fmt.Println(fid)
				fmt.Println(cfg.Upstream[fid])
				return cfg.Upstream[fid]
			}
		}else {
			getUpstream= func() string {
				return cfg.Upstream[0]
			}
		}
		var Cocoon = func(conn net.Conn) {
			//处理进来的连接
			defer conn.Close()
			fmt.Println("执行getforward")
			fora :=getUpstream()
			log.Println(fora)
			facon,err:=net.Dial("tcp", fora)
			if err != nil {
				log.Println(err)
				return
			}
			defer facon.Close()
			go io.Copy(conn, facon)
			io.Copy(facon,conn)
		}
		//处理
		lis,err:=net.Listen("tcp",fmt.Sprintf("0.0.0.0:%v",cfg.Listen))
		if err != nil {
			panic(err)
		}
		defer lis.Close()
		log.Println("listen on",cfg.Listen)
		for {
			conn,err:=lis.Accept()
			if err != nil {
				continue
			}
			go Cocoon(conn)
		}
	}
	for _,cfg :=range config{
		go handle(cfg)
	}
}
