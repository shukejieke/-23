package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CmdSchema struct {
	CmdID        int    `json:"cmd_id" gorm:"column:cmd_id;primaryKey"`
	CmdName      string `json:"cmd_name" gorm:"column:cmd_name"`
	DataType     int    `json:"data_type" gorm:"column:data_type"`
	SampleText   string `json:"sample_text" gorm:"column:sample_text;type:TEXT"`
	RawSampleHex string `json:"raw_sample_hex" gorm:"column:raw_sample_hex;type:LONGTEXT"`
	SeenCount    int64  `json:"seen_count" gorm:"column:seen_count"`
	FirstSeen    int64  `json:"first_seen" gorm:"column:first_seen"`
	UpdatedAt    int64  `json:"updated_at" gorm:"column:updated_at"`
}

func (CmdSchema) TableName() string {
	return "cmd_schema"
}

type CmdFieldSchema struct {
	ID          int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CmdID       int    `json:"cmd_id" gorm:"column:cmd_id;index:idx_cmd_field,unique"`
	FieldPath   string `json:"field_path" gorm:"column:field_path;index:idx_cmd_field,unique"`
	FieldType   string `json:"field_type" gorm:"column:field_type"`
	SampleValue string `json:"sample_value" gorm:"column:sample_value;type:TEXT"`
	HitCount    int64  `json:"hit_count" gorm:"column:hit_count"`
	UpdatedAt   int64  `json:"updated_at" gorm:"column:updated_at"`
}

func (CmdFieldSchema) TableName() string {
	return "cmd_field_schema"
}

func SaveCmdSchema(cmdID int, cmdName string, dataType int, decoded string, rawHex string) {
	if Conn == nil || cmdID <= 0 {
		return
	}
	now := time.Now().Unix()
	sample := decoded
	if len(sample) > 2000 {
		sample = sample[:2000] + "..."
	}
	if len(rawHex) > 8000 {
		rawHex = rawHex[:8000] + "..."
	}

	_ = Conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cmd_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			// 若当前上报为“未映射”，保留已确认的人工语义，避免被覆盖回退。
			"cmd_name":       gorm.Expr("IF(cmd_name IS NULL OR cmd_name = '' OR cmd_name = '未映射', VALUES(cmd_name), cmd_name)"),
			"data_type":      dataType,
			"sample_text":    sample,
			"raw_sample_hex": rawHex,
			"updated_at":     now,
		}),
	}).Create(&CmdSchema{CmdID: cmdID, CmdName: cmdName, DataType: dataType, SampleText: sample, RawSampleHex: rawHex, SeenCount: 1, FirstSeen: now, UpdatedAt: now}).Error

	_ = Conn.Exec("UPDATE cmd_schema SET seen_count = seen_count + 1 WHERE cmd_id = ?", cmdID).Error

	var obj any
	if err := json.Unmarshal([]byte(decoded), &obj); err != nil {
		return
	}
	fields := map[string]CmdFieldSchema{}
	flattenJSON("$", obj, fields)
	for _, f := range fields {
		f.CmdID = cmdID
		f.UpdatedAt = now
		_ = Conn.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "cmd_id"}, {Name: "field_path"}},
			DoUpdates: clause.Assignments(map[string]any{
				"field_type":   f.FieldType,
				"sample_value": f.SampleValue,
				"updated_at":   now,
			}),
		}).Create(&f).Error
		_ = Conn.Exec("UPDATE cmd_field_schema SET hit_count = hit_count + 1 WHERE cmd_id = ? AND field_path = ?", cmdID, f.FieldPath).Error
	}
}

func ConfirmCmdName(cmdID int, cmdName string) error {
	if Conn == nil || cmdID <= 0 || cmdName == "" {
		return nil
	}
	now := time.Now().Unix()
	return Conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cmd_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"cmd_name":   cmdName,
			"updated_at": now,
		}),
	}).Create(&CmdSchema{
		CmdID:      cmdID,
		CmdName:    cmdName,
		DataType:   0,
		SampleText: "",
		SeenCount:  0,
		FirstSeen:  now,
		UpdatedAt:  now,
	}).Error
}

func flattenJSON(path string, v any, out map[string]CmdFieldSchema) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			flattenJSON(path+"."+k, child, out)
		}
	case []any:
		for i, child := range val {
			flattenJSON(fmt.Sprintf("%s[%d]", path, i), child, out)
		}
	default:
		typeName := "null"
		sample := "null"
		switch t := val.(type) {
		case string:
			typeName = "string"
			sample = t
		case float64:
			typeName = "number"
			sample = fmt.Sprintf("%v", t)
		case bool:
			typeName = "bool"
			sample = fmt.Sprintf("%v", t)
		case nil:
			typeName = "null"
			sample = "null"
		default:
			typeName = fmt.Sprintf("%T", val)
			sample = fmt.Sprintf("%v", val)
		}
		if len(sample) > 300 {
			sample = sample[:300] + "..."
		}
		out[path] = CmdFieldSchema{FieldPath: path, FieldType: typeName, SampleValue: sample, HitCount: 1}
	}
}
