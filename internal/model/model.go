package model

type OrderModel struct {
	OrderID  string `gorm:"column:order_id;not null"`
	TicketID string `gorm:"column:ticket_id;not null"`
	UserID   string `gorm:"column:user_id;not null"`
	Status   int8   `gorm:"column:status;not null;default:(0)"`
	Deleted  int8   `gorm:"column:deleted;not null;default:(0)"`
}

func (d OrderModel) TableName() string {
	return "order"
}

type MqModel struct {
	TicketID string `json:"ticket_id"`
	Style    int8   `json:"style"`  //需要做的的操作，1表示加1，-1表示减1
	Amount   int    `json:"amount"` // 票数操作的数字
}
