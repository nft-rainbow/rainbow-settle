package models

import (
	"math"

	. "github.com/ahmetalpbalkan/go-linq"
	"github.com/nft-rainbow/conflux-gin-helper/utils"
	"github.com/nft-rainbow/rainbow-settle/common/models/enums"
	"gorm.io/gorm"
)

const (
	KYC_STATUS_FAIL = iota - 1
	KYC_STATUS_INIT
	KYC_STATUS_OK
)

const (
	USER_TYPE_NORMAL = iota + 1
	USER_TYPE_COMPANY
)

var (
	useCreatedHandlers []UserCreatedHandler
)

type UserCreatedHandler func(tx *gorm.DB, user *User) error

func RegisterUserCreatedEvent(handler UserCreatedHandler) {
	useCreatedHandlers = append(useCreatedHandlers, handler)
}

type User struct {
	BaseModel
	Email         string            `gorm:"uniqueIndex;type:varchar(64)" json:"email"`
	Password      string            `gorm:"type:varchar(128)" json:"password"`
	Name          string            `gorm:"type:varchar(64)" json:"name"`
	Phone         string            `gorm:"type:varchar(64);index" json:"phone"`
	Type          uint              `gorm:"type:int;default:0" json:"type"`   // 1-common, 2-company
	Status        int               `gorm:"type:int;default:0" json:"status"` // -1-fail, 0-init, 1-pass
	IdName        string            `gorm:"type:varchar(64)" json:"id_name"`
	IdNo          string            `gorm:"type:varchar(64);index" json:"id_no"`
	IdImage       string            `gorm:"type:varchar(256)" json:"id_image"`
	KycMsg        string            `gorm:"type:varchar(256)" json:"kyc_msg"`
	EmailVerified bool              `gorm:"type:tinyint;default:0" json:"email_verified"`
	UserPayType   enums.UserPayType `gorm:"type:tinyint;default:1" json:"user_pay_type"`
}

func FindUserByEmail(email string) (*User, error) {
	var user User
	err := db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func FindUserById(id uint) (*User, error) {
	var user User
	err := db.First(&user, id).Error
	return &user, err
}

func FindUserByIds(ids []uint) (map[uint]*User, error) {
	var _users []*User
	if err := db.Where("id in (?)", ids).Find(&_users).Error; err != nil {
		return nil, err
	}

	users := make(map[uint]*User)
	From(_users).
		ToMapByT(&users,
			func(p *User) uint { return p.ID },
			func(p *User) *User { return p },
		)

	return users, nil
}

func UserCount(filter map[string]interface{}) int64 {
	var count int64
	db.Model(&User{}).Where(filter).Count(&count)
	return count
}

func KycReviewCount() int64 {
	var count int64
	db.Model(&User{}).Where("status = 0 and id_name != ''").Count(&count)
	return count
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	if err := tx.Create(NewUserBalance(u.ID)).Error; err != nil {
		return err
	}
	for _, h := range useCreatedHandlers {
		if err := h(tx, u); err != nil {
			return err
		}
	}
	return nil
}

func GetAllUser(filter map[string]interface{}, offset, limit int) ([]*User, error) {
	items := []*User{}
	err := db.Model(&User{}).Where(filter).Offset(offset).Limit(limit).Order("id DESC").Find(&items).Error
	return items, err
}

func GetAllUserIds() ([]uint, error) {
	users, err := GetAllUser(nil, 0, math.MaxInt)
	if err != nil {
		return nil, err
	}
	return utils.MapSlice(users, func(u *User) (uint, error) { return u.ID, nil })
}

func MustGetAllUserIds() []uint {
	return utils.Must(GetAllUserIds())
}
