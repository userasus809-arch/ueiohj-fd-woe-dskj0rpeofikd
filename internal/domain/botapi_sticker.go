package domain

// BotAPIInputSticker carries one createNewStickerSet/addStickerToSet
// sticker entry. Exactly one of LocationKey (an already-uploaded file_id,
// decoded to its internal location key by the HTTP layer) or FileBytes (a
// fresh multipart upload) is set.
type BotAPIInputSticker struct {
	LocationKey string
	FileBytes   []byte
	FileName    string
	MimeType    string
	Emoji       string
	Keywords    string
}
