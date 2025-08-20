package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	ZIP     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int64  `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (d Delivery) Validate() []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(d.Name) == "" {
		errors = append(errors, ValidationError{"delivery.name", "required"})
	}

	if strings.TrimSpace(d.Phone) == "" {
		errors = append(errors, ValidationError{"delivery.phone", "required"})
	}

	if strings.TrimSpace(d.ZIP) == "" {
		errors = append(errors, ValidationError{"delivery.zip", "required"})
	}

	if strings.TrimSpace(d.City) == "" {
		errors = append(errors, ValidationError{"delivery.city", "required"})
	}

	if strings.TrimSpace(d.Address) == "" {
		errors = append(errors, ValidationError{"delivery.address", "required"})
	}

	if strings.TrimSpace(d.Region) == "" {
		errors = append(errors, ValidationError{"delivery.region", "required"})
	}

	if strings.TrimSpace(d.Email) == "" {
		errors = append(errors, ValidationError{"delivery.email", "required"})
	} else if !isValidEmail(d.Email) {
		errors = append(errors, ValidationError{"delivery.email", "invalid format"})
	}

	return errors
}

func (p Payment) Validate() []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(p.Transaction) == "" {
		errors = append(errors, ValidationError{"payment.transaction", "required"})
	}

	if strings.TrimSpace(p.Currency) == "" || len(p.Currency) != 3 {
		errors = append(errors, ValidationError{"payment.currency", "must be 3 characters"})
	}

	if strings.TrimSpace(p.Provider) == "" {
		errors = append(errors, ValidationError{"payment.provider", "required"})
	}

	if p.Amount < 0 {
		errors = append(errors, ValidationError{"payment.amount", "must be non-negative"})
	}

	if p.PaymentDT <= 0 {
		errors = append(errors, ValidationError{"payment.payment_dt", "invalid timestamp"})
	}

	if strings.TrimSpace(p.Bank) == "" {
		errors = append(errors, ValidationError{"payment.bank", "required"})
	}

	if p.DeliveryCost < 0 {
		errors = append(errors, ValidationError{"payment.delivery_cost", "must be non-negative"})
	}

	if p.GoodsTotal < 0 {
		errors = append(errors, ValidationError{"payment.goods_total", "must be non-negative"})
	}

	if p.CustomFee < 0 {
		errors = append(errors, ValidationError{"payment.custom_fee", "must be non-negative"})
	}

	return errors
}

func (i Item) Validate() []ValidationError {
	var errors []ValidationError

	if i.ChrtID <= 0 {
		errors = append(errors, ValidationError{"item.chrt_id", "must be positive"})
	}

	if strings.TrimSpace(i.TrackNumber) == "" {
		errors = append(errors, ValidationError{"item.track_number", "required"})
	}

	if i.Price < 0 {
		errors = append(errors, ValidationError{"item.price", "must be non-negative"})
	}

	if strings.TrimSpace(i.RID) == "" {
		errors = append(errors, ValidationError{"item.rid", "required"})
	}

	if strings.TrimSpace(i.Name) == "" {
		errors = append(errors, ValidationError{"item.name", "required"})
	}

	if i.Sale < 0 {
		errors = append(errors, ValidationError{"item.sale", "must be non-negative"})
	}

	if strings.TrimSpace(i.Size) == "" {
		errors = append(errors, ValidationError{"item.size", "required"})
	}

	if i.TotalPrice < 0 {
		errors = append(errors, ValidationError{"item.total_price", "must be non-negative"})
	}

	if i.NMID <= 0 {
		errors = append(errors, ValidationError{"item.nm_id", "must be positive"})
	}

	if strings.TrimSpace(i.Brand) == "" {
		errors = append(errors, ValidationError{"item.brand", "required"})
	}

	if i.Status <= 0 {
		errors = append(errors, ValidationError{"item.status", "must be positive"})
	}

	return errors
}

func (o Order) Validate() []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(o.OrderUID) == "" {
		errors = append(errors, ValidationError{"order_uid", "required"})
	}

	if strings.TrimSpace(o.TrackNumber) == "" {
		errors = append(errors, ValidationError{"track_number", "required"})
	}

	if strings.TrimSpace(o.Entry) == "" {
		errors = append(errors, ValidationError{"entry", "required"})
	}

	if strings.TrimSpace(o.Locale) == "" {
		errors = append(errors, ValidationError{"locale", "required"})
	}

	if strings.TrimSpace(o.CustomerID) == "" {
		errors = append(errors, ValidationError{"customer_id", "required"})
	}

	if strings.TrimSpace(o.DeliveryService) == "" {
		errors = append(errors, ValidationError{"delivery_service", "required"})
	}

	if strings.TrimSpace(o.ShardKey) == "" {
		errors = append(errors, ValidationError{"shardkey", "required"})
	}

	if o.SmID <= 0 {
		errors = append(errors, ValidationError{"sm_id", "must be positive"})
	}

	if o.DateCreated.IsZero() {
		errors = append(errors, ValidationError{"date_created", "required"})
	}

	if strings.TrimSpace(o.OofShard) == "" {
		errors = append(errors, ValidationError{"oof_shard", "required"})
	}

	deliveryErrors := o.Delivery.Validate()
	for _, err := range deliveryErrors {
		err.Field = "delivery." + err.Field
		errors = append(errors, err)
	}

	paymentErrors := o.Payment.Validate()
	for _, err := range paymentErrors {
		err.Field = "payment." + err.Field
		errors = append(errors, err)
	}

	if len(o.Items) == 0 {
		errors = append(errors, ValidationError{"items", "at least one item required"})
	} else {
		for i, item := range o.Items {
			itemErrors := item.Validate()
			for _, err := range itemErrors {
				err.Field = fmt.Sprintf("items[%d].%s", i, err.Field)
				errors = append(errors, err)
			}
		}
	}

	return errors
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

func (o Order) IsValid() bool {
	return len(o.Validate()) == 0
}

func (o Order) GetValidationError() error {
	errors := o.Validate()
	if len(errors) == 0 {
		return nil
	}

	var errorMessages []string
	for _, err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
}
