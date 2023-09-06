package models

import (
	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/shopspring/decimal"
)

// 流量包
type DataBundle struct {
	BaseModel
	Name              string              `json:"name"`
	Price             decimal.Decimal     `json:"price"`
	DataBundleDetails []*DataBundleDetail `json:"data_bundle_details"`
}

type DataBundleFilter struct {
	ID uint `form:"id" json:"id"`
}

func QueryDataBundle(filter *DataBundleFilter, offset, limit int) (*ginutils.List[*DataBundle], error) {
	var aps []*DataBundle
	var count int64
	if err := GetDB().Model(&DataBundle{}).Where(&filter).Count(&count).Offset(offset).Limit(limit).Find(&aps).Error; err != nil {
		return nil, err
	}
	return &ginutils.List[*DataBundle]{Items: aps, Count: count}, nil
}

func GetDataBundleById(id uint) (*DataBundle, error) {
	var d *DataBundle
	if err := GetDB().Model(&DataBundle{}).Preload("DataBundleDetails").First(id, &d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

func GetAllDataBundles() ([]*DataBundle, error) {
	var ds []*DataBundle
	if err := GetDB().Model(&DataBundle{}).Preload("DataBundleDetails").Find(&ds).Error; err != nil {
		return nil, err
	}
	return ds, nil
}
