package service

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"lingosql/internal/models"
	"lingosql/internal/utils"
)

func (s *MaintenanceService) runMySQLBackup(dbConfig *models.DbConfig, password, backupDir, database string, tables []string, compress, schemaOnly bool, maxFileSizeMB int, taskID int, taskService *TaskService) ([]struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}, int64, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, 0, err
	}
	for _, t := range tables {
		if t != "" && utils.ValidateTableName(t) != nil {
			return nil, 0, fmt.Errorf("非法表名: %s", t)
		}
	}

	charset := "utf8mb4"
	if dbConfig.Options != nil && dbConfig.Options.Charset != "" {
		charset = dbConfig.Options.Charset
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.Username, password, dbConfig.Host, dbConfig.Port, database, charset)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, 0, fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, 0, fmt.Errorf("MySQL 连接测试失败: %w", err)
	}

	var tableList []string
	if len(tables) > 0 {
		for _, t := range tables {
			if t != "" {
				tableList = append(tableList, t)
			}
		}
	}
	if len(tableList) == 0 {
		rows, err := db.Query("SHOW TABLES")
		if err != nil {
			return nil, 0, fmt.Errorf("获取表列表失败: %w", err)
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				rows.Close()
				return nil, 0, err
			}
			tableList = append(tableList, name)
		}
		rows.Close()
	}

	var files []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	var totalSize int64
	maxBytes := int64(0)
	if maxFileSizeMB > 0 {
		maxBytes = int64(maxFileSizeMB) * 1024 * 1024
	}
	part := 1
	ext := ".sql"
	if compress {
		ext = ".sql.gz"
	}

	openNextFile := func() (io.Writer, *os.File, *gzip.Writer, string, error) {
		name := fmt.Sprintf("part_%03d%s", part, ext)
		path := filepath.Join(backupDir, name)
		f, err := os.Create(path)
		if err != nil {
			return nil, nil, nil, "", err
		}
		var w io.Writer = f
		var gw *gzip.Writer
		if compress {
			gw = gzip.NewWriter(f)
			w = gw
		}
		return w, f, gw, name, nil
	}

	writeAndRotate := func(buf *strings.Builder, currentW io.Writer, currentF *os.File, currentGW *gzip.Writer, currentName string, currentSize *int64) (io.Writer, *os.File, *gzip.Writer, string, error) {
		if buf.Len() == 0 {
			return currentW, currentF, currentGW, currentName, nil
		}
		line := buf.String()
		buf.Reset()
		if currentF == nil {
			w, f, gw, name, err := openNextFile()
			if err != nil {
				return nil, nil, nil, "", err
			}
			n, _ := io.WriteString(w, line)
			*currentSize += int64(n)
			return w, f, gw, name, nil
		}
		n, err := io.WriteString(currentW, line)
		if err != nil {
			return nil, nil, nil, "", err
		}
		*currentSize += int64(n)
		if maxBytes > 0 && *currentSize >= maxBytes {
			if currentGW != nil {
				currentGW.Close()
			}
			currentF.Close()
			files = append(files, struct {
				Name string `json:"name"`
				Size int64  `json:"size"`
			}{currentName, *currentSize})
			totalSize += *currentSize
			*currentSize = 0
			part++
			w, f, gw, name, err := openNextFile()
			if err != nil {
				return nil, nil, nil, "", err
			}
			return w, f, gw, name, nil
		}
		return currentW, currentF, currentGW, currentName, nil
	}

	var currentW io.Writer
	var currentF *os.File
	var currentGW *gzip.Writer
	var currentName string
	var currentSize int64
	var buf strings.Builder

	emit := func(s string) error {
		buf.WriteString(s)
		return nil
	}

	for _, table := range tableList {
		if err := utils.ValidateTableName(table); err != nil {
			continue
		}
		var ddl string
		err := db.QueryRow("SHOW CREATE TABLE `" + table + "`").Scan(&table, &ddl)
		if err != nil {
			return nil, 0, fmt.Errorf("获取表 %s 结构失败: %w", table, err)
		}
		emit("-- LingoSQL native backup: table " + table + "\n")
		emit("DROP TABLE IF EXISTS `" + table + "`;\n")
		emit(ddl + ";\n\n")
		var err2 error
		currentW, currentF, currentGW, currentName, err2 = writeAndRotate(&buf, currentW, currentF, currentGW, currentName, &currentSize)
		if err2 != nil {
			return nil, 0, err2
		}

		if schemaOnly {
			continue
		}

		rows, err := db.Query("SELECT * FROM `" + table + "`")
		if err != nil {
			return nil, 0, fmt.Errorf("查询表 %s 失败: %w", table, err)
		}
		cols, _ := rows.Columns()
		colCount := len(cols)
		if colCount == 0 {
			rows.Close()
			continue
		}

		colList := "`" + strings.Join(cols, "`,`") + "`"
		batchSize := 100
		values := make([]interface{}, colCount)
		valuePtrs := make([]interface{}, colCount)
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		var rowBatch []string
		for rows.Next() {
			if err := rows.Scan(valuePtrs...); err != nil {
				rows.Close()
				return nil, 0, err
			}
			var parts []string
			for _, v := range values {
				parts = append(parts, escapeMySQLValue(v))
			}
			rowBatch = append(rowBatch, "("+strings.Join(parts, ",")+")")
			if len(rowBatch) >= batchSize {
				emit("INSERT INTO `" + table + "` (" + colList + ") VALUES " + strings.Join(rowBatch, ",") + ";\n")
				rowBatch = rowBatch[:0]
				var err error
				currentW, currentF, currentGW, currentName, err = writeAndRotate(&buf, currentW, currentF, currentGW, currentName, &currentSize)
				if err != nil {
					rows.Close()
					return nil, 0, err
				}
			}
		}
		rows.Close()
		if len(rowBatch) > 0 {
			emit("INSERT INTO `" + table + "` (" + colList + ") VALUES " + strings.Join(rowBatch, ",") + ";\n")
		}
		emit("\n")
	}

	if buf.Len() > 0 {
		if currentF == nil {
			w, f, gw, name, err := openNextFile()
			if err != nil {
				return nil, 0, err
			}
			content := buf.String()
			io.Copy(w, strings.NewReader(content))
			if gw != nil {
				gw.Close()
			}
			f.Close()
			sz := int64(len(content))
			files = append(files, struct {
				Name string `json:"name"`
				Size int64  `json:"size"`
			}{name, sz})
			totalSize += sz
		} else {
			io.WriteString(currentW, buf.String())
			currentSize += int64(buf.Len())
			if currentGW != nil {
				currentGW.Close()
			}
			currentF.Close()
			files = append(files, struct {
				Name string `json:"name"`
				Size int64  `json:"size"`
			}{currentName, currentSize})
			totalSize += currentSize
		}
	}

	if len(files) == 0 {
		return nil, 0, fmt.Errorf("备份未生成任何文件")
	}
	return files, totalSize, nil
}

func escapeMySQLValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		return "'" + escapeMySQLString(string(val)) + "'"
	case string:
		return "'" + escapeMySQLString(val) + "'"
	case time.Time:
		return "'" + val.Format("2006-01-02 15:04:05") + "'"
	case bool:
		if val {
			return "1"
		}
		return "0"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		return "'" + escapeMySQLString(fmt.Sprintf("%v", v)) + "'"
	}
}

func escapeMySQLString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case 0:
			b.WriteString(`\0`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
