package goxiaoai

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	USBS             = "https://api.mina.mi.com/remote/ubus"
	SERVICE_AUTH     = "https://account.xiaomi.com/pass/serviceLoginAuth2"
	SERVICE_LOGIN    = "https://account.xiaomi.com/pass/serviceLogin"
	PLAYLIST         = "https://api2.mina.mi.com/music/playlist/v2/lists"
	PLAYLIST_SONGS   = "https://api2.mina.mi.com/music/playlist/v2/songs"
	DEVICE_LIST      = "https://api.mina.mi.com/admin/v2/device_list"
	ASK_API          = "https://userprofile.mina.mi.com/device_profile/v2/conversation"
	SONG_INFO        = "https://api2.mina.mi.com/music/song_info"
	APP_DEVICE_ID    = "3C861A5820190429"
	SDK_VER          = "3.4.1"
	APP_UA           = "APP/com.xiaomi.mico APPV/2.1.17 iosPassportSDK/3.4.1 iOS/13.3.1"
	MINA_UA          = "MISoundBox/2.1.17 (com.xiaomi.mico; build:2.1.55; iOS 13.3.1) Alamofire/4.8.2 MICO/iOSApp/appStore/2.1.17"
	SID              = "micoapi"
	JSON             = "true"
	APPLICATION_JSON = "application/json"
)

type Sign struct {
	Sign string `url:"_sign" json:"_sign"`
	Qs   string `url:"qs" json:"qs"`
}

type AuthInfo struct {
	Qs             string `json:"qs"`
	Ssecurity      string `json:"ssecurity"`
	Code           int    `json:"code"`
	PassToken      string `json:"passToken"`
	Description    string `json:"description"`
	SecurityStatus int    `json:"securityStatus"`
	Nonce          int    `json:"nonce"`
	UserID         int    `json:"userId"`
	CUserID        string `json:"cUserId"`
	Result         string `json:"result"`
	Psecurity      string `json:"psecurity"`
	CAPTCHAURL     string `json:"captchaUrl"`
	Location       string `json:"location"`
	Pwd            int    `json:"pwd"`
	Child          int    `json:"child"`
	Desc           string `json:"desc"`
}

type CommonParam struct {
	Sid  string `url:"sid"`
	Json string `url:"_json"`
}

type Info struct {
	Status         int64          `json:"status"`
	Volume         int64          `json:"volume"`
	LoopType       int64          `json:"loop_type"`
	MediaType      int64          `json:"media_type"`
	PlaySongDetail PlaySongDetail `json:"play_song_detail"`
	TrackList      []string       `json:"track_list"`
}

type PlaySongDetail struct {
	AudioID  string `json:"audio_id"`
	Position int64  `json:"position"`
	Duration int64  `json:"duration"`
}

type UbusParam struct {
	Method    string `url:"method" `
	Message   string `url:"message" `
	Path      string `url:"path" `
	RequestId string `url:"requestId" `
	DeviceId  string `url:"deviceId" `
}

type Message struct {
	Text   string `url:"text" json:"text"`
	Save   int8   `url:"save" json:"save"`
	Media  string `url:"media" json:"media"`
	Volume int8   `url:"volume" json:"volume"`
	Action string `url:"action" json:"action"`
	Url    string `url:"url" json:"url"`
	Type   int8   `url:"type" json:"type"`
}

type Msg struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    []*DeviceInfo `json:"data"`
}

type DeviceInfo struct {
	DeviceID        string           `json:"deviceID"`
	SerialNumber    string           `json:"serialNumber"`
	Name            string           `json:"name"`
	Alias           string           `json:"alias"`
	Current         bool             `json:"current"`
	Presence        string           `json:"presence"`
	Address         string           `json:"address"`
	MiotDID         string           `json:"miotDID"`
	Hardware        string           `json:"hardware"`
	ROMVersion      string           `json:"romVersion"`
	Capabilities    map[string]int64 `json:"capabilities"`
	RemoteCtrlType  string           `json:"remoteCtrlType"`
	DeviceSNProfile string           `json:"deviceSNProfile"`
	DeviceProfile   string           `json:"deviceProfile"`
	BrokerEndpoint  string           `json:"brokerEndpoint"`
	BrokerIndex     int64            `json:"brokerIndex"`
	MAC             string           `json:"mac"`
	SSID            string           `json:"ssid"`
}

