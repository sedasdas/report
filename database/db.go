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
		last_updated TEXT,
		status TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}

// InsertClientInfo 将客户信息插入数据库
func (db *SQLiteDB) InsertClientInfo(clientInfo client.ClientInfo) error {
	_, err := db.db.Exec("INSERT INTO clients (local_ip, system_info, last_updated, status) VALUES (?, ?, ?, ?)",
		clientInfo.LocalIP, clientInfo.SystemInfo, clientInfo.LastUpdated, clientInfo.Status)
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
