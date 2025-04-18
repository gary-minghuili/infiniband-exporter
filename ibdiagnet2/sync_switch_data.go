package ibdiagnet2

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"infiniband_exporter/log"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// var (
// 	ipAddress     = "10.4.101.1"
// 	user          = "admin"
// 	password      = "Canopy@123456"
// 	hostUser      = "test"
// 	hostPassword  = "test1234"
// 	hostIpAddress = "10.4.254.250"
// 	hostFilePath  = fmt.Sprintf("/home/%s/ib.tgz", hostUser)
// )

type SyncData interface {
	SyncSwitchData() (bool, error)
}
type SyncSwitchData struct {
	IpAddress     string
	User          string
	Password      string
	HostUser      string
	HostPassword  string
	HostIpAddress string
	HostFilePath  string
}

// 定义与JSON响应对应的结构体
type SyncResponse struct {
	Results []SyncResult `json:"results"`
}

type SyncResult struct {
	Status          string `json:"status"`
	ExecutedCommand string `json:"executed_command"`
	StatusMessage   string `json:"status_message"`
	Data            string `json:"data"`
}

func (s *SyncSwitchData) SyncSwitchData() (bool, error) {
	loginURL := fmt.Sprintf("https://%s/admin/launch?script=rh&template=login&action=login", s.IpAddress)
	data := url.Values{}
	data.Set("d_user_id", "user_id")
	data.Set("t_user_id", "string")
	data.Set("c_user_id", "string")
	data.Set("e_user_id", "true")
	data.Set("f_user_id", s.User)
	data.Set("f_password", s.Password)
	data.Set("Login", "")
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Login failed:%s", err))
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	resp, err := client.Do(req)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Send login request error:%s", err))
		return false, err
	}
	defer resp.Body.Close()
	sessionCookie := ""
	for _, cookie := range resp.Cookies() {
		fmt.Println(cookie.Name)
		if strings.HasPrefix(cookie.Name, "session") {
			sessionCookie = fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)
			break
		}
	}

	if sessionCookie == "" {
		sessionMessage := "Not found session cookie"
		log.GetLogger().Error(sessionMessage)
		return false, errors.New(sessionMessage)
	}

	secondURL := fmt.Sprintf("https://%s/admin/launch?script=json", s.IpAddress)
	// "commands": ["file ibdiagnet upload ibdiagnet_output.tgz scp://test:test1234@10.4.254.250/home/test/ib.tgz"]
	cmd := fmt.Sprintf(
		"file ibdiagnet upload ibdiagnet_output.tgz scp://%s:%s@%s%s",
		s.HostUser, s.HostPassword, s.HostIpAddress, s.HostFilePath,
	)
	secondData := []byte(fmt.Sprintf(`{ "execution_type": "sync", "commands": ["%s"] }`, cmd))
	secondReq, err := http.NewRequest("POST", secondURL, bytes.NewBuffer(secondData))
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Send second request error:%s", err))
		return false, err
	}
	secondReq.Header.Set("Content-Type", "application/json")
	secondReq.Header.Set("Cookie", sessionCookie)

	secondResp, err := client.Do(secondReq)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Send second request error:%s", err))
		return false, err
	}
	defer secondResp.Body.Close()

	secondBody, err := io.ReadAll(secondResp.Body)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Read second response error:%s", err))
		return false, err
	}

	var syncResponse SyncResponse
	err = json.Unmarshal([]byte(secondBody), &syncResponse)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Unmarshal json error:%s", err))
		return false, err
	}
	for _, result := range syncResponse.Results {
		if result.Status != "OK" {
			log.GetLogger().Error(fmt.Sprintf("Sync switch data failed:%s", result.StatusMessage))
			return false, errors.New(result.StatusMessage)
		}
	}
	return true, nil
}
