package api

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"github.com/pkg/errors"
)

const (
	version = "5.73"
	apiURL  = "https://api.vk.com/method/"
)

type Error struct {
	Code          int    `json:"error_code"`
	Message       string `json:"error_msg"`
	RequestParams []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"request_params"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

type UsersGetResponse struct {
	Users []VkUser `json:"response"`
}

type VkUser struct {
	Id         int64  `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ScreenName string `json:"screen_name"`
	//ScreenName  string `json:"screen_name"`
}

type VK struct {
	UserId string
	AuthKey string
	ServiceKey string
	Version     string
}

func (vk *VK) Friends() ([]string, error) {

	responseData := struct {
		response []int64
	}{}

	if resp, err := vk.Request("friends.getAppUsers", nil); err != nil {
		return nil, err
	} else if err := json.Unmarshal(resp, &responseData); err != nil {
		return nil, err
	} else {
		result := make([]string, len(responseData.response))
		for index, val := range responseData.response {
			result[index] = strconv.FormatInt(val, 10)
		}
		return result, nil
	}
}

func (vk *VK) FriendsData() ([]User, error) {
	responseData := struct {
		response []int64
	}{}

	if resp, err := vk.Request("friends.getAppUser", nil); err != nil {
		return nil, err
	} else if err := json.Unmarshal(resp, &responseData); err != nil {
		return nil, err
	} else {
		return vk.UsersGet(responseData.response)
	}
}

func (vk *VK) Auth() (User, error) {

	var responseData = UsersGetResponse{}
	var user = User{}
	if resp, err := vk.Request("users.get", map[string]string{"user_ids":vk.UserId,"fields": "screen_name"}); err != nil {
		return user, err
	} else if err := json.Unmarshal(resp, &responseData.Users); err != nil {
		log.Printf("%s  >>> %s ", resp, err.Error())
		return user, err
	} else if len(responseData.Users) == 0 {
		return user, errors.New("User not found")
	} else {
		user.Uid = strconv.FormatInt(responseData.Users[0].Id, 10)
		user.FirstName = responseData.Users[0].FirstName
		user.LastName = responseData.Users[0].LastName
		return user, nil
	}
}

func (vk *VK) UsersGet(ids []int64) ([]User, error) {
	var requestParams map[string]string = nil
	if ids != nil {
		bytes, _ := json.Marshal(ids)
		requestParams = map[string]string{"user_ids": string(bytes)}
	}

	var responseData = UsersGetResponse{}
	if resp, err := vk.Request("users.get", requestParams); err != nil {
		return nil, err
	} else if err := json.Unmarshal(resp, &responseData); err != nil {
		return nil, err
	} else if len(responseData.Users) == 0 {
		return nil, errors.New("User not found")
	} else {
		var users = make([]User, len(responseData.Users))
		for index, user := range responseData.Users {
			users[index] = User{
				Uid:       strconv.FormatInt(user.Id, 10),
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}
		}
		return users, nil
	}
}

func (vk *VK) Request(method string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(apiURL + method)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if params != nil {
		for k, v := range params {
			query.Set(k, v)
		}
	}

	query.Set("access_token", vk.ServiceKey)
	query.Set("v", version) 
	u.RawQuery = query.Encode()

	log.Println(u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var handler struct {
		Error    *Error
		Response json.RawMessage
	}
	err = json.Unmarshal(body, &handler)

	if handler.Error != nil {
		return nil, handler.Error
	}

	return handler.Response, nil
}

func NewVkAuthProvider(user_id string, auth_key string, service_key string) AuthProvider {
	return &VK{UserId:user_id, AuthKey:auth_key, ServiceKey:service_key} //SessionData{session_key: session_key,session_secret:session_secret_key}}
}

//func NewVkAuthProvider(app_id string, session_key string, session_secret_key string) AuthProvider {
//	return &VK{}//SessionData{session_key: session_key,session_secret:session_secret_key}}
//}
