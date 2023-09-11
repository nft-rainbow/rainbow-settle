package models

import (
	"encoding/json"
	"errors"

	"github.com/nft-rainbow/conflux-gin-helper/utils/ginutils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"github.com/shopspring/decimal"
)

// 流量包
type DataBundle struct {
	BaseModel
	Name              string              `gorm:"unique" json:"name"`
	Price             decimal.Decimal     `json:"price"`
	DataBundleDetails []*DataBundleDetail `json:"data_bundle_details"`
}

func (u *DataBundle) MarshalJSON() ([]byte, error) {
	server, err := u.Server()
	if err != nil {
		return nil, err
	}
	type alias struct {
		DataBundle
		Server enums.ServerType `json:"server"`
	}
	return json.Marshal(alias{*u, server})
}

func (u *DataBundle) Server() (enums.ServerType, error) {
	if len(u.DataBundleDetails) == 0 {
		return 0, errors.New("no data bundle details")
	}
	profiles, err := GetApiProfiles()
	if err != nil {
		return 0, err
	}
	return profiles[u.DataBundleDetails[0].CostType].ServerType, nil
}

type DataBundleFilter struct {
	ID uint `form:"id" json:"id"`
}

func QueryDataBundle(filter *DataBundleFilter, offset, limit int) (*ginutils.List[*DataBundle], error) {
	var aps []*DataBundle
	var count int64
	if err := GetDB().Model(&DataBundle{}).Where(&filter).Count(&count).Preload("DataBundleDetails").Offset(offset).Limit(limit).Find(&aps).Error; err != nil {
		return nil, err
	}
	return &ginutils.List[*DataBundle]{Items: aps, Count: count}, nil
}

func GetDataBundleById(id uint) (*DataBundle, error) {
	var d *DataBundle
	if err := GetDB().Model(&DataBundle{}).Preload("DataBundleDetails").First(&d, id).Error; err != nil {
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
