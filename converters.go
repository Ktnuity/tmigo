package tmigo

// Converter functions to transform map[string]any tags into typed structs

// convertToChatUserstate converts tags to ChatUserstate
func convertToChatUserstate(tags map[string]any) ChatUserstate {
	userstate := ChatUserstate{
		CommonUserstate: convertToCommonUserstate(tags),
	}

	if val, ok := tags["username"].(string); ok {
		userstate.Username = val
	}
	if val, ok := tags["bits"].(string); ok {
		userstate.Bits = val
	}

	return userstate
}

// convertToSubUserstate converts tags to SubUserstate
func convertToSubUserstate(tags map[string]any) SubUserstate {
	userstate := SubUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}

	if val, ok := tags["msg-param-cumulative-months"].(string); ok {
		userstate.MsgParamCumulativeMonths = val
	}
	if val, ok := tags["msg-param-should-share-streak"].(bool); ok {
		userstate.MsgParamShouldShareStreak = val
	}
	if val, ok := tags["msg-param-streak-months"].(string); ok {
		userstate.MsgParamStreakMonths = val
	}

	return userstate
}

// convertToSubMysteryGiftUserstate converts tags to SubMysteryGiftUserstate
func convertToSubMysteryGiftUserstate(tags map[string]any) SubMysteryGiftUserstate {
	userstate := SubMysteryGiftUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}

	if val, ok := tags["msg-param-sender-count"].(string); ok {
		userstate.MsgParamSenderCount = val
	}
	if val, ok := tags["msg-param-origin-id"].(string); ok {
		userstate.MsgParamOriginID = val
	}

	return userstate
}

// convertToSubGiftUserstate converts tags to SubGiftUserstate
func convertToSubGiftUserstate(tags map[string]any) SubGiftUserstate {
	userstate := SubGiftUserstate{
		CommonGiftSubUserstate: convertToCommonGiftSubUserstate(tags),
	}

	if val, ok := tags["msg-param-sender-count"].(string); ok {
		userstate.MsgParamSenderCount = val
	}
	if val, ok := tags["msg-param-origin-id"].(string); ok {
		userstate.MsgParamOriginID = val
	}

	return userstate
}

// convertToAnonSubGiftUserstate converts tags to AnonSubGiftUserstate
func convertToAnonSubGiftUserstate(tags map[string]any) AnonSubGiftUserstate {
	return AnonSubGiftUserstate{
		CommonGiftSubUserstate: convertToCommonGiftSubUserstate(tags),
	}
}

// convertToAnonSubMysteryGiftUserstate converts tags to AnonSubMysteryGiftUserstate
func convertToAnonSubMysteryGiftUserstate(tags map[string]any) AnonSubMysteryGiftUserstate {
	return AnonSubMysteryGiftUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}
}

// convertToSubGiftUpgradeUserstate converts tags to SubGiftUpgradeUserstate
func convertToSubGiftUpgradeUserstate(tags map[string]any) SubGiftUpgradeUserstate {
	userstate := SubGiftUpgradeUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}

	if val, ok := tags["msg-param-sender-name"].(string); ok {
		userstate.MsgParamSenderName = val
	}
	if val, ok := tags["msg-param-sender-login"].(string); ok {
		userstate.MsgParamSenderLogin = val
	}

	return userstate
}

// convertToAnonSubGiftUpgradeUserstate converts tags to AnonSubGiftUpgradeUserstate
func convertToAnonSubGiftUpgradeUserstate(tags map[string]any) AnonSubGiftUpgradeUserstate {
	return AnonSubGiftUpgradeUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}
}

// convertToPrimeUpgradeUserstate converts tags to PrimeUpgradeUserstate
func convertToPrimeUpgradeUserstate(tags map[string]any) PrimeUpgradeUserstate {
	return PrimeUpgradeUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}
}

