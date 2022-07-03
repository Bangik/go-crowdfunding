package transaction

import (
	"errors"
	"go-crowdfunding/campaign"
	"go-crowdfunding/payment"
	"strconv"
)

type service struct {
	repository         Repository
	campaignRepository campaign.Repository
	paymentService     payment.Service
}

type Service interface {
	GetTransactionByCampaignID(input GetCampaignTransacitionsInput) ([]Transaction, error)
	GetTransactionByUserID(userID int) ([]Transaction, error)
	CreateTransaction(input CreateTransactionInput) (Transaction, error)
	ProcessPayment(input TransactionNotificationInput) error
}

func NewService(repository Repository, campaignRepository campaign.Repository, paymentService payment.Service) *service {
	return &service{repository, campaignRepository, paymentService}
}

func (s *service) GetTransactionByCampaignID(input GetCampaignTransacitionsInput) ([]Transaction, error) {
	campaign, err := s.campaignRepository.FindByID(input.ID)
	if err != nil {
		return []Transaction{}, err
	}

	if campaign.UserID != input.User.ID {
		return []Transaction{}, errors.New("you are not allowed to get transactions of this campaign")
	}

	transaction, err := s.repository.GetByCampaignID(input.ID)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *service) GetTransactionByUserID(userID int) ([]Transaction, error) {
	transaction, err := s.repository.GetByUserID(userID)
	if err != nil {
		return transaction, err
	}

	return transaction, nil
}

func (s *service) CreateTransaction(input CreateTransactionInput) (Transaction, error) {
	transaction := Transaction{}
	transaction.Amount = input.Amount
	transaction.CampaignID = input.CampaignID
	transaction.UserID = input.User.ID
	transaction.Status = "pending"

	newTransaction, err := s.repository.Save(transaction)
	if err != nil {
		return newTransaction, err
	}

	paymentTransaction := payment.Transaction{
		ID:     newTransaction.ID,
		Amount: newTransaction.Amount,
	}

	paymentURL, err := s.paymentService.GetPaymentURL(paymentTransaction, input.User)
	if err != nil {
		return newTransaction, err
	}

	newTransaction.PaymentURL = paymentURL

	newTransaction, err = s.repository.Update(newTransaction)
	if err != nil {
		return newTransaction, err
	}

	return newTransaction, nil
}

func (s *service) ProcessPayment(input TransactionNotificationInput) error {
	transactionID, _ := strconv.Atoi(input.OrderID)

	transaction, err := s.repository.GetByID(transactionID)
	if err != nil {
		return err
	}

	if input.PaymentType == "credit_card" {
		if input.FraudStatus == "challenge" {
			transaction.Status = "pending"
		} else if input.FraudStatus == "accept" {
			transaction.Status = "success"
		} else if input.FraudStatus == "reject" {
			transaction.Status = "failed"
		}
	} else if input.PaymentType == "bank_transfer" {
		transaction.Status = "success"
	}

	updatedTransaction, err := s.repository.Update(transaction)
	if err != nil {
		return err
	}

	campaign, err := s.campaignRepository.FindByID(updatedTransaction.CampaignID)
	if err != nil {
		return err
	}

	if updatedTransaction.Status == "success" {
		campaign.BackerCount += 1
		campaign.CurrentAmount += updatedTransaction.Amount

		_, err = s.campaignRepository.Update(campaign)
		if err != nil {
			return err
		}
	}

	return nil
}
