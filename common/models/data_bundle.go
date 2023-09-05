package models

import "github.com/shopspring/decimal"

// 流量包
type DataBundle struct {
	BaseModel
	Name              string              `json:"name"`
	Price             decimal.Decimal     `json:"price"`
	DataBundleDetails []*DataBundleDetail `json:"data_bundle_details"`
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