// convertToRaidUserstate converts tags to RaidUserstate
func convertToRaidUserstate(tags map[string]any) RaidUserstate {
	userstate := RaidUserstate{
		UserNoticeState: convertToUserNoticeState(tags),
	}

	if val, ok := tags["msg-param-displayName"].(string); ok {
		userstate.MsgParamDisplayName = val
	}
	if val, ok := tags["msg-param-login"].(string); ok {
		userstate.MsgParamLogin = val
	}
	if val, ok := tags["msg-param-viewerCount"].(string); ok {
		userstate.MsgParamViewerCount = val
	}

	return userstate
}

// convertToRitualUserstate converts tags to RitualUserstate
func convertToRitualUserstate(tags map[string]any) RitualUserstate {
	userstate := RitualUserstate{
		UserNoticeState: convertToUserNoticeState(tags),
	}

	if val, ok := tags["msg-param-ritual-name"].(string); ok {
		userstate.MsgParamRitualName = val
	}

	return userstate
}

// convertToBanUserstate converts tags to BanUserstate
func convertToBanUserstate(tags map[string]any) BanUserstate {
	userstate := BanUserstate{}

	if val, ok := tags["room-id"].(string); ok {
		userstate.RoomID = val
	}
	if val, ok := tags["target-user-id"].(string); ok {
		userstate.TargetUserID = val
	}
	if val, ok := tags["tmi-sent-ts"].(string); ok {
		userstate.TMISentTs = val
	}

	return userstate
}

// convertToTimeoutUserstate converts tags to TimeoutUserstate
func convertToTimeoutUserstate(tags map[string]any) TimeoutUserstate {
	userstate := TimeoutUserstate{
		BanUserstate: convertToBanUserstate(tags),
	}

	if val, ok := tags["ban-duration"].(string); ok {
		userstate.BanDuration = val
	}

	return userstate
}

// convertToDeleteUserstate converts tags to DeleteUserstate
func convertToDeleteUserstate(tags map[string]any) DeleteUserstate {
	userstate := DeleteUserstate{}

	if val, ok := tags["login"].(string); ok {
		userstate.Login = val
	}
	if val, ok := tags["message"].(string); ok {
		userstate.Message = val
	}
	if val, ok := tags["target-msg-id"].(string); ok {
		userstate.TargetMsgID = val
	}

	return userstate
}

// convertToRoomState converts tags to RoomState
func convertToRoomState(tags map[string]any) RoomState {
	roomstate := RoomState{}

	if val, ok := tags["broadcaster-lang"].(string); ok {
		roomstate.BroadcasterLang = val
	}
	if val, ok := tags["emote-only"].(bool); ok {
		roomstate.EmoteOnly = val
	}
	if val, ok := tags["followers-only"].(string); ok {
		roomstate.FollowersOnly = val
	}
	if val, ok := tags["r9k"].(bool); ok {
		roomstate.R9K = val
	}
	if val, ok := tags["rituals"].(bool); ok {
		roomstate.Rituals = val
	}
	if val, ok := tags["room-id"].(string); ok {
		roomstate.RoomID = val
	}
	if val, ok := tags["slow"].(string); ok {
		roomstate.Slow = val
	}
	if val, ok := tags["subs-only"].(bool); ok {
		roomstate.SubsOnly = val
	}
	if val, ok := tags["channel"].(string); ok {
		roomstate.Channel = val
	}

	return roomstate
}

// Helper converters for embedded structs

