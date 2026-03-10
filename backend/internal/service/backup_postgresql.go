package service

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"lingosql/internal/models"
	"lingosql/internal/utils"
)

func (s *MaintenanceService) runPostgreSQLBackup(dbConfig *models.DbConfig, password, backupDir, database string, tables []string, compress, schemaOnly bool, maxFileSizeMB int, taskID int, taskService *TaskService) ([]struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}, int64, error) {
	if err := utils.ValidateDatabaseName(database); err != nil {
		return nil, 0, err
	}
	if database == "" {
		database = "postgres"
	}
	for _, t := range tables {
		if t != "" && utils.ValidateTableName(t) != nil {
			return nil, 0, fmt.Errorf("非法表名: %s", t)
		}
	}

	sslMode := "disable"
	if dbConfig.Options != nil && dbConfig.Options.SslMode != "" {
		sslMode = dbConfig.Options.SslMode
	}
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(dbConfig.Username, password),
		Host:     fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port),
		Path:     "/" + database,
		RawQuery: "sslmode=" + url.QueryEscape(sslMode),
	}
	dsn := u.String()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, 0, fmt.Errorf("连接 PostgreSQL 失败: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, 0, fmt.Errorf("PostgreSQL 连接测试失败: %w", err)
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
		rows, err := db.Query(`
			SELECT tablename FROM pg_tables 
			WHERE schemaname = 'public' AND tablename !~ '^pg_' AND tablename != 'sql_features'
			ORDER BY tablename`)
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

	emit := func(s string) {
		buf.WriteString(s)
	}

	for _, table := range tableList {
		if err := utils.ValidateTableName(table); err != nil {
			continue
		}
		var ddl string
		err := db.QueryRow(`
			SELECT 'CREATE TABLE "public"."' || c.relname || '" (' ||
				string_agg(
					'"' || a.attname || '" ' || pg_catalog.format_type(a.atttypid, a.atttypmod) ||
					CASE WHEN a.attnotnull THEN ' NOT NULL' ELSE '' END,
					', ' ORDER BY a.attnum
				) || ');'
			FROM pg_catalog.pg_class c
			JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
			JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid
			WHERE c.relname = $1 AND n.nspname = 'public'
				AND a.attnum > 0 AND NOT a.attisdropped
			GROUP BY n.nspname, c.relname`, table).Scan(&ddl)
		if err != nil {
			return nil, 0, fmt.Errorf("获取表 %s 结构失败: %w", table, err)
		}
		emit("-- LingoSQL native backup: table " + table + "\n")
		emit(fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";\n", table))
		emit(ddl + "\n\n")
		var err2 error
		currentW, currentF, currentGW, currentName, err2 = writeAndRotate(&buf, currentW, currentF, currentGW, currentName, &currentSize)
		if err2 != nil {
			return nil, 0, err2
		}

		if schemaOnly {
			continue
		}

		rows, err := db.Query(fmt.Sprintf(`SELECT * FROM "public"."%s"`, table))
		if err != nil {
			return nil, 0, fmt.Errorf("查询表 %s 失败: %w", table, err)
		}
		cols, _ := rows.Columns()
		colCount := len(cols)
		if colCount == 0 {
			rows.Close()
			continue
		}

		quotedCols := make([]string, colCount)
		for i, c := range cols {
			quotedCols[i] = `"` + c + `"`
		}
		colList := strings.Join(quotedCols, ",")
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
				parts = append(parts, escapePGValue(v))
			}
			rowBatch = append(rowBatch, "("+strings.Join(parts, ",")+")")
			if len(rowBatch) >= batchSize {
				emit(fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES %s;\n", table, colList, strings.Join(rowBatch, ",")))
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
			emit(fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES %s;\n", table, colList, strings.Join(rowBatch, ",")))
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

func escapePGValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		return "'" + escapePGString(string(val)) + "'"
	case string:
		return "'" + escapePGString(val) + "'"
	case time.Time:
		return "'" + val.Format("2006-01-02 15:04:05") + "'"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		return "'" + escapePGString(fmt.Sprintf("%v", v)) + "'"
	}
}

func escapePGString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
