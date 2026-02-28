package service

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"lingosql/internal/dao/sqlite"
	"lingosql/internal/models"
	"lingosql/internal/utils"
	"lingosql/pkg/db"
)

type ExportService struct {
	connectionDAO *sqlite.ConnectionDAO
}

func NewExportService(connectionDAO *sqlite.ConnectionDAO) *ExportService {
	return &ExportService{connectionDAO: connectionDAO}
}

func (s *ExportService) getExecutor(connectionID, userID int, database string) (db.Executor, *models.Connection, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, nil, err
	}
	if conn.UserID != userID {
		return nil, nil, errors.New("无权访问此连接")
	}
	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, nil, errors.New("配置解析失败")
	}
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, nil, err
	}
	executor, err := db.GetPool().GetExecutor(
		connectionID, conn.DBType, dbConfig.Host, dbConfig.Port,
		database, dbConfig.Username, password,
	)
	if err != nil {
		return nil, nil, err
	}
	return executor, conn, nil
}

func (s *ExportService) ExportData(userID int, req *models.ExportDataRequest, taskID int) (map[string]interface{}, error) {
	if err := utils.ValidateDatabaseName(req.Database); err != nil {
		return nil, err
	}
	if err := utils.ValidateTableName(req.Table); err != nil {
		return nil, err
	}

	format := req.Format
	if format == "" {
		format = "csv"
	}
	maxRows := req.MaxRows
	if maxRows <= 0 {
		maxRows = 10000
	}

	executor, conn, err := s.getExecutor(req.ConnectionID, userID, req.Database)
	if err != nil {
		return nil, err
	}

	queryTable := req.Table
	if conn.DBType == "postgresql" {
		queryTable = fmt.Sprintf("\"%s\"", req.Table)
	} else {
		queryTable = fmt.Sprintf("`%s`", req.Table)
	}

	sql := fmt.Sprintf("SELECT * FROM %s LIMIT %d", queryTable, maxRows)
	columns, rows, _, err := executor.Execute(sql)
	if err != nil {
		return nil, err
	}

	exportDir := "./data/exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("export_%d_%d.%s", taskID, time.Now().Unix(), format)
	filePath := filepath.Join(exportDir, fileName)

	if format == "json" {
		payload := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			item := make(map[string]interface{}, len(columns))
			for i, col := range columns {
				if i < len(row) {
					item[col] = row[i]
				}
			}
			payload = append(payload, item)
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return nil, err
		}
	} else {
		file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		if err := writer.Write(columns); err != nil {
			return nil, err
		}
		for _, row := range rows {
			record := make([]string, len(columns))
			for i := range columns {
				if i < len(row) && row[i] != nil {
					record[i] = fmt.Sprintf("%v", row[i])
				}
			}
			if err := writer.Write(record); err != nil {
				return nil, err
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"file_path": filePath,
		"file_name": fileName,
		"format":    format,
		"rows":      len(rows),
	}, nil
}