type AuthBody struct {
	User     string `url:"user"`
	Hash     string `url:"hash"`
	Callback string `url:"callback"`
	*CommonParam
	*Sign
}

type Xiaoai struct {
	ServiceToken string
	UserId       string
	DeviceId     string
	SerialNumber string
	client       *http.Client
}

func New(username, password string) (x *Xiaoai, err error) {
	x = &Xiaoai{
		client: &http.Client{},
	}

	authInfo, err := x.login(username, password)
	if err != nil {
		err = fmt.Errorf("login fail %v", err)
		return
	}

	token, err := x.getToken(authInfo)
	if err != nil {
		err = fmt.Errorf("get token fail %v", err)
		return
	}

	x.ServiceToken = token
	x.UserId = strconv.Itoa(authInfo.UserID)

	msg, err := x.GetDevices()
	if err != nil {
		err = fmt.Errorf("get device fail %v", err)
		return
	}
	x.DeviceId = msg.Data[0].DeviceID
	x.SerialNumber = msg.Data[0].SerialNumber
	return
}

func (x *Xiaoai) getSign() (sign *Sign, err error) {
	sign = &Sign{}
	req, err := NewRequest(http.MethodGet, SERVICE_LOGIN, nil)
	if err != nil {
		return
	}

	req.URL.RawQuery = fmt.Sprintf("sid=%s&_json=%s", SID, JSON)
	resp, err := x.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data[11:], &sign); err != nil {
		return
	}

	return

}

func (x *Xiaoai) login(username, password string) (authInfo *AuthInfo, err error) {
	authInfo = &AuthInfo{}
	sign, err := x.getSign()
	if err != nil {
		err = fmt.Errorf("get sign error %v", err)
		return
	}

	hash := md5.Sum([]byte(password))
	hashStr := fmt.Sprintf("%x", hash)
	authData := &AuthBody{
		User:     username,
		Hash:     strings.ToUpper(hashStr),
		Callback: "https://api.mina.mi.com/sts",
		CommonParam: &CommonParam{
			Sid:  SID,
			Json: JSON,
		},
		Sign: sign,
	}

	v, err := query.Values(authData)
	if err != nil {
		return
	}

	payload := bytes.NewReader([]byte(v.Encode()))
	req, err := NewRequest(http.MethodPost, SERVICE_AUTH, payload)
	if err != nil {
		return
	}

	req.Header.Add("Cookie", fmt.Sprintf("deviceId=%s;sdkVersion=%s", APP_DEVICE_ID, SDK_VER))
	req.Header.Add("Content-Length", strconv.Itoa(len(v.Encode())))

	resp, err := x.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data[11:], authInfo); err != nil {
		return
	}

	return

}

func (x *Xiaoai) getToken(authInfo *AuthInfo) (token string, err error) {
	signStr := fmt.Sprintf("nonce=%s&%s", strconv.Itoa(authInfo.Nonce), authInfo.Ssecurity)
	clientSign := Sha1Base64(signStr)
	authInfo.Location = fmt.Sprintf("%s&clientSign=%s", authInfo.Location, clientSign)

	req, err := NewRequest(http.MethodGet, authInfo.Location, nil)
	if err != nil {
		return
	}

	resp, err := x.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		token = ParseToekn(resp.Header.Get("Set-Cookie"))
	}

	return

}

func (x *Xiaoai) getCookie() string {

	cookie := fmt.Sprintf("userId=%s;serviceToken=%s", x.UserId, x.ServiceToken)

	if x.DeviceId != "" && x.SerialNumber != "" {
		cookie = fmt.Sprintf("%s;deviceId=%s;sn=%s", cookie, x.DeviceId, x.ServiceToken)
	}
	return cookie
}

func (x *Xiaoai) GetDevices() (msg *Msg, err error) {
	msg = &Msg{}

	req, err := http.NewRequest(http.MethodGet, DEVICE_LIST, nil)
	if err != nil {
		return
	}

	req.Header.Add("Cookie", x.getCookie())
	req.URL.RawQuery = fmt.Sprintf("master=1&requestId=%s", GetRandomString(30))
	//
	resp, err := x.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, msg); err != nil {
		return
	}

	return
}

