package botapi

import (
	"errors"
	"net/http"
	"strings"

	"telesrv/internal/domain"
)

// rejectLanguageCode mirrors validateDefaultBotCommandScope: this deployment
// does not store per-language bot profile fields, so a non-empty
// language_code must be rejected rather than silently ignored (which would
// make callers believe a localized value was stored when it was not).
func rejectLanguageCode(values map[string]string) error {
	if strings.TrimSpace(values["language_code"]) != "" {
		return errors.New("LANGUAGE_CODE_UNSUPPORTED")
	}
	return nil
}

func (h *handler) setMyName(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.bots == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if err := rejectLanguageCode(values); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	upd := domain.BotInfoUpdate{SetName: true, Name: values["name"]}
	if _, err := h.bots.SetBotInfo(r.Context(), botID, upd); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) getMyName(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.bots == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if err := rejectLanguageCode(values); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	name, _, _, err := h.bots.GetBotInfo(r.Context(), botID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, map[string]any{"name": name})
}

func (h *handler) setMyDescription(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.bots == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if err := rejectLanguageCode(values); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	upd := domain.BotInfoUpdate{SetDescription: true, Description: values["description"]}
	if _, err := h.bots.SetBotInfo(r.Context(), botID, upd); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) getMyDescription(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.bots == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if err := rejectLanguageCode(values); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	_, _, description, err := h.bots.GetBotInfo(r.Context(), botID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, map[string]any{"description": description})
}

func (h *handler) setMyShortDescription(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.bots == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if err := rejectLanguageCode(values); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Bot API's short_description is the text shown alongside the bot's
	// photo/profile and in link previews; this deployment stores it as the
	// bot user's `about` field (mirrors real Telegram's server-side mapping).
	upd := domain.BotInfoUpdate{SetAbout: true, About: values["short_description"]}
	if _, err := h.bots.SetBotInfo(r.Context(), botID, upd); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) getMyShortDescription(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.bots == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if err := rejectLanguageCode(values); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	_, about, _, err := h.bots.GetBotInfo(r.Context(), botID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, map[string]any{"short_description": about})
}
