package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // 导入 SQLite3 驱动程序
	"log"
	"report/client"
)

// SQLiteDB 是一个封装了 SQLite 数据库连接的结构体
type SQLiteDB struct {
	db *sql.DB
}

// OpenSQLiteDB 打开 SQLite 数据库连接并返回 SQLiteDB 实例
func OpenSQLiteDB(dbFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	return &SQLiteDB{
		db: db,
	}, nil
}

// CreateClientsTable 创建 clients 表格
func (db *SQLiteDB) CreateClientsTable() error {
	_, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS clients (
		local_ip TEXT PRIMARY KEY,
		system_info TEXT,
		last_updated TIME,
		status TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}

// InsertClientInfo 将客户信息插入数据库
func (db *SQLiteDB) InsertClientInfo(clientInfo client.ClientInfo) error {
	existingClient, err := db.GetClientByLocalIP(clientInfo.LocalIP)
	if err != nil {
		return err
	}
	if existingClient != nil {
		existingClient.SystemInfo = clientInfo.SystemInfo
		existingClient.LastUpdated = clientInfo.LastUpdated
		existingClient.Status = clientInfo.Status

		err = db.UpdateClient(existingClient)
		if err != nil {
			return err
		}

		return nil
	}

	_, err = db.db.Exec("INSERT INTO clients (local_ip, system_info, last_updated, status) VALUES (?, ?, ?, ?)",
		clientInfo.LocalIP, clientInfo.SystemInfo, clientInfo.LastUpdated, clientInfo.Status)
	if err != nil {
		return err
	}

	return nil
}

// GetClientByLocalIP 根据 local_ip 查询客户信息
func (db *SQLiteDB) GetClientByLocalIP(localIP string) (*client.ClientInfo, error) {
	row := db.db.QueryRow("SELECT local_ip, system_info, last_updated, status FROM clients WHERE local_ip = ?", localIP)

	var clientInfo client.ClientInfo
	err := row.Scan(&clientInfo.LocalIP, &clientInfo.SystemInfo, &clientInfo.LastUpdated, &clientInfo.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到记录
		}
		return nil, err
	}

	return &clientInfo, nil
}

// UpdateClient 更新客户信息
func (db *SQLiteDB) UpdateClient(clientInfo *client.ClientInfo) error {
	_, err := db.db.Exec("UPDATE clients SET system_info = ?, last_updated = ?, status = ? WHERE local_ip = ?",
		clientInfo.SystemInfo, clientInfo.LastUpdated, clientInfo.Status, clientInfo.LocalIP)
	if err != nil {
		return err
	}

	return nil
}

// Close 关闭数据库连接
func (db *SQLiteDB) Close() {
	err := db.db.Close()
	if err != nil {
		log.Println("Error closing database:", err)
	}
}

// GetClients 从数据库中获取所有客户端信息
func (db *SQLiteDB) GetClients() ([]client.ClientInfo, error) {
	query := "SELECT local_ip, system_info, last_updated, status FROM clients"

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientList []client.ClientInfo

	for rows.Next() {
		var localIP, systemInfo, lastUpdatedStr, status string

		err := rows.Scan(&localIP, &systemInfo, &lastUpdatedStr, &status)
		if err != nil {
			return nil, err
		}

		//lastUpdated, err := time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
		if err != nil {
			return nil, err
		}

		clientInfo := client.ClientInfo{
			LocalIP:     localIP,
			SystemInfo:  systemInfo,
			LastUpdated: lastUpdatedStr,
			Status:      status,
		}

		clientList = append(clientList, clientInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clientList, nil
}
