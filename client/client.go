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
	DiskInfo    []Disk `json:"disk_info"`
	LastUpdated string `json:"last_updated"`
	Status      string `json:"status"`
}

type Disk struct {
	Name      string `json:"name"`
	Size      string `json:"size"`
	Used      string `json:"used"`
	Available string `json:"available"`
	Usage     string `json:"usage"`
}

func getSystemInfo() string {
	cmd := exec.Command("sh", "-c", "uname -a")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getDiskInfo() []Disk {
	cmd := exec.Command("sh", "-c", "df -hl | grep sd")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	disks := make([]Disk, 0, len(lines)-1) // 减去标题行

	for i := 1; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 6 {
			continue
		}

		size := fields[1]
		used := fields[2]
		available := fields[3]
		usage := fields[4]
		mountPoint := fields[5]

		disk := Disk{
			Name:      mountPoint,
			Size:      size,
			Used:      used,
			Available: available,
			Usage:     usage,
		}

		disks = append(disks, disk)
	}

	return disks
}

func getClientInfo() ClientInfo {
	localIP := ""
	cmd := exec.Command("sh", "-c", "ip a | grep inet | grep -v inet6 | awk -F 'inet ' '{print $2}' | awk -F '/' '{print $1}' | grep 10")
	output, err := cmd.Output()
	localIP = strings.TrimSpace(string(output))
	if err == nil && localIP != "" && strings.Contains(localIP, "10") {
		systemInfo := getSystemInfo()
		diskInfo := getDiskInfo()
		lastUpdated := time.Now().Format("2006-01-02 15:04:05")
		clientInfo := ClientInfo{
			LocalIP:     localIP,
			SystemInfo:  systemInfo,
			DiskInfo:    diskInfo,
			LastUpdated: lastUpdated,
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

		fmt.Println("Received response:", response)

		conn.Close()

		time.Sleep(30 * time.Second)
	}
}
