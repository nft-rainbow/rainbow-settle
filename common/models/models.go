package models

import (
	"fmt"
	"time"

	"github.com/nft-rainbow/rainbow-settle/common/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db          *gorm.DB
	mysqlConfig *config.Mysql
	fee         *config.Fee
	cfxPrice    float64
)

const (
	STATUS_INIT = iota
	STATUS_SUCCESS
	STATUS_FAIL
)

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (b BaseModel) GetID() uint { return b.ID }

type Count struct {
	Count int64 `json:"count"`
}

type ItemsWithCount[T any] struct {
	Count int `json:"count"`
	Items []T `json:"items"`
}

func NewItemsWithCount[T any](items []T) *ItemsWithCount[T] {
	return &ItemsWithCount[T]{
		Count: len(items),
		Items: items,
	}
}

func InitWithCreatedDB(_db *gorm.DB, _fee config.Fee, _cfxPrice float64) {
	fee = &_fee
	cfxPrice = _cfxPrice
	UseDB(_db)
	logrus.Info("connect db done")
	InitApiProfile()
	logrus.Info("init api profiles done")
	InitUserBalances()
	logrus.Info("init user balance done")
	InitUserSettleds()
	logrus.Info("init user settles done")
	InitBillPlan()
	logrus.Info("init bill plans done")
}

func Init(mysqlConfig config.Mysql, fee config.Fee, cfxPrice float64) {
	initConfigs(mysqlConfig, fee, cfxPrice)
	ConnectDB(mysqlConfig)
	logrus.Info("connect db done")
	InitApiProfile()
	logrus.Info("init api profiles done")
	InitUserBalances()
	logrus.Info("init user balance done")
	InitUserSettleds()
	logrus.Info("init user settles done")
	InitBillPlan()
	logrus.Info("init bill plans done")
}

func initConfigs(_mysqlConfig config.Mysql, _fee config.Fee, _cfxPrice float64) {
	mysqlConfig = &_mysqlConfig
	fee = &_fee
	cfxPrice = _cfxPrice
}

func UseDB(_db *gorm.DB) {
	db = _db
	err := MigrateSchemas()

	if err != nil {
		panic(err)
	}
}

func ConnectDB(dbConfig config.Mysql) {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	var err error
	// dbConfig := config.GetConfig().Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Db)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = MigrateSchemas()

	if err != nil {
		panic(err)
	}
}

func MigrateSchemas() error {
	// Migrate the schema
	return db.AutoMigrate(
		&User{},
		&ApiProfile{},
		&BillPlan{},
		&BillPlanDetail{},
		&DataBundle{},
		&DataBundleDetail{},
		&FiatLog{},
		&FiatLogCache{},
		&UserBalance{},
		&UserApiQuota{},
		&UserSettled{},
		&UserBillPlan{},
		&UserDataBundle{},
		&DepositOrder{},
		&CmbDepositNo{},
	)
}

func GetDB() *gorm.DB {
	return db
}

type IdReader interface {
	GetID() uint
}

func GetIds[T IdReader](items []T) []uint {
	var ids []uint
	for _, item := range items {
		ids = append(ids, item.GetID())
	}
	return ids
}
