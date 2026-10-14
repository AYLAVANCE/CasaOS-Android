/*
 * @Author: LinkLeong link@icewhale.com
 * @Date: 2022-05-13 18:15:46
 * @LastEditors: LinkLeong
 * @LastEditTime: 2022-08-31 13:39:24
 * @FilePath: /CasaOS/pkg/sqlite/db.go
 * @Description:
 * @Website: https://www.casaos.io
 * Copyright (c) 2022 by icewhale, All Rights Reserved.
 */
package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS/pkg/platform"
	model2 "github.com/IceWhaleTech/CasaOS/service/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var gdb *gorm.DB

func GetDb(dbPath string) *gorm.DB {
	if gdb != nil {
		return gdb
	}

	// Termux / Android path sanitization
	if platform.IsAndroid() || strings.HasPrefix(dbPath, "/var/") || strings.HasPrefix(dbPath, "/usr/") {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		dbPath = filepath.Join(home, ".casaos", "db")
	}

	// Clean up duplicate /db/db if present
	if strings.HasSuffix(dbPath, filepath.Join("db", "db")) {
		dbPath = filepath.Dir(dbPath)
	}

	// Ensure database directory exists with proper permissions
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		dbPath = filepath.Join(home, ".casaos", "db")
		_ = os.MkdirAll(dbPath, 0755)
	}

	dbFile := filepath.Join(dbPath, "casaOS.db")
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		fmt.Printf("[error] failed to initialize database at %s: %v\n", dbFile, err)
		panic(fmt.Sprintf("sqlite connect error: %v", err))
	}

	c, _ := db.DB()
	c.SetMaxIdleConns(10)
	c.SetMaxOpenConns(1)
	c.SetConnMaxIdleTime(time.Second * 1000)
	gdb = db

	err = db.AutoMigrate(&model2.AppNotify{}, model2.SharesDBModel{}, model2.ConnectionsDBModel{}, model2.PeerDriveDBModel{})
	if err != nil {
		fmt.Println(err)
	}

	db.Exec("DROP TABLE IF EXISTS o_application")
	db.Exec("DROP TABLE IF EXISTS o_friend")
	db.Exec("DROP TABLE IF EXISTS o_person_download")
	db.Exec("DROP TABLE IF EXISTS o_person_down_record")
	return db
}
