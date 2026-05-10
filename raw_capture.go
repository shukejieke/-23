package model

import "time"

type RawCapture struct {
	ID           int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CmdID        int    `json:"cmd_id" gorm:"column:cmd_id;index"`
	CmdName      string `json:"cmd_name" gorm:"column:cmd_name"`
	DataType     int    `json:"data_type" gorm:"column:data_type"`
	Src          string `json:"src" gorm:"column:src;index"`
	Dst          string `json:"dst" gorm:"column:dst;index"`
	PayloadLen   int    `json:"payload_len" gorm:"column:payload_len"`
	PacketBufLen int    `json:"packet_buf_len" gorm:"column:packet_buf_len"`
	Preview      string `json:"preview" gorm:"column:preview;type:TEXT"`
	RawHex       string `json:"raw_hex" gorm:"column:raw_hex;type:LONGTEXT"`
	DecodedText  string `json:"decoded_text" gorm:"column:decoded_text;type:LONGTEXT"`
	CaptureTime  int64  `json:"capture_time" gorm:"column:capture_time;index"`
}

func (RawCapture) TableName() string { return "raw_capture" }

func SaveRawCapture(rec RawCapture) {
	if Conn == nil {
		return
	}
	if rec.CaptureTime == 0 {
		rec.CaptureTime = time.Now().Unix()
	}
	_ = Conn.Create(&rec).Error
}
