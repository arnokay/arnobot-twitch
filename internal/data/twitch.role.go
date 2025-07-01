package data

import (
	"github.com/arnokay/arnobot-shared/data"
	"github.com/nicklaw5/helix/v2"
)

func GetChatterRole(badges []helix.EventSubChatBadge) data.ChatterRole {
	role := data.ChatterPleb
	for _, badge := range badges {
		if badge.SetID == "subscriber" {
			role = data.ChatterSub
		}
		if badge.SetID == "vip" {
			role = data.ChatterVIP
		}
		if badge.SetID == "moderator" {
			role = data.ChatterModerator
		}
		if badge.SetID == "broadcaster" {
			role = data.ChatterBroadcaster
		}
	}

	return role
}
