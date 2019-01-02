package storage

import (
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb"
	"os/user"
	"github.com/symphonyprotocol/log"
)

var (
	storageLogger = log.GetLogger("storage")
	CURRENT_USER, _ = user.Current()
	LEVEL_DB_FILE = CURRENT_USER.HomeDir + "/.symchaindb"
)

func GetData(key string) ([]byte, error) {
	db, err := leveldb.OpenFile(LEVEL_DB_FILE, nil)
	if err != nil {
		storageLogger.Error("cannot open leveldb:", err)
	}
	defer db.Close()
	data, err := db.Get([]byte(key), nil)
	if err != nil {
		storageLogger.Error("GetKey error:%v\n", key, err)
		return nil, err
	}
	return data, err
}

func GetInt(key string) (int64, error) {
	data, err := GetData(key)
	if err == nil {
		res, _ := binary.Varint(data)
		return res, nil
	}
	return -1, err
}

func SaveInt(key string, value int64) error {
	buf := make([]byte, 200, 200)
	len := binary.PutVarint(buf, value)
	return SaveData(key, buf[:len])
}

func SaveData(key string, value []byte) error {
	db, err := leveldb.OpenFile(LEVEL_DB_FILE, nil)
	if err != nil {
		storageLogger.Error("cannot open leveldb:", err)
	}
	defer db.Close()
	err = db.Put([]byte(key), value, nil)
	if err != nil {
		storageLogger.Error("SetKey %v error:%v\n", key, err)
		return err
	}
	return nil
}

