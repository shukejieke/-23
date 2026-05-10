package model

import (
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Conn *gorm.DB

func InitDB(_ string) {
	dsn := os.Getenv("STZB_MYSQL_DSN")
	if strings.TrimSpace(dsn) == "" {
		dsn = "root:123456@tcp(127.0.0.1:3306)/stzb?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect mysql database")
	}

	err = db.AutoMigrate(&TeamUser{}, &Task{}, &Report{}, &BattleReport{}, &CmdSchema{}, &CmdFieldSchema{}, &UnionLeaderboard{}, &PersonalLeaderboard{}, &PlayerTerritoryRank{}, &UnionGroupMeta{}, &RawCapture{})
	if err != nil {
		return
	}

	Conn = db
}
