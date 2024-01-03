package models

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
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
	// err := GetDB().Transaction(func(tx *gorm.DB) error {

	if err := mergePayApiFeeFiatlogs(time.Now().AddDate(0, 0, -1), time.Now()); err != nil {
		assert.NoError(t, err)
	}

	if err := mergeRefundApiFeeFiatlogs(time.Now().AddDate(0, 0, -1), time.Now()); err != nil {
		assert.NoError(t, err)
	}

	// return nil

	// })
	// assert.NoError(t, err)
}

func _TestMergeApiQuotaFiatlog(t *testing.T) {
	// err := GetDB().Transaction(func(tx *gorm.DB) error {
	err := mergeApiQuotaFiatlogs(time.Now().AddDate(0, 0, -1), time.Now())
	// })

	assert.NoError(t, err)
}

func _TestRelateBuySponsorFiatlog(t *testing.T) {
	meta, _ := json.Marshal(FiatMetaRefundSponsor{
		RefundForFiatlogId:   4550,
		RefundForFiatlogType: FIAT_LOG_TYPE_BUY_STORAGE,
		TxId:                 259435,
		Reason:               "no",
	})
	err := GetDB().Save(&FiatLogCache{
		FiatLogCore: FiatLogCore{
			UserId:  1,
			Type:    FIAT_LOG_TYPE_REFUND_SPONSOR,
			Amount:  decimal.NewFromInt(1),
			Meta:    meta,
			OrderNO: RandomOrderNO(),
		},
	}).Error
	assert.NoError(t, err)
}
