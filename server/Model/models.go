package Model

//========================================================================================================
/* Model data */
type Command struct { // bắt buộc có
	Code   string `json:"code,omitempty"`   // mã lệnh
	ID     string `json:"id,omitempty"`     // Id 4 kí tự
	Status string `json:"status,omitempty"` // của server trả về
	//Key    string `json:"key,omitempty"`    // public key for diffihellman
}
type Room struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"` //"private" or "share" == '0' || '1'
	Members []User `json:"members,omitempty"`
}
type User struct {
	Name  string `json:"name,omitempty"`
	Alias string `json:"alias,omitempty"`
	Pass  string `json:"pass,omitempty"`
	Email string `json:"email,omitempty"`
	Ava   string `json:"ava,omitempty"` // byte array được encode sang string base64 của 1 file ảnh
}
type Message struct {
	Data      string `json:"data,omitempty"`
	Type      string `json:"type,omitempty"`
	Author    string `json:"author,omitempty"`
	Date      string `json:"date,omitempty"`
	ExistRoom []Room `json:"existroom,omitempty"`
}
type Object struct {
	RoomData   Room    `json:"room,omitempty"`
	UserData   User    `json:"user,omitempty"`
	MessData   Message `json:"message,omitempty"`
	Commandata Command `json:"command,omitempty"`
}
