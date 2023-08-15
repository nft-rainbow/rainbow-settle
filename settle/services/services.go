package services

import (
	"github.com/nft-rainbow/rainbow-fiat/common/models"
)

var (
	userQuotaOperater *models.UserQuotaOperator
)

func Init() {
	userQuotaOperater = &models.UserQuotaOperator{}
}

func GetUserQuotaOperator() *models.UserQuotaOperator {
	return userQuotaOperater
}
