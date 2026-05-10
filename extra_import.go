package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type colType int

const (
	colBool colType = iota
	colInt
	colFloat
	colText
)

func ImportExtraJSONToMySQL() {
	dsn := os.Getenv("STZB_MYSQL_DSN")
	if strings.TrimSpace(dsn) == "" {
		dsn = "root:123456@tcp(127.0.0.1:3306)/stzb?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("[extra-import] MySQL 连接失败: %v", err)
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Printf("[extra-import] MySQL 不可用: %v", err)
		return
	}

	tasks := []struct {
		filePath  string
		tableName string
	}{
		{"skill_extra.json", "skill_extra"},
		{"hero_extra.json", "hero_extra"},
		{"gear_extra.json", "gear_extra"},
	}

	for _, t := range tasks {
		if err := importOneJSONFile(db, t.filePath, t.tableName); err != nil {
			log.Printf("[extra-import] 导入失败 %s -> %s: %v", t.filePath, t.tableName, err)
			continue
		}
		log.Printf("[extra-import] 导入成功 %s -> %s", t.filePath, t.tableName)
	}
}

func importOneJSONFile(db *sql.DB, filePath, tableName string) error {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var rows []map[string]interface{}
	if err = json.Unmarshal(b, &rows); err != nil {
		return fmt.Errorf("JSON 解析失败: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	cols, types := inferColumns(rows)
	if len(cols) == 0 {
		return nil
	}

	if err = createTableIfNeeded(db, tableName, cols, types); err != nil {
		return err
	}

	if _, err = db.Exec("DELETE FROM `" + tableName + "`"); err != nil {
		return fmt.Errorf("清空表失败: %w", err)
	}

	return batchInsert(db, tableName, cols, types, rows)
}

func inferColumns(rows []map[string]interface{}) ([]string, map[string]colType) {
	types := make(map[string]colType)
	for _, row := range rows {
		for k, v := range row {
			t, ok := types[k]
			if !ok {
				types[k] = detectType(v)
				continue
			}
			types[k] = mergeType(t, detectType(v))
		}
	}

	cols := make([]string, 0, len(types))
	for k := range types {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols, types
}

func detectType(v interface{}) colType {
	switch x := v.(type) {
	case nil:
		return colText
	case bool:
		return colBool
	case float64:
		if math.Abs(x-math.Round(x)) < 1e-9 {
			return colInt
		}
		return colFloat
	case string:
		return colText
	default:
		return colText
	}
}

func mergeType(a, b colType) colType {
	if a == colText || b == colText {
		return colText
	}
	if a == colFloat || b == colFloat {
		return colFloat
	}
	if a == colInt || b == colInt {
		return colInt
	}
	return colBool
}

func createTableIfNeeded(db *sql.DB, tableName string, cols []string, types map[string]colType) error {
	defs := make([]string, 0, len(cols)+1)
	for _, c := range cols {
		defs = append(defs, fmt.Sprintf("`%s` %s NULL", c, sqlType(types[c])))
	}
	if contains(cols, "id") {
		defs = append(defs, "PRIMARY KEY (`id`)")
	}

	sqlStr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", tableName, strings.Join(defs, ","))
	if _, err := db.Exec(sqlStr); err != nil {
		return fmt.Errorf("建表失败: %w", err)
	}

	// 查已有列，只补缺失的（ADD COLUMN IF NOT EXISTS 部分 MySQL 版本不支持）
	existingCols := map[string]bool{}
	rows, err := db.Query("SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", tableName)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var col string
			if rows.Scan(&col) == nil {
				existingCols[col] = true
			}
		}
	}
	for _, c := range cols {
		if existingCols[c] {
			continue
		}
		alter := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s NULL", tableName, c, sqlType(types[c]))
		if _, err := db.Exec(alter); err != nil {
			return fmt.Errorf("补列失败(%s): %w", c, err)
		}
	}
	return nil
}

func sqlType(t colType) string {
	switch t {
	case colBool:
		return "TINYINT(1)"
	case colInt:
		return "BIGINT"
	case colFloat:
		return "DOUBLE"
	default:
		return "LONGTEXT"
	}
}

func batchInsert(db *sql.DB, tableName string, cols []string, types map[string]colType, rows []map[string]interface{}) error {
	batchSize := 200
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		if err := insertChunk(db, tableName, cols, types, rows[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func insertChunk(db *sql.DB, tableName string, cols []string, types map[string]colType, rows []map[string]interface{}) error {
	placeOne := "(" + strings.TrimRight(strings.Repeat("?,", len(cols)), ",") + ")"
	places := make([]string, 0, len(rows))
	args := make([]interface{}, 0, len(rows)*len(cols))

	for _, r := range rows {
		places = append(places, placeOne)
		for _, c := range cols {
			args = append(args, castValue(r[c], types[c]))
		}
	}

	sqlStr := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES %s", tableName, strings.Join(cols, "`,`"), strings.Join(places, ","))
	if _, err := db.Exec(sqlStr, args...); err != nil {
		return fmt.Errorf("插入失败: %w", err)
	}
	return nil
}

func castValue(v interface{}, t colType) interface{} {
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	switch t {
	case colBool:
		if b, ok := v.(bool); ok {
			if b {
				return 1
			}
			return 0
		}
	case colInt:
		if f, ok := v.(float64); ok {
			return int64(math.Round(f))
		}
	case colFloat:
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return fmt.Sprintf("%v", v)
}

func contains(arr []string, target string) bool {
	for _, v := range arr {
		if v == target {
			return true
		}
	}
	return false
}
