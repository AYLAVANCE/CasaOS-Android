/*
 * @Author: LinkLeong link@icewhale.com
 * @Date: 2021-09-30 18:18:14
 * @LastEditors: LinkLeong
 * @LastEditTime: 2022-08-31 17:04:02
 * @FilePath: /CasaOS/pkg/config/config.go
 * @Description:
 * @Website: https://www.casaos.io
 * Copyright (c) 2022 by icewhale, All Rights Reserved.
 */
package config

import (
	"os"
	"path/filepath"

	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/IceWhaleTech/CasaOS/pkg/platform"
)

var CasaOSConfigFilePath = filepath.Join(constants.DefaultConfigPath, "casaos.conf")

func init() {
	if platform.IsAndroid() {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		prefix := os.Getenv("PREFIX")
		if prefix == "" {
			prefix = "/data/data/com.termux/files/usr"
		}
		prefixConf := filepath.Join(prefix, "etc", "casaos", "casaos.conf")
		if _, err := os.Stat(prefixConf); err == nil {
			CasaOSConfigFilePath = prefixConf
		} else {
			CasaOSConfigFilePath = filepath.Join(home, ".casaos", "conf", "casaos.conf")
		}
	}
}
