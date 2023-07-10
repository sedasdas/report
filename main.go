package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var clients = make(map[string]ClientInfo)

// 定义客户端信息结构体
type ClientInfo struct {
	LocalIP     string    `json:"local_ip"`
	SystemInfo  string    `json:"system_info"`
	LastUpdated time.Time `json:"last_updated"`
	Status      string    `json:"status"`
}

func hello(w http.ResponseWriter, r *http.Request) {

	log.Println("hand")
	var clientList []ClientInfo
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
		//w.Write(jsonData)

	}
}

// 获取系统信息
func getSystemInfo() string {
	cmd := exec.Command("sh", "-c", "uname -a")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func checkLive() {

}
func checkClientLastUpdated() {
	for {
		fmt.Println("checking")

		time.Sleep(1 * time.Minute) // 每分钟检查一次
		for ip, client := range clients {
			// 检查最后更新时间是否超过2分钟
			if time.Since(client.LastUpdated) > 2*time.Minute {
				// 设置客户端状态为离线
				client.Status = "offline"
				clients[ip] = client
				fmt.Println("Set client status to offline:", client)
			}
		}
	}
}

// 启动服务器
func startServer(addr string) {
	//http.HandleFunc("/hello", hello)
	//http.ListenAndServe(":9999", nil)
	//clients = make(map[string]ClientInfo)
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

			// 读取客户端信息
			reader := bufio.NewReader(conn)
			data, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading data:", err.Error())
				return
			}
			data = strings.TrimSpace(data) // 去除额外的换行符
			fmt.Println("Received data:", data)

			// 解析客户端信息
			var clientInfo ClientInfo
			err = json.Unmarshal([]byte(data), &clientInfo)
			if err != nil {
				fmt.Println("Error decoding JSON:", err.Error())
				return
			}

			// 更新客户端信息
			clientInfo.LastUpdated = time.Now()
			clients[clientInfo.LocalIP] = clientInfo
			fmt.Println("Updated client info:", clientInfo)

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
		}()
	}
}

// 启动客户端

func startClient(serverAddr string) {
	// 启动客户端
	localIP := ""
	cmd := exec.Command("sh", "-c", "ip a | grep inet | grep -v inet6 | awk -F 'inet ' '{print $2}' | awk -F '/' '{print $1}' | grep 10")
	output, err := cmd.Output()
	if err == nil {
		localIP = strings.TrimSpace(string(output))
	}

	systemInfo := getSystemInfo()
	clientInfo := ClientInfo{
		LocalIP:     localIP,
		SystemInfo:  systemInfo,
		LastUpdated: time.Now(),
		Status:      "online",
	}

	for {
		data, err := json.Marshal(clientInfo)
		if err != nil {
			fmt.Println("Error encoding JSON:", err.Error())
			return
		}

		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			fmt.Println("Error connecting to server:", err.Error())
			return
		}

		_, err = fmt.Fprintf(conn, string(data)+"\n")
		if err != nil {
			fmt.Println("Error sending data:", err.Error())
			conn.Close()
			return
		}
		fmt.Println("Sent data:", string(data))

		// 读取服务器的响应
		decoder := json.NewDecoder(conn)
		var response map[string]interface{}
		err = decoder.Decode(&response)
		if err != nil {
			fmt.Println("Error decoding JSON:", err.Error())
			conn.Close()
			return
		}
		fmt.Println("Received response:", response)

		conn.Close()

		time.Sleep(30 * time.Second)
	}
}

// 定义API处理函数

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
		startClient(serverAddr)
	} else {

		handler := http.HandlerFunc(hello)
		// 创建一个 CORS 中间件
		//c := gocors.New()
		//c.SetAllowOrigin("*")
		// 使用 CORS 中间件处理 HTTP 请求
		//h := c.Handler(handler)

		// 启动 HTTP 服务器
		go http.ListenAndServe(":9999", handler)
		go checkClientLastUpdated()
		startServer(serverAddr)

	}
}
