package botapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"telesrv/internal/domain"
)

// botAPIStickerJSON projects a stored Document holding a sticker onto the
// official Sticker object shape, reusing apiDocument for the shared
// file_id/file_unique_id/mime_type/file_size fields.
func botAPIStickerJSON(doc domain.Document, setName string) map[string]any {
	sticker := apiDocument(doc)
	stickerType := "regular"
	width, height := 0, 0
	emoji := ""
	for _, attr := range doc.Attributes {
		switch attr.Kind {
		case domain.DocAttrSticker:
			width, height = attr.W, attr.H
			emoji = attr.Alt
			if attr.Mask {
				stickerType = "mask"
			}
		case domain.DocAttrCustomEmoji:
			stickerType = "custom_emoji"
			if attr.Alt != "" {
				emoji = attr.Alt
			}
		}
	}
	sticker["type"] = stickerType
	sticker["width"] = width
	sticker["height"] = height
	sticker["is_animated"] = hasDocumentAttribute(doc, domain.DocAttrAnimated)
	sticker["is_video"] = hasDocumentAttribute(doc, domain.DocAttrVideo)
	if emoji != "" {
		sticker["emoji"] = emoji
	}
	if setName != "" {
		sticker["set_name"] = setName
	}
	return sticker
}

func botAPIStickerSetTypeJSON(kind domain.StickerSetKind) string {
	switch kind {
	case domain.StickerSetKindMasks:
		return "mask"
	case domain.StickerSetKindEmoji:
		return "custom_emoji"
	default:
		return "regular"
	}
}

func (h *handler) getStickerSet(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	name := strings.TrimSpace(values["name"])
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "STICKERSET_INVALID")
		return
	}
	set, docs, err := h.gateway.BotAPIGetStickerSet(r.Context(), botID, name)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	stickers := make([]map[string]any, 0, len(docs))
	for _, doc := range docs {
		stickers = append(stickers, botAPIStickerJSON(doc, set.ShortName))
	}
	writeAPIOK(w, map[string]any{
		"name":         set.ShortName,
		"title":        set.Title,
		"sticker_type": botAPIStickerSetTypeJSON(set.Kind),
		"stickers":     stickers,
	})
}

func (h *handler) uploadStickerFile(w http.ResponseWriter, r *http.Request, botID int64) {
	values, files, err := requestValuesWithFiles(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	_, fileBytes, fileName, mimeType, ok := botAPIStickerFileFromRequest(values, files, "sticker")
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
		return
	}
	doc, err := h.gateway.BotAPIUploadStickerFile(r.Context(), botID, fileBytes, fileName, mimeType)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, apiDocument(doc))
}

// botAPIStickerFileFromRequest resolves an InputSticker's "sticker" field:
// either raw multipart bytes (direct upload or attach://field) or an
// already-uploaded file_id. locationKey is only set in the file_id case.
func botAPIStickerFileFromRequest(values map[string]string, files map[string]uploadedFile, field string) (locationKey string, fileBytes []byte, fileName, mimeType string, ok bool) {
	raw := strings.TrimSpace(values[field])
	if strings.HasPrefix(raw, "attach://") {
		key := strings.TrimPrefix(raw, "attach://")
		if file, found := files[key]; found && len(file.Bytes) > 0 {
			return "", file.Bytes, file.Name, file.MimeType, true
		}
		return "", nil, "", "", false
	}
	if file, found := files[field]; found && len(file.Bytes) > 0 {
		return "", file.Bytes, file.Name, file.MimeType, true
	}
	if key, decoded := decodeBotAPIFileID(raw); decoded {
		return key, nil, "", "", true
	}
	return "", nil, "", "", false
}