// convertToCommonUserstate converts tags to CommonUserstate
func convertToCommonUserstate(tags map[string]any) CommonUserstate {
	userstate := CommonUserstate{
		Extra: make(map[string]any),
	}

	if val, ok := tags["display-name"].(string); ok {
		userstate.DisplayName = val
	}
	if val, ok := tags["color"].(string); ok {
		userstate.Color = val
	}
	if val, ok := tags["badges"].(map[string]string); ok {
		userstate.Badges = val
	}
	if val, ok := tags["badge-info"].(map[string]string); ok {
		userstate.BadgeInfo = val
	}
	if val, ok := tags["mod"].(bool); ok {
		userstate.Mod = val
	}
	if val, ok := tags["subscriber"].(bool); ok {
		userstate.Subscriber = val
	}
	if val, ok := tags["turbo"].(bool); ok {
		userstate.Turbo = val
	}
	if val, ok := tags["user-id"].(string); ok {
		userstate.UserID = val
	}
	if val, ok := tags["room-id"].(string); ok {
		userstate.RoomID = val
	}
	if val, ok := tags["user-type"].(string); ok {
		userstate.UserType = val
	}
	if val, ok := tags["id"].(string); ok {
		userstate.ID = val
	}
	if val, ok := tags["emotes-raw"].(string); ok {
		userstate.EmotesRaw = val
	}
	if val, ok := tags["badges-raw"].(string); ok {
		userstate.BadgesRaw = val
	}
	if val, ok := tags["badge-info-raw"].(string); ok {
		userstate.BadgeInfoRaw = val
	}
	if val, ok := tags["tmi-sent-ts"].(string); ok {
		userstate.TMISentTs = val
	}
	if val, ok := tags["flags"].(string); ok {
		userstate.Flags = val
	}
	if val, ok := tags["message-type"].(string); ok {
		userstate.MessageType = val
	}
	if val, ok := tags["emotes"].(map[string][]string); ok {
		userstate.Emotes = val
	}

	// Store all tags in Extra for additional fields
	for k, v := range tags {
		userstate.Extra[k] = v
	}

	return userstate
}

// convertToUserNoticeState converts tags to UserNoticeState
func convertToUserNoticeState(tags map[string]any) UserNoticeState {
	userstate := UserNoticeState{
		CommonUserstate: convertToCommonUserstate(tags),
	}

	if val, ok := tags["login"].(string); ok {
		userstate.Login = val
	}
	if val, ok := tags["message"].(string); ok {
		userstate.Message = val
	}
	if val, ok := tags["system-msg"].(string); ok {
		userstate.SystemMsg = val
	}

	return userstate
}

// convertToCommonSubUserstate converts tags to CommonSubUserstate
func convertToCommonSubUserstate(tags map[string]any) CommonSubUserstate {
	userstate := CommonSubUserstate{
		UserNoticeState: convertToUserNoticeState(tags),
	}

	if val, ok := tags["msg-param-sub-plan"].(string); ok {
		userstate.MsgParamSubPlan = SubMethod(val)
	}
	if val, ok := tags["msg-param-sub-plan-name"].(string); ok {
		userstate.MsgParamSubPlanName = val
	}

	return userstate
}

// convertToCommonGiftSubUserstate converts tags to CommonGiftSubUserstate
func convertToCommonGiftSubUserstate(tags map[string]any) CommonGiftSubUserstate {
	userstate := CommonGiftSubUserstate{
		CommonSubUserstate: convertToCommonSubUserstate(tags),
	}

	if val, ok := tags["msg-param-recipient-display-name"].(string); ok {
		userstate.MsgParamRecipientDisplayName = val
	}
	if val, ok := tags["msg-param-recipient-id"].(string); ok {
		userstate.MsgParamRecipientID = val
	}
	if val, ok := tags["msg-param-recipient-user-name"].(string); ok {
		userstate.MsgParamRecipientUserName = val
	}
	if val, ok := tags["msg-param-months"].(string); ok {
		userstate.MsgParamMonths = val
	}

	return userstate
}

// convertToSubMethods converts tags to SubMethods
func convertToSubMethods(tags map[string]any) SubMethods {
	methods := SubMethods{}

	if val, ok := tags["msg-param-sub-plan"].(string); ok {
		methods.Plan = SubMethod(val)
	}
	if val, ok := tags["msg-param-sub-plan-name"].(string); ok {
		methods.PlanName = val
	}
	// Check for Prime
	if methods.Plan == SubMethodPrime {
		methods.Prime = true
	}

	return methods
}