func (x *Xiaoai) SwitchDevice(index int64) (err error) {
	msg, err := x.GetDevices()
	if err != nil {
		return
	}

	if msg.Data[index] == nil {
		err = fmt.Errorf("not found device by index %d", index)
		return
	}

	x.DeviceId = msg.Data[index].DeviceID
	x.SerialNumber = msg.Data[index].SerialNumber

	return
}

func (x *Xiaoai) Ubus(p *UbusParam) (msg string, err error) {
	p.DeviceId = x.DeviceId
	p.RequestId = fmt.Sprintf("app_ios_%s", GetRandomString(30))
	//
	v, err := query.Values(p)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, USBS, nil)
	if err != nil {
		return
	}
	//
	req.URL.RawQuery = v.Encode()
	req.Header.Add("Cookie", x.getCookie())
	//
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	msg = string(data)

	return
}

func (x *Xiaoai) GetLastAsk() (err error) {
	url := fmt.Sprintf("%s?source=dialogu&hardware=LX06&timestamp=%d&limit=2", ASK_API, time.Now().UnixNano()/1000000)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Add("Cookie", x.getCookie())

	resp, err := x.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	//
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	fmt.Println(string(body))
	//
	return
}

func (x *Xiaoai) Say(text string) (err error) {
	msg, err := json.Marshal(Message{
		Text:  text,
		Save:  0,
		Media: "app_ios",
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "text_to_speech",
		Message: string(msg),
		Path:    "mibrain",
	}); err != nil {
		return
	}

	return
}

func (x *Xiaoai) SetVolume(volume int8) (err error) {
	msg, err := json.Marshal(Message{
		Volume: volume,
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "player_set_volume",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return
}

func (x *Xiaoai) GetVolume() (volume string) {
	msg, err := json.Marshal(Message{})
	if err != nil {
		return
	}
	data, err := x.Ubus(&UbusParam{
		Method:  "player_get_play_status",
		Message: string(msg),
		Path:    "mediaplayer",
	})
	if err != nil {
		return
	}

	c := regexp.MustCompile("\"volume\":(.*?),")
	s := c.FindStringSubmatch(string(data))
	return s[len(s)-1]
}

func (x *Xiaoai) Play() (err error) {
	msg, err := json.Marshal(Message{
		Action: "play",
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "player_play_operation",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return

}

func (x *Xiaoai) Pause() (err error) {
	msg, err := json.Marshal(Message{
		Action: "pause",
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "player_play_operation",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return

}

func (x *Xiaoai) Prev() (err error) {
	msg, err := json.Marshal(Message{
		Action: "prev",
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "player_play_operation",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return
}

func (x *Xiaoai) Next() (err error) {
	msg, err := json.Marshal(Message{
		Action: "next",
	})
	if err != nil {
		return
	}

	//
	if _, err = x.Ubus(&UbusParam{
		Method:  "player_play_operation",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return
}

func (x *Xiaoai) TogglePlayState() (err error) {
	msg, err := json.Marshal(Message{
		Action: "toggle",
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "player_play_operation",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return
}

func (x *Xiaoai) GetStatus() (info *Info, err error) {
	info = &Info{}
	msg, err := json.Marshal(Message{})
	if err != nil {
		return
	}

	data, err := x.Ubus(&UbusParam{
		Method:  "player_get_play_status",
		Message: string(msg),
		Path:    "mediaplayer",
	})
	if err != nil {
		return
	}

	c := regexp.MustCompile("\"info\":\"(.*)\"}}")
	s := c.FindStringSubmatch(data)
	s2 := s[len(s)-1]
	s3 := strings.Replace(s2, "\\", "", -1)

	if err := json.Unmarshal([]byte(s3), info); err != nil {
		log.Print(err)
	}
	return
}

func (x *Xiaoai) PlayUrl(url string) (err error) {
	msg, err := json.Marshal(Message{
		Url:   url,
		Media: "app_ios",
		Type:  1,
	})
	if err != nil {
		return
	}

	if _, err = x.Ubus(&UbusParam{
		Method:  "player_play_url",
		Message: string(msg),
		Path:    "mediaplayer",
	}); err != nil {
		return
	}

	return
}
