package botapi

type Animation struct {
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     int       `json:"duration"`
	Thumb        PhotoSize `json:"thumb"`
	FileName     string    `json:"file_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int       `json:"file_size"`
}

type Audio struct {
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	Duration     int       `json:"duration"`
	Performer    string    `json:"performer"`
	Title        string    `json:"title"`
	Thumb        PhotoSize `json:"thumb"`
	FileName     string    `json:"file_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int       `json:"file_size"`
}

type Document struct {
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	Thumb        PhotoSize `json:"thumb"`
	FileName     string    `json:"file_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int       `json:"file_size"`
}

type Sticker struct {
	FileID       string       `json:"file_id"`
	FileUniqueID string       `json:"file_unique_id"`
	Width        int          `json:"width"`
	Height       int          `json:"height"`
	IsAnimated   bool         `json:"is_animated"`
	Thumb        PhotoSize    `json:"thumb"`
	Emoji        string       `json:"emoji"`
	SetName      string       `json:"set_name"`
	MaskPosition MaskPosition `json:"mask_position"`
	FileSize     int          `json:"file_size"`
}

type Video struct {
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     int       `json:"duration"`
	Thumb        PhotoSize `json:"thumb"`
	FileName     string    `json:"file_name"`
	MimeType     string    `json:"mime_type"`
	FileSize     int       `json:"file_size"`
}

type VideoNote struct {
	FileID       string    `json:"file_id"`
	FileUniqueID string    `json:"file_unique_id"`
	Length       int       `json:"length"`
	Duration     int       `json:"duration"`
	Thumb        PhotoSize `json:"thumb"`
	FileSize     int       `json:"file_size"`
}

type Voice struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Duration     int    `json:"duration"`
	MimeType     string `json:"mime_type"`
	FileSize     int    `json:"file_size"`
}

type Chat struct {
	ID            int      `json:"id"`
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	Username      string   `json:"username"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Bio           string   `json:"bio"`
	Description   string   `json:"description"`
	InviteLink    string   `json:"invite_link"`
	PinnedMessage *Message `json:"pinned_message"`
}

type Message struct {
	MessageID int         `json:"message_id"`
	Chat      Chat        `json:"chat"`
	Animation Animation   `json:"animation"`
	Audio     Audio       `json:"audio"`
	Document  Document    `json:"document"`
	Photo     []PhotoSize `json:"photo"`
	Sticker   Sticker     `json:"sticker"`
	Video     Video       `json:"video"`
	VideoNote VideoNote   `json:"video_note"`
	Voice     Voice       `json:"voice"`
}

type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size"`
}

type MaskPosition struct {
	Point  string  `json:"point"`
	XShift float64 `json:"x_shift"`
	YShift float64 `json:"y_shift"`
	Scale  float64 `json:"scale"`
}
