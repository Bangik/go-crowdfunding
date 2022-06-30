package transaction

import "go-crowdfunding/user"

type GetCampaignTransacitionsInput struct {
	ID   int `uri:"id" binding:"required"`
	User user.User
}
