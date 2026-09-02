package botapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"telesrv/internal/domain"
)

var errAPIBadRequest = errors.New("BAD_REQUEST")

func parseBotAPIChatID(values map[string]string) (int64, bool) {
	chatID, err := strconv.ParseInt(strings.TrimSpace(values["chat_id"]), 10, 64)
	return chatID, err == nil
}

// inviteLinkParamsFromValues builds BotAPIInviteLinkParams from the request
// body, using key presence (not just a non-zero/non-empty value) to derive
// the Has* flags so editChatInviteLink can tell "explicitly cleared" apart
// from "omitted".
func inviteLinkParamsFromValues(values map[string]string) (domain.BotAPIInviteLinkParams, error) {
	params := domain.BotAPIInviteLinkParams{}
	if name, has := values["name"]; has {
		params.HasName = true
		params.Name = name
	}
	if raw, has := values["expire_date"]; has {
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return domain.BotAPIInviteLinkParams{}, errAPIBadRequest
		}
		params.HasExpireDate = true
		params.ExpireDate = v
	}
	if raw, has := values["member_limit"]; has {
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return domain.BotAPIInviteLinkParams{}, errAPIBadRequest
		}
		params.HasMemberLimit = true
		params.MemberLimit = v
	}
	if raw, has := values["creates_join_request"]; has {
		params.HasCreatesJoinRequest = true
		params.CreatesJoinRequest = raw == "true" || raw == "1"
	}
	return params, nil
}

func writeBotAPIChatInviteLink(w http.ResponseWriter, link domain.BotAPIChatInviteLink) {
	result := map[string]any{
		"invite_link":          link.InviteLink,
		"creator":              map[string]any{"id": link.CreatorUserID, "is_bot": true},
		"creates_join_request": link.CreatesJoinRequest,
		"is_primary":           link.IsPrimary,
		"is_revoked":           link.IsRevoked,
	}
	if link.Name != "" {
		result["name"] = link.Name
	}
	if link.ExpireDate > 0 {
		result["expire_date"] = link.ExpireDate
	}
	if link.MemberLimit > 0 {
		result["member_limit"] = link.MemberLimit
	}
	if link.PendingJoinRequestCount > 0 {
		result["pending_join_request_count"] = link.PendingJoinRequestCount
	}
	writeAPIOK(w, result)
}

func (h *handler) exportChatInviteLink(w http.ResponseWriter, r *http.Request, botID int64) {
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
	link, err := h.gateway.BotAPIExportChatInviteLink(r.Context(), botID, chatID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, link)
}

func (h *handler) createChatInviteLink(w http.ResponseWriter, r *http.Request, botID int64) {
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
	params, err := inviteLinkParamsFromValues(values)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVITE_LINK_INVALID")
		return
	}
	link, err := h.gateway.BotAPICreateChatInviteLink(r.Context(), botID, chatID, params)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeBotAPIChatInviteLink(w, link)
}

func (h *handler) editChatInviteLink(w http.ResponseWriter, r *http.Request, botID int64) {
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
	link := strings.TrimSpace(values["invite_link"])
	if link == "" {
		writeAPIError(w, http.StatusBadRequest, "INVITE_HASH_EMPTY")
		return
	}
	params, err := inviteLinkParamsFromValues(values)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVITE_LINK_INVALID")
		return
	}
	updated, err := h.gateway.BotAPIEditChatInviteLink(r.Context(), botID, chatID, link, params)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeBotAPIChatInviteLink(w, updated)
}

func (h *handler) revokeChatInviteLink(w http.ResponseWriter, r *http.Request, botID int64) {
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
	link := strings.TrimSpace(values["invite_link"])
	if link == "" {
		writeAPIError(w, http.StatusBadRequest, "INVITE_HASH_EMPTY")
		return
	}
	revoked, err := h.gateway.BotAPIRevokeChatInviteLink(r.Context(), botID, chatID, link)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeBotAPIChatInviteLink(w, revoked)
}

func (h *handler) approveChatJoinRequest(w http.ResponseWriter, r *http.Request, botID int64) {
	h.hideChatJoinRequest(w, r, botID, true)
}

func (h *handler) declineChatJoinRequest(w http.ResponseWriter, r *http.Request, botID int64) {
	h.hideChatJoinRequest(w, r, botID, false)
}

func (h *handler) hideChatJoinRequest(w http.ResponseWriter, r *http.Request, botID int64, approved bool) {
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
	userID, err := strconv.ParseInt(strings.TrimSpace(values["user_id"]), 10, 64)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	var callErr error
	if approved {
		_, callErr = h.gateway.BotAPIApproveChatJoinRequest(r.Context(), botID, chatID, userID)
	} else {
		_, callErr = h.gateway.BotAPIDeclineChatJoinRequest(r.Context(), botID, chatID, userID)
	}
	if callErr != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(callErr))
		return
	}
	writeAPIOK(w, true)
}
