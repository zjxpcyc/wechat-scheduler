package database

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zjxpcyc/wechat-scheduler/lib"

	"github.com/tidwall/buntdb"
)

// DBDir 存放数据库文件
const DBDir = "./database"

var logger lib.LogService

// NewDB 初始化数据库引擎
func NewDB(appid string) (*buntdb.DB, error) {
	if m, ok := AllModel[appid]; ok {
		return m.GetDB(), nil
	}

	p := DBDir + "/" + appid + ".db"
	return buntdb.Open(p)
}

// Init 初始启动
// 遍历当前 database 目录下面所有 .db 结尾的文件
func Init() error {
	logger = lib.GetLogger()
	logger.Info("开始进行 Model 初始化 ...")

	// 目录不存在, 则创建
	if _, err := os.Stat(DBDir); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(DBDir, 0700); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// 遍历目录下所有 .db 文件
	return filepath.Walk(DBDir, func(f string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(f)
		if ext == ".db" {
			_, fname := filepath.Split(filepath.ToSlash(f))
			appid := strings.TrimSuffix(fname, ext)
			logger.Info(fname, appid)

			if _, err := NewModel(appid); err != nil {
				return err
			}
		}

		return nil
	})
}
