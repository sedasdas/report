package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"report/client"
	"report/database"
	"report/server"
	"time"
)

func main() {
	var serverAddr string
	var isClient bool

	flag.StringVar(&serverAddr, "server", "", "server address")
	flag.BoolVar(&isClient, "client", false, "run as client")
	flag.Parse()

	if serverAddr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if isClient {
		for {
			client.StartClient(serverAddr)
			fmt.Println("Retry connecting to server...")
			time.Sleep(10 * time.Second)
		}
	} else {
		// 创建一个 HTTP 服务器并注册 API 处理函数
		db, err := database.OpenSQLiteDB("clients.db")
		if err != nil {
			log.Fatal("Error opening database:", err.Error())
			os.Exit(1)
		}
		defer db.Close()

		err = db.CreateClientsTable()
		if err != nil {
			log.Fatal("Error creating clients table:", err.Error())
			os.Exit(1)
		}
		http.HandleFunc("/hello", server.Hello())
		// 启动 HTTP 服务器和客户端状态检查
		go http.ListenAndServe(":9999", nil)
		go server.CheckClientLastUpdated(db)

		server.StartServer(serverAddr, db)
	}
}
