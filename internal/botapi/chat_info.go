package botapi

import (
	"net/http"
	"strconv"
	"strings"
)

// officialBotAPIChatActions are the values Telegram's own Bot API accepts;
// anything else is rejected rather than silently accepted and ignored.
var officialBotAPIChatActions = map[string]bool{
	"typing":            true,
	"upload_photo":      true,
	"record_video":      true,
	"upload_video":      true,
	"record_voice":      true,
	"upload_voice":      true,
	"upload_document":   true,
	"choose_sticker":    true,
	"find_location":     true,
	"record_video_note": true,
	"upload_video_note": true,
}

func (h *handler) sendChatAction(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, err := strconv.ParseInt(strings.TrimSpace(values["chat_id"]), 10, 64)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	action := strings.TrimSpace(values["action"])
	if !officialBotAPIChatActions[action] {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ACTION_INVALID")
		return
	}
	if _, err := h.gateway.BotAPISendChatAction(r.Context(), botID, chatID, action); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) getChat(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	chatID, err := strconv.ParseInt(strings.TrimSpace(values["chat_id"]), 10, 64)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "CHAT_ID_INVALID")
		return
	}
	info, err := h.gateway.BotAPIGetChat(r.Context(), botID, chatID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	result := map[string]any{"id": info.ID, "type": info.Type}
	if info.Title != "" {
		result["title"] = info.Title
	}
	if info.Username != "" {
		result["username"] = info.Username
	}
	if info.FirstName != "" {
		result["first_name"] = info.FirstName
	}
	if info.LastName != "" {
		result["last_name"] = info.LastName
	}
	writeAPIOK(w, result)
}
