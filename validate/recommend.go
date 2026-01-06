package validate

type PGameListRequestParams struct {
	DeviceId    string `json:"device_id,omitempty"`
	Platform    string `json:"os,omitempty"`
	VersionCode int    `json:"version_code,omitempty"`
	UserSports  string `json:"usersports,omitempty"`
	IsDebug     bool   `json:"is_debug,omitempty"`
	RankKeys    string `json:"rank_keys,omitempty"`
}
