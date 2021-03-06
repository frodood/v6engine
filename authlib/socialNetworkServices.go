package authlib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"duov6.com/common"
	"github.com/ChimeraCoder/anaconda"
)

type facebookAuth struct {
}

type twitterAuth struct {
}

type googlePlusAuth struct {
}

func (g *googlePlusAuth) RegisterToken(object map[string]string) (AuthCertificate, string) {
	var auth AuthCertificate

	profileID := object["id"]
	profileName := object["name"]
	domain := object["domain"]
	oauthKey := object["access_token"]
	emailAddress := object["email"]
	isAuthenticated := false

	url := "https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token=" + oauthKey

	h := AuthHandler{}
	s, eErr := h.GetSession(oauthKey, domain)
	if eErr == "" {
		return s, ""
	}

	//Authenticate from GooglePlus
	err, body := common.HTTP_GET(url, nil, false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		data := make(map[string]interface{})
		_ = json.Unmarshal(body, &data)
		if (strings.EqualFold(profileID, data["id"].(string))) && (strings.EqualFold(emailAddress, data["email"].(string))) {
			isAuthenticated = true
			profileName = data["name"].(string)
		}

	}

	if isAuthenticated {
		user, status := h.GetUser(emailAddress)
		if status != "" {
			user.EmailAddress = emailAddress
			user.UserID = profileID
			user.Name = profileName
			randText := common.RandText(5)
			user.Password = randText
			user.ConfirmPassword = randText
			user.Active = true
			user, _ = h.SaveUser(user, false, "registertoken")
		}

		auth.Email = user.EmailAddress
		auth.SecurityToken = oauthKey
		auth.Domain = domain
		auth.UserID = user.UserID
		auth.Name = user.Name
		auth.Otherdata = make(map[string]string)
		auth.Otherdata["auth0"] = oauthKey
		return auth, ""
	}

	return auth, "Error Authendicating"
}

func (t *twitterAuth) RegisterToken(object map[string]string) (AuthCertificate, string) {
	var auth AuthCertificate

	profileID := object["user_id"]
	profileName := object["screen_name"]
	domain := object["domain"]
	consumerKey := object["consumer_key"]
	consumerSecret := object["consumer_secret"]
	oauthToken := object["oauth_token"]
	oauthSecret := object["oauth_token_secret"]
	emailAddress := object["email"]
	isAuthenticated := false

	h := AuthHandler{}
	s, eErr := h.GetSession(oauthToken, domain)
	if eErr == "" {
		return s, ""
	}

	//Authenticate from Twitter
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(oauthToken, oauthSecret)

	intVal, _ := strconv.Atoi(profileID)
	profileIDinINT64 := int64(intVal)

	result, err := api.GetUsersShowById(profileIDinINT64, nil)
	if err != nil {
		fmt.Println(err.Error())
		return auth, ("Error Authenticating from Twitter : " + err.Error())
	} else {
		profileName = result.Name
		if strings.EqualFold(profileID, result.IdStr) {
			isAuthenticated = true
		}
	}

	if isAuthenticated {
		user, status := h.GetUser(emailAddress)
		if status != "" {
			user.EmailAddress = emailAddress
			user.UserID = profileID
			user.Name = profileName
			randText := common.RandText(5)
			user.Password = randText
			user.ConfirmPassword = randText
			user.Active = true
			user, _ = h.SaveUser(user, false, "registertoken")
		}

		auth.Email = user.EmailAddress
		auth.SecurityToken = oauthToken
		auth.Domain = domain
		auth.UserID = user.UserID
		auth.Name = user.Name
		auth.Otherdata = make(map[string]string)
		auth.Otherdata["auth0"] = oauthToken
		return auth, ""
	}
	return auth, "Error Authendicating"
}

func (f *facebookAuth) RegisterToken(object map[string]string) (AuthCertificate, string) {
	var auth AuthCertificate

	access_token := object["access_token"]
	//authority := object["authority"]
	user_id := object["user_id"]
	domain := object["domain"]

	name := ""
	email := ""
	errorString := ""

	isAuthenticated := false

	url := "https://graph.facebook.com/me?access_token=" + access_token + "&fields=id,email,name"

	h := AuthHandler{}
	s, eErr := h.GetSession(access_token, domain)
	if eErr == "" {
		return s, ""
	}

	//Authenticate from GooglePlus
	err, body := common.HTTP_GET(url, nil, false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		data := make(map[string]interface{})
		_ = json.Unmarshal(body, &data)
		if data["email"] == nil {
			//data private
			errorString = "Error :  Email Address is Unavailable or private!"
		} else {
			email = data["email"].(string)
			name = data["name"].(string)
			isAuthenticated = true
		}
	}

	if isAuthenticated {
		user, status := h.GetUser(email)
		if status != "" {
			user.EmailAddress = email
			user.UserID = user_id
			user.Name = name
			randText := common.RandText(5)
			user.Password = randText
			user.ConfirmPassword = randText
			user.Active = true
			user, _ = h.SaveUser(user, false, "registertoken")
		}

		auth.Email = user.EmailAddress
		auth.SecurityToken = access_token
		auth.Domain = domain
		auth.UserID = user.UserID
		auth.Name = user.Name
		auth.Otherdata = make(map[string]string)
		auth.Otherdata["auth0"] = access_token
		return auth, ""
	}

	return auth, errorString
	//return auth, ""
}
