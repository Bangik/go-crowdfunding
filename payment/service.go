package payment

import (
	"go-crowdfunding/user"
	"os"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type service struct {
}

type Service interface {
	GetPaymentURL(transaction Transaction, user user.User) (string, error)
}

func NewService() *service {
	return &service{}
}

func (s *service) GetPaymentURL(transaction Transaction, user user.User) (string, error) {
	var snap snap.Client
	snap.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Sandbox)

	resp, err := snap.CreateTransaction(GenerateSnapReq(user.Name, user.Email, transaction.ID, transaction.Amount))
	if err != nil {
		return "", err
	}

	return resp.RedirectURL, nil
}

func GenerateSnapReq(userName string, email string, orderID int, amount int) *snap.Request {
	snapReq := &snap.Request{
		CustomerDetail: &midtrans.CustomerDetails{
			FName: userName,
			Email: email,
		},
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  strconv.Itoa(orderID),
			GrossAmt: int64(amount),
		},
	}

	return snapReq
}
