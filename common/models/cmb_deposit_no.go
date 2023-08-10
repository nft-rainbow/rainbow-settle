package models

// 招行对公充值卡号信息

type CmbDepositNo struct {
	BaseModel
	UserId       uint   `gorm:"type:int;index" json:"user_id"`
	UserName     string `gorm:"type:varchar(255);index" json:"user_name"`    // 用户充值卡开户名
	UserBankNo   string `gorm:"type:varchar(255);index" json:"user_bank_no"` // 用户充值卡号
	UserBankName string `gorm:"type:varchar(255)" json:"user_bank_name"`     // 用户充值卡开户行
	CmbNo        string `gorm:"type:varchar(255);index" json:"cmb_no"`       // 用户专属招行对公充值卡号
}

func (item *CmbDepositNo) FindByUserId(uid uint) (*CmbDepositNo, error) {
	var item1 CmbDepositNo
	err := GetDB().Where("user_id = ?", uid).First(&item1).Error
	return &item1, err
}

func (item *CmbDepositNo) FindByCmbNo(cmbNo string) (*CmbDepositNo, error) {
	var item1 CmbDepositNo
	err := GetDB().Where("cmb_no = ?", cmbNo).First(&item1).Error
	return &item1, err
}
