package botapi

import (
	"net/http"
	"strconv"
	"strings"

	"telesrv/internal/domain"
)

func (h *handler) createForumTopic(w http.ResponseWriter, r *http.Request, botID int64) {
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
	name := strings.TrimSpace(values["name"])
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "TOPIC_TITLE_EMPTY")
		return
	}
	iconColor := 0
	if raw, has := values["icon_color"]; has {
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		iconColor = v
	}
	var iconCustomEmojiID int64
	if raw, has := values["icon_custom_emoji_id"]; has && strings.TrimSpace(raw) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		iconCustomEmojiID = v
	}
	topic, err := h.gateway.BotAPICreateForumTopic(r.Context(), botID, chatID, name, iconColor, iconCustomEmojiID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, botAPIForumTopicJSON(topic))
}

func botAPIForumTopicJSON(topic domain.BotAPIForumTopic) map[string]any {
	result := map[string]any{
		"message_thread_id": topic.MessageThreadID,
		"name":              topic.Name,
		"icon_color":        topic.IconColor,
	}
	if topic.IconCustomEmojiID != 0 {
		result["icon_custom_emoji_id"] = strconv.FormatInt(topic.IconCustomEmojiID, 10)
	}
	return result
}

func parseBotAPITopicID(values map[string]string) (int, bool) {
	topicID, err := strconv.Atoi(strings.TrimSpace(values["message_thread_id"]))
	return topicID, err == nil && topicID > 0
}

func (h *handler) editForumTopic(w http.ResponseWriter, r *http.Request, botID int64) {
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
	topicID, ok := parseBotAPITopicID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "MESSAGE_ID_INVALID")
		return
	}
	var name *string
	if raw, has := values["name"]; has {
		trimmed := strings.TrimSpace(raw)
		name = &trimmed
	}
	var iconCustomEmojiID *int64
	if raw, has := values["icon_custom_emoji_id"]; has {
		v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
			return
		}
		iconCustomEmojiID = &v
	}
	if _, err := h.gateway.BotAPIEditForumTopic(r.Context(), botID, chatID, topicID, name, iconCustomEmojiID); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) closeForumTopic(w http.ResponseWriter, r *http.Request, botID int64) {
	h.toggleForumTopicClosed(w, r, botID, true)
}

func (h *handler) reopenForumTopic(w http.ResponseWriter, r *http.Request, botID int64) {
	h.toggleForumTopicClosed(w, r, botID, false)
}

func (h *handler) toggleForumTopicClosed(w http.ResponseWriter, r *http.Request, botID int64, closed bool) {
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
	topicID, ok := parseBotAPITopicID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "MESSAGE_ID_INVALID")
		return
	}
	var callErr error
	if closed {
		_, callErr = h.gateway.BotAPICloseForumTopic(r.Context(), botID, chatID, topicID)
	} else {
		_, callErr = h.gateway.BotAPIReopenForumTopic(r.Context(), botID, chatID, topicID)
	}
	if callErr != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(callErr))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) deleteForumTopic(w http.ResponseWriter, r *http.Request, botID int64) {
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
	topicID, ok := parseBotAPITopicID(values)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "MESSAGE_ID_INVALID")
		return
	}
	if _, err := h.gateway.BotAPIDeleteForumTopic(r.Context(), botID, chatID, topicID); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}
