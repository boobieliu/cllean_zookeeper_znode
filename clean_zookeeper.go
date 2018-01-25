package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"zookeeper/models"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	logF = flag.String("log", "clean_zookeepr_znode.log", "Log file name")
)

func main() {
	flag.Parse()
	outfile, err := os.OpenFile(*logF, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(*outfile, "open failed")
		os.Exit(1)
	}
	log.SetOutput(outfile) //设置log的输出文件，不设置log输出默认为stdout
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cfg := flag.String("c", "cfg.json", "configuration file")
	models.ParseConfig(*cfg)
	//每30分钟执行一次
	ticker := time.NewTicker(30 * time.Minute)
	//ticker := time.NewTicker(10 * time.Second)
	for _ = range ticker.C {
		//var hosts = []string{"172.20.132.206:2181"}
		var hosts = []string{models.IP() + ":" + models.PORT()}
		var path = "/userPointLocks"
		var flags = int32(-1)
		//var flags int32 = zk.FlagEphemeral
		timestamp := time.Now().UnixNano() / 1000000

		conn, _, err := zk.Connect(hosts, time.Second*5)
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer conn.Close()

		//获取第二级目录节点
		children, _, err_children := conn.Children(path)
		log.Println("第二节点目录", children)
		if err_children != nil {
			log.Fatalln("get children status wrong", err_children)
			return
		}
		for _, children_znode := range children {
			//引入包判断第二节点是否为数字
			a := new(models.RegexCheck)
			b := a.IsInteger(children_znode)
			log.Println(b)
			if b {
				znode_path := path + "/" + children_znode
				log.Println(znode_path)
				//获取第三级临时节点路径
				ephemeral_path, _, err_children := conn.Children(znode_path)
				if err_children != nil {
					log.Println("get children status wrong", err_children)
					return
				}
				log.Println(ephemeral_path)
				for _, p := range ephemeral_path {
					ephemeral_znode_path := znode_path + "/" + p
					_, node_stat, err_get := conn.Get(ephemeral_znode_path)
					log.Println(ephemeral_znode_path)
					if err_get != nil {
						log.Println("Get ephemeral_znode failed!!", err_get)
						continue
					}
					//判断znode最后更新时间间隔是否相差10分钟(10 * 60000毫秒)，如果大于10分钟则删除该节点
					//fmt.Println(timestamp, node_stat.Mtime)
					if val := timestamp - node_stat.Mtime; val > 10*60000 {
						log.Println("deleting: ", ephemeral_znode_path)
						err_delete := conn.Delete(ephemeral_znode_path, flags)
						if err_delete != nil {
							log.Println("delete znode failed:", err_delete)
							continue
						}
					}
				}
			} else {
				continue
			}
		}
	}
}
