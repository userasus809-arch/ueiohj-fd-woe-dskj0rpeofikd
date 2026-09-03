package botapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"telesrv/internal/domain"
)

func parseBotAPIUserID(values map[string]string) (int64, bool) {
	userID, err := strconv.ParseInt(strings.TrimSpace(values["user_id"]), 10, 64)
	return userID, err == nil && userID > 0
}

func (h *handler) banChatMember(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	userID, ok := parseBotAPIUserID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	untilDate := 0
	if raw, has := values["until_date"]; has {
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
			return
		}
		untilDate = v
	}
	if _, err := h.gateway.BotAPIBanChatMember(r.Context(), botID, chatID, userID, untilDate); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) unbanChatMember(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	userID, ok := parseBotAPIUserID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	onlyIfBanned := values["only_if_banned"] == "true" || values["only_if_banned"] == "1"
	if _, err := h.gateway.BotAPIUnbanChatMember(r.Context(), botID, chatID, userID, onlyIfBanned); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

// botAPIChatPermissionsInput mirrors the official ChatPermissions object.
// Fields the caller omits keep Go's zero value (false = not allowed),
// matching restrictChatMember's own documented default.
type botAPIChatPermissionsInput struct {
	CanSendMessages       bool `json:"can_send_messages"`
	CanSendAudios         bool `json:"can_send_audios"`
	CanSendDocuments      bool `json:"can_send_documents"`
	CanSendPhotos         bool `json:"can_send_photos"`
	CanSendVideos         bool `json:"can_send_videos"`
	CanSendVideoNotes     bool `json:"can_send_video_notes"`
	CanSendVoiceNotes     bool `json:"can_send_voice_notes"`
	CanSendPolls          bool `json:"can_send_polls"`
	CanSendOtherMessages  bool `json:"can_send_other_messages"`
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews"`
	CanChangeInfo         bool `json:"can_change_info"`
	CanInviteUsers        bool `json:"can_invite_users"`
	CanPinMessages        bool `json:"can_pin_messages"`
	CanManageTopics       bool `json:"can_manage_topics"`
}

// bannedRightsFromChatPermissions inverts the official "can_*  = allowed"
// shape into this deployment's "field = denied" ChannelBannedRights.
func bannedRightsFromChatPermissions(p botAPIChatPermissionsInput) domain.ChannelBannedRights {
	deniedMedia := !p.CanSendPhotos || !p.CanSendVideos || !p.CanSendVideoNotes ||
		!p.CanSendVoiceNotes || !p.CanSendAudios || !p.CanSendDocuments
	return domain.ChannelBannedRights{
		SendMessages:    !p.CanSendMessages,
		SendMedia:       deniedMedia,
		SendPhotos:      !p.CanSendPhotos,
		SendVideos:      !p.CanSendVideos,
		SendRoundvideos: !p.CanSendVideoNotes,
		SendVoices:      !p.CanSendVoiceNotes,
		SendAudios:      !p.CanSendAudios,
		SendDocs:        !p.CanSendDocuments,
		SendPolls:       !p.CanSendPolls,
		SendStickers:    !p.CanSendOtherMessages,
		SendGifs:        !p.CanSendOtherMessages,
		SendGames:       !p.CanSendOtherMessages,
		SendInline:      !p.CanSendOtherMessages,
		EmbedLinks:      !p.CanAddWebPagePreviews,
		ChangeInfo:      !p.CanChangeInfo,
		InviteUsers:     !p.CanInviteUsers,
		PinMessages:     !p.CanPinMessages,
		ManageTopics:    !p.CanManageTopics,
	}
}

func (h *handler) restrictChatMember(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	userID, ok := parseBotAPIUserID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	raw := strings.TrimSpace(values["permissions"])
	if raw == "" {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	var permissionsInput botAPIChatPermissionsInput
	if err := json.Unmarshal([]byte(raw), &permissionsInput); err != nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	untilDate := 0
	if v, has := values["until_date"]; has {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
			return
		}
		untilDate = n
	}
	rights := bannedRightsFromChatPermissions(permissionsInput)
	if _, err := h.gateway.BotAPIRestrictChatMember(r.Context(), botID, chatID, userID, rights, untilDate); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) promoteChatMember(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	userID, ok := parseBotAPIUserID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	boolField := func(key string) bool {
		v := values[key]
		return v == "true" || v == "1"
	}
	rights := domain.ChannelAdminRights{
		ChangeInfo:     boolField("can_change_info"),
		PostMessages:   boolField("can_post_messages"),
		EditMessages:   boolField("can_edit_messages"),
		DeleteMessages: boolField("can_delete_messages"),
		BanUsers:       boolField("can_restrict_members"),
		InviteUsers:    boolField("can_invite_users"),
		PinMessages:    boolField("can_pin_messages"),
		AddAdmins:      boolField("can_promote_members"),
		ManageCall:     boolField("can_manage_video_chats"),
		ManageChat:     boolField("can_manage_chat"),
		ManageTopics:   boolField("can_manage_topics"),
		Anonymous:      boolField("is_anonymous"),
	}
	if _, err := h.gateway.BotAPIPromoteChatMember(r.Context(), botID, chatID, userID, rights); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) pinChatMessage(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	messageID, err := strconv.Atoi(strings.TrimSpace(values["message_id"]))
	if err != nil || messageID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "MESSAGE_ID_INVALID")
		return
	}
	silent := values["disable_notification"] == "true" || values["disable_notification"] == "1"
	if _, err := h.gateway.BotAPIPinChatMessage(r.Context(), botID, chatID, messageID, silent); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) unpinChatMessage(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	messageID := 0
	if raw, has := values["message_id"]; has {
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "MESSAGE_ID_INVALID")
			return
		}
		messageID = v
	}
	if _, err := h.gateway.BotAPIUnpinChatMessage(r.Context(), botID, chatID, messageID); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) unpinAllChatMessages(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	if _, err := h.gateway.BotAPIUnpinAllChatMessages(r.Context(), botID, chatID); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) leaveChat(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	if _, err := h.gateway.BotAPILeaveChat(r.Context(), botID, chatID); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) getChatMemberCount(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	count, err := h.gateway.BotAPIGetChatMemberCount(r.Context(), botID, chatID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, count)
}

