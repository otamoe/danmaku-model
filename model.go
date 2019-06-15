package model

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/globalsign/mgo/bson"
	authModel "github.com/otamoe/auth-model"
	"github.com/otamoe/gin-server/errs"
	"github.com/otamoe/gin-server/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type (
	Application struct {
		ID bson.ObjectId `json:"_id,omitempty"`

		Errors     []*errs.Error `json:"errors,omitempty"`
		StatusCode int           `json:"status_code,omitempty"`
		Client     *http.Client  `json:"-"`
	}

	Post struct {
		ID      bson.ObjectId `json:"_id"`
		OwnerID bson.ObjectId `json:"owner_id"`
		Secret  string        `json:"secret"`
		URI     string        `json:"uri"`
		Allow   bool          `json:"allow"`
		Member  bool          `json:"member"`

		Errors     []*errs.Error `json:"errors,omitempty"`
		StatusCode int           `json:"status_code,omitempty"`
	}
)

var (
	APIOrigin          string
	ApplicationOrigin  string
	defaultApplication *Application
)

func Config(apiOrigin, applicationOrigin string) {
	APIOrigin = apiOrigin
	ApplicationOrigin = applicationOrigin
}

func Start() {
	var application *Application
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
		defaultApplication = application
	}()

	client := authModel.GetClientCredentials([]string{"danmaku:all"})
	transport := client.Transport.(*oauth2.Transport)
	var token *oauth2.Token
	if token, err = transport.Source.Token(); err != nil {
		return
	}

	application = &Application{
		ID:     bson.ObjectIdHex(token.Extra("application_id").(string)),
		Client: client,
	}

	if err = application.Get(); err != nil {
		return
	}
	return
}

func (application *Application) Get() (err error) {
	var req *http.Request
	var res *http.Response
	if req, err = http.NewRequest("GET", ApplicationOrigin+"/"+application.ID.Hex()+"/", nil); err != nil {
		return
	}
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*20)
	defer timeoutCancel()
	if res, err = application.Client.Do(req.WithContext(timeoutCtx)); err != nil {
		return
	}
	defer res.Body.Close()

	var bodyBytes []byte
	if bodyBytes, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}
	logrus.Debugf("[COMMENT_MODEL] getApplication %d %s", res.StatusCode, string(bodyBytes))

	if res.StatusCode >= 500 {
		err = &errs.Error{
			Message:    "Application: Status code error",
			StatusCode: res.StatusCode,
		}
		return
	}

	if err = json.Unmarshal(bodyBytes, application); err != nil {
		return
	}

	if len(application.Errors) != 0 {
		err = application.Errors[0]
		return
	}

	return
}

func (application *Application) Update() (err error) {
	var req *http.Request
	var res *http.Response

	var bodyBytes []byte
	if bodyBytes, err = json.Marshal(application); err != nil {
		return
	}

	if req, err = http.NewRequest("POST", ApplicationOrigin+"/"+application.ID.Hex()+"/", bytes.NewReader(bodyBytes)); err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*20)
	defer timeoutCancel()
	if res, err = application.Client.Do(req.WithContext(timeoutCtx)); err != nil {
		return
	}
	defer res.Body.Close()

	if bodyBytes, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}
	logrus.Debugf("[COMMENT_MODEL] updateApplication %d %s", res.StatusCode, string(bodyBytes))

	if res.StatusCode >= 500 {
		err = &errs.Error{
			Message:    "Application: Status code error",
			StatusCode: res.StatusCode,
		}
		return
	}

	resApplication := &Application{}
	if err = json.Unmarshal(bodyBytes, resApplication); err != nil {
		return
	}

	if len(resApplication.Errors) != 0 {
		err = resApplication.Errors[0]
		return
	}

	return
}

func (post *Post) Save() (err error) {
	var req *http.Request
	var res *http.Response

	if post.Secret == "" {
		post.Secret = string(utils.RandByte(8, utils.RandAlphaLowerNumber))
	}

	var bodyBytes []byte
	if bodyBytes, err = json.Marshal(post); err != nil {
		return
	}

	url := ApplicationOrigin + "/" + defaultApplication.ID.Hex() + "/post/"
	if post.ID != "" {
		url += post.ID.Hex() + "/"
	}

	if req, err = http.NewRequest("POST", url, bytes.NewReader(bodyBytes)); err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*20)
	defer timeoutCancel()
	if res, err = defaultApplication.Client.Do(req.WithContext(timeoutCtx)); err != nil {
		return
	}
	defer res.Body.Close()
	if bodyBytes, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}
	logrus.Debugf("[COMMENT_MODEL] SavePost %s %d %s", req.URL.String(), res.StatusCode, string(bodyBytes))

	if res.StatusCode >= 500 {
		err = &errs.Error{
			Message:    "Post: Status code error",
			StatusCode: res.StatusCode,
		}
		return
	}

	if err = json.Unmarshal(bodyBytes, post); err != nil {
		return
	}

	if len(post.Errors) != 0 {
		err = post.Errors[0]
		return
	}

	return
}
func (post *Post) Get(err error) {
	var req *http.Request
	var res *http.Response
	if post.ID == "" {
		err = &errs.Error{
			Message: "ID is required",
		}
		return
	}

	if req, err = http.NewRequest("GET", ApplicationOrigin+"/"+defaultApplication.ID.Hex()+"/"+post.ID.Hex()+"/", nil); err != nil {
		return
	}

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*20)
	defer timeoutCancel()
	if res, err = defaultApplication.Client.Do(req.WithContext(timeoutCtx)); err != nil {
		return
	}
	defer res.Body.Close()

	var bodyBytes []byte
	if bodyBytes, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}
	logrus.Debugf("[COMMENT_MODEL] GetPost %d %s", res.StatusCode, string(bodyBytes))

	if res.StatusCode >= 500 {
		err = &errs.Error{
			Message:    "Post: Status code error",
			StatusCode: res.StatusCode,
		}
		return
	}

	if err = json.Unmarshal(bodyBytes, post); err != nil {
		return
	}

	if len(post.Errors) != 0 {
		err = post.Errors[0]
		return
	}

	return

}
