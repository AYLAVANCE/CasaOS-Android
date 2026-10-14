/*
 * @Author: LinkLeong link@icewhale.com
 * @Date: 2022-05-13 18:15:46
 * @LastEditors: LinkLeong
 * @LastEditTime: 2022-09-05 11:58:02
 * @FilePath: /CasaOS/pkg/config/init.go
 * @Description:
 * @Website: https://www.casaos.io
 * Copyright (c) 2022 by icewhale, All Rights Reserved.
 */
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/IceWhaleTech/CasaOS/common"
	"github.com/IceWhaleTech/CasaOS/model"
	"github.com/IceWhaleTech/CasaOS/pkg/platform"
	"github.com/go-ini/ini"
)

var (
	SysInfo = &model.SysInfoModel{}
	AppInfo = &model.APPModel{
		DBPath:       constants.DefaultDataPath,
		LogPath:      constants.DefaultLogPath,
		LogSaveName:  common.SERVICENAME,
		LogFileExt:   "log",
		ShellPath:    "/usr/share/casaos/shell",
		UserDataPath: filepath.Join(constants.DefaultDataPath, "conf"),
	}
	CommonInfo = &model.CommonModel{
		RuntimePath: constants.DefaultRuntimePath,
	}
	ServerInfo       = &model.ServerModel{}
	SystemConfigInfo = &model.SystemConfig{}
	FileSettingInfo  = &model.FileSetting{}

	Cfg            *ini.File
	ConfigFilePath string
)

// 初始化设置，获取系统的部分信息。
func InitSetup(config string, sample string) {
	ConfigFilePath = CasaOSConfigFilePath
	if len(config) > 0 {
		ConfigFilePath = config
	}

	// On Android or if pointing to root /etc/, adapt ConfigFilePath to Termux path
	if platform.IsAndroid() || strings.HasPrefix(ConfigFilePath, "/etc/") {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		prefix := os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
		prefixConf := filepath.Join(prefix, "etc", "casaos", "casaos.conf")
		homeConf := filepath.Join(home, ".casaos", "conf", "casaos.conf")

		if _, err := os.Stat(prefixConf); err == nil {
			ConfigFilePath = prefixConf
		} else if _, err := os.Stat(homeConf); err == nil {
			ConfigFilePath = homeConf
		} else if len(config) == 0 {
			if err := os.MkdirAll(filepath.Dir(prefixConf), 0755); err == nil {
				ConfigFilePath = prefixConf
			} else {
				ConfigFilePath = homeConf
			}
		}
	}

	// Ensure parent directory of ConfigFilePath exists
	if err := os.MkdirAll(filepath.Dir(ConfigFilePath), 0755); err != nil {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		ConfigFilePath = filepath.Join(home, ".casaos", "conf", "casaos.conf")
		_ = os.MkdirAll(filepath.Dir(ConfigFilePath), 0755)
	}

	// create default config file if not exist
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		fmt.Println("config file not exist, create it")

		// If Android, adapt sample config paths before writing
		if platform.IsAndroid() {
			home := os.Getenv("HOME")
			if home == "" {
				home = "/data/data/com.termux/files/home"
			}
			prefix := os.Getenv("PREFIX")
			if prefix == "" {
				prefix = "/data/data/com.termux/files/usr"
			}
			casaosDir := filepath.Join(home, ".casaos")
			sample = strings.ReplaceAll(sample, "/var/log/casaos/", filepath.Join(casaosDir, "log")+"/")
			sample = strings.ReplaceAll(sample, "/var/lib/casaos/conf", filepath.Join(casaosDir, "conf"))
			sample = strings.ReplaceAll(sample, "/var/lib/casaos", casaosDir)
			sample = strings.ReplaceAll(sample, "/usr/share/casaos/shell", filepath.Join(prefix, "share", "casaos", "shell"))
		}

		// create config file
		file, err := os.Create(ConfigFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// write default config
		_, err = file.WriteString(sample)
		if err != nil {
			panic(err)
		}
	}

	var err error

	// 读取文件
	Cfg, err = ini.Load(ConfigFilePath)
	if err != nil {
		panic(err)
	}

	mapTo("app", AppInfo)
	mapTo("server", ServerInfo)
	mapTo("system", SystemConfigInfo)
	mapTo("file", FileSettingInfo)
	mapTo("common", CommonInfo)

	// Post-mapping sanitization: ensure Linux root paths are overridden on Android
	if platform.IsAndroid() || strings.HasPrefix(AppInfo.DBPath, "/var/") || strings.HasPrefix(AppInfo.LogPath, "/var/") {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		prefix := os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
		casaosDir := filepath.Join(home, ".casaos")

		AppInfo.DBPath = casaosDir
		AppInfo.LogPath = filepath.Join(casaosDir, "log")
		AppInfo.UserDataPath = filepath.Join(casaosDir, "conf")
		AppInfo.ShellPath = filepath.Join(prefix, "share", "casaos", "shell")
		CommonInfo.RuntimePath = filepath.Join(casaosDir, "run")
	}

	// Pre-create all application directories
	_ = os.MkdirAll(AppInfo.DBPath, 0755)
	_ = os.MkdirAll(filepath.Join(AppInfo.DBPath, "db"), 0755)
	_ = os.MkdirAll(AppInfo.LogPath, 0755)
	_ = os.MkdirAll(AppInfo.UserDataPath, 0755)
	_ = os.MkdirAll(CommonInfo.RuntimePath, 0755)
}

// 映射
func mapTo(section string, v interface{}) {
	err := Cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}