// botAPIInputStickerFromJSON parses one element of createNewStickerSet's
// "stickers" array (the official InputSticker object).
func botAPIInputStickerFromJSON(raw map[string]any, files map[string]uploadedFile) (domain.BotAPIInputSticker, bool) {
	stickerField, _ := raw["sticker"].(string)
	emoji := ""
	if list, ok := raw["emoji_list"].([]any); ok {
		parts := make([]string, 0, len(list))
		for _, e := range list {
			if s, ok := e.(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
		emoji = strings.Join(parts, "")
	}
	keywords := ""
	if list, ok := raw["keywords"].([]any); ok {
		parts := make([]string, 0, len(list))
		for _, k := range list {
			if s, ok := k.(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
		keywords = strings.Join(parts, ",")
	}
	locationKey, fileBytes, fileName, mimeType, ok := botAPIStickerFileFromRequest(
		map[string]string{"sticker": stickerField}, files, "sticker")
	if !ok {
		return domain.BotAPIInputSticker{}, false
	}
	return domain.BotAPIInputSticker{
		LocationKey: locationKey,
		FileBytes:   fileBytes,
		FileName:    fileName,
		MimeType:    mimeType,
		Emoji:       emoji,
		Keywords:    keywords,
	}, true
}

func (h *handler) createNewStickerSet(w http.ResponseWriter, r *http.Request, botID int64) {
	values, files, err := requestValuesWithFiles(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	ownerUserID, err := strconv.ParseInt(strings.TrimSpace(values["user_id"]), 10, 64)
	if err != nil || ownerUserID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	name := strings.TrimSpace(values["name"])
	title := strings.TrimSpace(values["title"])
	if name == "" || title == "" {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	var rawStickers []map[string]any
	if raw := strings.TrimSpace(values["stickers"]); raw != "" {
		if err := json.Unmarshal([]byte(raw), &rawStickers); err != nil {
			writeAPIError(w, http.StatusBadRequest, "STICKERS_EMPTY")
			return
		}
	}
	if len(rawStickers) == 0 {
		writeAPIError(w, http.StatusBadRequest, "STICKERS_EMPTY")
		return
	}
	stickers := make([]domain.BotAPIInputSticker, 0, len(rawStickers))
	for _, item := range rawStickers {
		sticker, ok := botAPIInputStickerFromJSON(item, files)
		if !ok {
			writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
			return
		}
		stickers = append(stickers, sticker)
	}
	stickerType := strings.TrimSpace(values["sticker_type"])
	if _, err := h.gateway.BotAPICreateNewStickerSet(r.Context(), botID, ownerUserID, name, title, stickerType, stickers); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) addStickerToSet(w http.ResponseWriter, r *http.Request, botID int64) {
	values, files, err := requestValuesWithFiles(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	ownerUserID, err := strconv.ParseInt(strings.TrimSpace(values["user_id"]), 10, 64)
	if err != nil || ownerUserID <= 0 {
		writeAPIError(w, http.StatusBadRequest, "USER_ID_INVALID")
		return
	}
	name := strings.TrimSpace(values["name"])
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "STICKERSET_INVALID")
		return
	}
	raw := strings.TrimSpace(values["sticker"])
	if raw == "" {
		writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
		return
	}
	var stickerObj map[string]any
	if err := json.Unmarshal([]byte(raw), &stickerObj); err != nil {
		writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
		return
	}
	sticker, ok := botAPIInputStickerFromJSON(stickerObj, files)
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
		return
	}
	if _, err := h.gateway.BotAPIAddStickerToSet(r.Context(), botID, ownerUserID, name, sticker); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) deleteStickerFromSet(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	locationKey, ok := decodeBotAPIFileID(values["sticker"])
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
		return
	}
	if _, err := h.gateway.BotAPIDeleteStickerFromSet(r.Context(), botID, locationKey); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) setStickerPositionInSet(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	locationKey, ok := decodeBotAPIFileID(values["sticker"])
	if !ok {
		writeAPIError(w, http.StatusBadRequest, "FILE_ID_INVALID")
		return
	}
	position, err := strconv.Atoi(strings.TrimSpace(values["position"]))
	if err != nil || position < 0 {
		writeAPIError(w, http.StatusBadRequest, "STICKER_POSITION_INVALID")
		return
	}
	if _, err := h.gateway.BotAPISetStickerPositionInSet(r.Context(), botID, locationKey, position); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) setStickerSetTitle(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	name := strings.TrimSpace(values["name"])
	title := strings.TrimSpace(values["title"])
	if name == "" || title == "" {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	if _, err := h.gateway.BotAPISetStickerSetTitle(r.Context(), botID, name, title); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}

func (h *handler) deleteStickerSet(w http.ResponseWriter, r *http.Request, botID int64) {
	values, err := requestValues(r)
	if err != nil || h.gateway == nil {
		writeAPIError(w, http.StatusBadRequest, "BAD_REQUEST")
		return
	}
	name := strings.TrimSpace(values["name"])
	if name == "" {
		writeAPIError(w, http.StatusBadRequest, "STICKERSET_INVALID")
		return
	}
	if _, err := h.gateway.BotAPIDeleteStickerSet(r.Context(), botID, name); err != nil {
		writeAPIError(w, http.StatusBadRequest, apiErrorDescription(err))
		return
	}
	writeAPIOK(w, true)
}
