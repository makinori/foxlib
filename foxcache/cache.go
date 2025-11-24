package foxcache

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"go.etcd.io/bbolt"
)

var (
	// to be set in the init function
	currentDB     *bbolt.DB
	currentBucket []byte
)

type cacheData[T any] struct {
	Data    T         `json:"data"`
	Updated time.Time `json:"updated"`
	Expires time.Time `json:"expires"`
}

func setCache[T any](key string, data cacheData[T]) error {
	if currentDB == nil {
		return errors.New("database not set")
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return currentDB.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(currentBucket)
		if bucket == nil {
			return errors.New("bucket not found")
		}

		return bucket.Put([]byte(key), jsonBytes)
	})
}

func getCache[T any](key string) (*cacheData[T], error) {
	if currentDB == nil {
		return nil, errors.New("database not set")
	}

	var bytes []byte

	err := currentDB.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(currentBucket)
		if bucket == nil {
			return errors.New("bucket not found")
		}

		found := bucket.Get([]byte(key))
		bytes = make([]byte, len(found))
		copy(bytes, found)

		return nil
	})

	if err != nil {
		return nil, err
	}

	var cacheData cacheData[T]

	err = json.Unmarshal(bytes, &cacheData)
	if err != nil {
		return nil, err
	}

	if time.Now().After(cacheData.Expires) {
		os.Remove("cache/" + key + ".json")
		return nil, errors.New("cache data expired")
	}

	return &cacheData, nil
}
