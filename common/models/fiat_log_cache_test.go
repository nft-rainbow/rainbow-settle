package models

import (
	"os"
	"testing"
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	ConnectDB(config.Mysql{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "new-password",
		Db:       "rainbowtest",
	})
}

func _TestMergeApiFeeFiatlog(t *testing.T) {
	err := GetDB().Transaction(func(tx *gorm.DB) error {

		if err := mergePayApiFeeFiatlogs(tx, time.Now().AddDate(0, 0, -1), time.Now()); err != nil {
			return err
		}

		if err := mergeRefundApiFeeFiatlogs(tx, time.Now().AddDate(0, 0, -1), time.Now()); err != nil {
			return err
		}

		return nil

	})
	assert.NoError(t, err)
}

func _TestMergeApiQuotaFiatlog(t *testing.T) {
	err := GetDB().Transaction(func(tx *gorm.DB) error {
		return mergeApiQuotaFiatlogs(tx, time.Now().AddDate(0, 0, -1), time.Now())
	})

	assert.NoError(t, err)
}
