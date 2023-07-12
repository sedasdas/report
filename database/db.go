package database

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"report/client"
)

type SQLiteDB struct {
	db *sql.DB
}

func OpenSQLiteDB(dbFile string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	return &SQLiteDB{
		db: db,
	}, nil
}

func (db *SQLiteDB) CreateClientsTable() error {
	_, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS clients (
		local_ip TEXT PRIMARY KEY,
		system_info TEXT,
		disk_info TEXT,
		last_updated TIME,
		status TEXT
	)`)
	return err
}

func (db *SQLiteDB) InsertClientInfo(clientInfo client.ClientInfo) error {
	existingClient, err := db.GetClientByLocalIP(clientInfo.LocalIP)
	if err != nil {
		return err
	}
	if existingClient != nil {
		existingClient.SystemInfo = clientInfo.SystemInfo
		existingClient.DiskInfo = clientInfo.DiskInfo
		existingClient.LastUpdated = clientInfo.LastUpdated
		existingClient.Status = clientInfo.Status

		err = db.UpdateClient(existingClient)
		if err != nil {
			return err
		}

		return nil
	}

	diskInfoJSON, err := json.Marshal(clientInfo.DiskInfo)
	if err != nil {
		return err
	}

	_, err = db.db.Exec("INSERT INTO clients (local_ip, system_info, disk_info, last_updated, status) VALUES (?, ?, ?, ?, ?)",
		clientInfo.LocalIP, clientInfo.SystemInfo, string(diskInfoJSON), clientInfo.LastUpdated, clientInfo.Status)
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLiteDB) GetClientByLocalIP(localIP string) (*client.ClientInfo, error) {
	row := db.db.QueryRow("SELECT local_ip, system_info, disk_info, last_updated, status FROM clients WHERE local_ip = ?", localIP)

	var clientInfo client.ClientInfo
	var diskInfoJSON string
	err := row.Scan(&clientInfo.LocalIP, &clientInfo.SystemInfo, &diskInfoJSON, &clientInfo.LastUpdated, &clientInfo.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(diskInfoJSON), &clientInfo.DiskInfo)
	if err != nil {
		return nil, err
	}

	return &clientInfo, nil
}

func (db *SQLiteDB) UpdateClient(clientInfo *client.ClientInfo) error {
	diskInfoJSON, err := json.Marshal(clientInfo.DiskInfo)
	if err != nil {
		return err
	}

	_, err = db.db.Exec("UPDATE clients SET system_info = ?, disk_info = ?, last_updated = ?, status = ? WHERE local_ip = ?",
		clientInfo.SystemInfo, string(diskInfoJSON), clientInfo.LastUpdated, clientInfo.Status, clientInfo.LocalIP)
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLiteDB) Close() {
	err := db.db.Close()
	if err != nil {
		log.Println("Error closing database:", err)
	}
}

func (db *SQLiteDB) GetClients() ([]client.ClientInfo, error) {
	query := "SELECT local_ip, system_info, disk_info, last_updated, status FROM clients"

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clientList []client.ClientInfo

	for rows.Next() {
		var localIP, systemInfo, diskInfoJSON, lastUpdatedStr, status string
		err := rows.Scan(&localIP, &systemInfo, &diskInfoJSON, &lastUpdatedStr, &status)
		if err != nil {
			return nil, err
		}

		var diskInfo []client.Disk
		err = json.Unmarshal([]byte(diskInfoJSON), &diskInfo)
		if err != nil {
			return nil, err
		}

		clientInfo := client.ClientInfo{
			LocalIP:     localIP,
			SystemInfo:  systemInfo,
			DiskInfo:    diskInfo,
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
