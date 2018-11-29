package database

import (
	"github.com/tidwall/buntdb"
)

// Model 层结构定义
type Model struct {
	// ApppID 来自微信, 同时也作为 model 的 index
	AppID string

	db *buntdb.DB
}

// AllModel 装载了当前系统所有可用的微信 APP 配置
var AllModel = make(map[string]*Model)

// NewModel 初始化 model
func NewModel(appid string) (*Model, error) {
	if m, ok := AllModel[appid]; ok {
		return m, nil
	}

	db, err := NewDB(appid)
	if err != nil {
		return nil, err
	}

	m := &Model{
		AppID: appid,
		db:    db,
	}

	AllModel[appid] = m
	return m, nil
}

// Query 依据 key 查询对应的 value
func (m *Model) Query(key string) (result string, err error) {
	err = m.db.View(func(tx *buntdb.Tx) error {
		result, err = tx.Get(key)

		return err
	})

	return
}

// Update 对 key 对应的 val 进行更新, 有更新，无插入
func (m *Model) Update(key, val string) error {
	err := m.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, val, nil)
		return err
	})

	return err
}

// GetDB 获取 Model 对应的 DB
func (m *Model) GetDB() *buntdb.DB {
	return m.db
}
