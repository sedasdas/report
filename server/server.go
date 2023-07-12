package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"report/client"
	"report/database"
	"strings"
	"time"
)

var clients = make(map[string]client.ClientInfo)

func Hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("hand")
		var clientList []client.ClientInfo
		for _, client := range clients {
			clientList = append(clientList, client)
			log.Print(client)
		}

		// Encode the client list to JSON
		jsonData, err := json.Marshal(clientList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the response headers and write the JSON data
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 处理跨域请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 处理其他请求
		if r.Method == "GET" {
			// 处理 GET 请求
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
		}
	}
}
func CheckClientLastUpdated(db *database.SQLiteDB) {

	for {
		fmt.Println("Checking client status...")
		time.Sleep(1 * time.Minute) // 每分钟检查一次

		for ip, client := range clients {
			// 检查最后更新时间是否超过2分钟
			lastUpdated, _ := time.Parse("2006-01-02 15:04:05", client.LastUpdated)

			if time.Since(lastUpdated) > 2*time.Minute {
				// 设置客户端状态为离线
				client.Status = "offline"
				clients[ip] = client
				fmt.Println("Set client status to offline:", client)
				db.UpdateClient(&client)
			}
		}
	}
}

func loadClients(db *database.SQLiteDB) {
	// 读取数据库中保存的客户端信息
	clientList, err := db.GetClients()
	if err != nil {
		fmt.Println("Error retrieving clients from database:", err.Error())
		return
	}

	// 加载客户端信息到内存中的 clients 映射
	for _, client := range clientList {
		clients[client.LocalIP] = client
	}
}

// 启动服务器
func StartServer(addr string, db *database.SQLiteDB) {
	loadClients(db)
	go CheckClientLastUpdated(db)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting server:", err.Error())
		return
	}

	fmt.Println("Server started, listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go func() {
			defer conn.Close()

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				data := scanner.Text()
				data = strings.TrimSpace(data)
				fmt.Println("Received data:", data)

				// 解析客户端信息
				var clientInfo client.ClientInfo
				err = json.Unmarshal([]byte(data), &clientInfo)
				if err != nil {
					fmt.Println("Error decoding JSON:", err.Error())
					return
				}

				// 更新客户端信息
				clientInfo.LastUpdated = time.Now().Format("2006-01-02 15:04:05")
				clients[clientInfo.LocalIP] = clientInfo
				fmt.Println("Updated client info:", clientInfo)
				err = db.InsertClientInfo(clientInfo)
				if err != nil {
					fmt.Println("Insert error:", err.Error())
					return
				}

				// 发送响应
				response := map[string]interface{}{
					"status": "ok",
				}
				encoder := json.NewEncoder(conn)
				err = encoder.Encode(response)
				if err != nil {
					fmt.Println("Error encoding JSON:", err.Error())
					return
				}
			}
			if err == scanner.Err() {
				fmt.Println("Error reading data:", err.Error())
				return
			}

		}()
	}
}
