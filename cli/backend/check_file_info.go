package backend

type CheckFileInfo struct {
	BaseFileName               string `json:"BaseFileName"`
	Size                       int64  `json:"Size"`
	Version                    string `json:"Version"`
	OwnerId                    string `json:"OwnerId"`
	UserId                     string `json:"UserId"`
	UserFriendlyName           string `json:"UserFriendlyName"`
	UserCanWrite               bool   `json:"UserCanWrite"`
	UserCanNotWriteRelative    bool   `json:"UserCanNotWriteRelative"`
	SupportsUpdate             bool   `json:"SupportsUpdate"`
	SupportsLocks              bool   `json:"SupportsLocks"`
	SupportsGetLock            bool   `json:"SupportsGetLock"`
	SupportsExtendedLockLength bool   `json:"SupportsExtendedLockLength"`
	PostMessageOrigin          string `json:"PostMessageOrigin"`
	EnableOwnerTermination     bool   `json:"EnableOwnerTermination"`
	HideSaveOption             bool   `json:"HideSaveOption"`
	HidePrintOption            bool   `json:"HidePrintOption"`
}