func writeBotAPIChatMember(w http.ResponseWriter, member domain.BotAPIChatMember) {
	writeAPIOK(w, botAPIChatMemberJSON(member))
}

func botAPIChatMemberJSON(member domain.BotAPIChatMember) map[string]any {
	user := map[string]any{"id": member.UserID, "is_bot": false}
	if member.FirstName != "" {
		user["first_name"] = member.FirstName
	}
	if member.LastName != "" {
		user["last_name"] = member.LastName
	}
	if member.Username != "" {
		user["username"] = member.Username
	}
	result := map[string]any{"status": member.Status, "user": user}
	if member.CustomTitle != "" {
		result["custom_title"] = member.CustomTitle
	}
	if member.UntilDate > 0 {
		result["until_date"] = member.UntilDate
	}
	switch member.Status {
	case "creator", "administrator":
		result["is_anonymous"] = member.IsAnonymous
		result["can_manage_chat"] = member.CanManageChat
		result["can_delete_messages"] = member.CanDeleteMessages
		result["can_manage_video_chats"] = member.CanManageVideoChats
		result["can_restrict_members"] = member.CanRestrictMembers
		result["can_promote_members"] = member.CanPromoteMembers
		result["can_change_info"] = member.CanChangeInfo
		result["can_invite_users"] = member.CanInviteUsers
		result["can_post_messages"] = member.CanPostMessages
		result["can_edit_messages"] = member.CanEditMessages
		result["can_pin_messages"] = member.CanPinMessages
		result["can_manage_topics"] = member.CanManageTopics
	case "member", "restricted":
		result["is_member"] = member.IsMember
		result["can_send_messages"] = member.CanSendMessages
		result["can_send_polls"] = member.CanSendPolls
		result["can_send_other_messages"] = member.CanSendOtherMessages
		result["can_add_web_page_previews"] = member.CanAddWebPagePreviews
		result["can_pin_messages"] = member.CanPinMessagesMember
		result["can_manage_topics"] = member.CanManageTopicsMember
	}
	return result
}

func (h *handler) getChatMember(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	userID, ok := parseBotAPIUserID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	member, err := h.gateway.BotAPIGetChatMember(r.Context(), botID, chatID, userID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeBotAPIChatMember(w, member)
}

func (h *handler) getChatAdministrators(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, ok := parseBotAPIChatID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	members, err := h.gateway.BotAPIGetChatAdministrators(r.Context(), botID, chatID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	result := make([]map[string]any, 0, len(members))
	for _, m := range members {
		result = append(result, botAPIChatMemberJSON(m))
	}
	writeAPIOK(w, result)
}
