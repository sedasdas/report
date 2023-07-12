package client

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

type ClientInfo struct {
	LocalIP     string `json:"local_ip"`
	SystemInfo  string `json:"system_info"`
	LastUpdated string `json:"last_updated"`
	Status      string `json:"status"`
}

func getSystemInfo() string {
	cmd := exec.Command("sh", "-c", "uname -a")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// 启动客户端
func getClientInfo() ClientInfo {
	localIP := ""
	cmd := exec.Command("sh", "-c", "ip a | grep inet | grep -v inet6 | awk -F 'inet ' '{print $2}' | awk -F '/' '{print $1}' | grep 10")
	output, err := cmd.Output()
	localIP = strings.TrimSpace(string(output))
	if err == nil && localIP != "" && strings.Contains(localIP, "10") {
		systemInfo := getSystemInfo()
		lastupdated := time.Now().Format("2006-01-02 15:04:05")
		clientInfo := ClientInfo{
			LocalIP:     localIP,
			SystemInfo:  systemInfo,
			LastUpdated: lastupdated,
			Status:      "online",
		}
		return clientInfo
	}
	return ClientInfo{}
}
func StartClient(serverAddr string) {
	// 启动客户端

	for {
		cf := getClientInfo()
		if cf.LocalIP == "" {
			return
		}
		data, err := json.Marshal(cf)
		if err != nil {
			fmt.Println("Error encoding JSON:", err.Error())
			return
		}

		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			fmt.Println("Error connecting to server:", err.Error())
			// 连接失败时进行重试
			time.Sleep(10 * time.Second)
			continue
		}

		_, err = fmt.Fprintf(conn, string(data)+"\n")
		if err != nil {
			fmt.Println("Error sending data:", err.Error())
			conn.Close()
			// 发送数据失败时进行重试
			time.Sleep(10 * time.Second)
			continue
		}
		fmt.Println("Sent data:", string(data))

		// 读取服务器的响应
		decoder := json.NewDecoder(conn)
		var response map[string]interface{}
		err = decoder.Decode(&response)
		if err != nil {
			fmt.Println("Error decoding JSON:", err.Error())
			conn.Close()
			// 接收响应失败时进行重试
			time.Sleep(10 * time.Second)
			continue
		}
		fmt.Println("Received response:", response)

		conn.Close()

		time.Sleep(30 * time.Second)
	}
}
