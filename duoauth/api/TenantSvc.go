package api

import (
	"duov6.com/common"
	// notifier "duov6.com/duonotifier/client"
	// "duov6.com/objectstore/client"
	// "duov6.com/session"
	"duov6.com/duoauth/azureapi"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/SiyaDlamini/gorest"
	"net/url"
	"strconv"
	// "strconv"
	"errors"
	"strings"
)

type TenantSvc struct {
	gorest.RestService
	IsServiceReferral    bool            // if the referal is a service based one dont check for session
	getAllTenants        gorest.EndPoint `method:"GET" path:"/tenants" output:"AuthResponse"`
	getTenant            gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}" output:"AuthResponse"`
	createTenant         gorest.EndPoint `method:"POST" path:"/tenants" postdata:"Tenant"`
	updateTenant         gorest.EndPoint `method:"PUT" path:"/tenants" postdata:"Tenant"`
	deleteTenant         gorest.EndPoint `method:"DELETE" path:"/tenants/{tid:string}"`
	getTenantUsers       gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}/users" output:"AuthResponse"`
	addUserToTenant      gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}/adduser/{Email:string}" output:"AuthResponse"`
	deleteUserFromTenant gorest.EndPoint `method:"DELETE" path:"/tenants/{tid:string}/removeuser/{Email:string}"`
	getUserDefaultTenant gorest.EndPoint `method:"GET" path:"/tenants/{userid:string}/getdefault" output:"AuthResponse"`
	setUserDefaultTenant gorest.EndPoint `method:"GET" path:"/tenants/{userid:string}/setdefault/{tid:string}" output:"AuthResponse"`
}

func (T TenantSvc) GetAllTenants() AuthResponse {
	term.Write("Executing Method : Get All Tenants", term.Blank)
	response := AuthResponse{}

	var err error
	id_token := T.Context.Request().Header.Get("Securitytoken")

	if id_token != "" {
		var access_token string
		access_token, err = azureapi.GetGraphApiToken()
		if err == nil {
			graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups?api-version=1.6"
			headers := make(map[string]string)
			headers["Authorization"] = "Bearer " + access_token
			headers["Content-Type"] = "application/json"

			var body []byte
			err, body = common.HTTP_GET(graphUrl, headers, false)
			if err == nil {
				data := make(map[string]interface{})
				_ = json.Unmarshal(body, &data)

				var allTenants []Tenant
				tenantsAsObjects := data["value"].([]interface{})

				for x := 0; x < len(tenantsAsObjects); x++ {
					singleObject := tenantsAsObjects[x].(map[string]interface{})
					descriptionString := (singleObject["description"].(string))
					tenant := Tenant{}
					if err = json.Unmarshal([]byte(descriptionString), &tenant); err == nil {
						tenant.TenantID = singleObject["displayName"].(string)
						tenant.ObjectID = singleObject["objectId"].(string)
					}
					allTenants = append(allTenants, tenant)
				}
				response.Status = true
				response.Message = "Successfully retrieved tenant information."
				response.Data = allTenants
			}
		}
	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
		response.Data = Tenant{}
	}

	return response
}

func (T TenantSvc) GetTenant(tid string) AuthResponse {
	term.Write("Executing Method : Get Tenant Info", term.Blank)
	response := AuthResponse{}
	var err error
	id_token := T.Context.Request().Header.Get("Securitytoken")
	if T.IsServiceReferral || id_token != "" {
		var access_token string
		access_token, err = azureapi.GetGraphApiToken()
		if err == nil {
			//token is good. proceed.
			graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups?api-version=1.6&$filter=" + url.QueryEscape("displayName eq '"+tid+"'")
			headers := make(map[string]string)
			headers["Authorization"] = "Bearer " + access_token
			headers["Content-Type"] = "application/json"

			var body []byte
			err, body = common.HTTP_GET(graphUrl, headers, false)
			if err == nil {
				data := make(map[string]interface{})
				_ = json.Unmarshal(body, &data)

				if len(data["value"].([]interface{})) > 0 {
					//tenant found.
					descriptionString := (((data["value"].([]interface{}))[0]).(map[string]interface{}))["description"].(string)
					tenant := Tenant{}
					if err = json.Unmarshal([]byte(descriptionString), &tenant); err == nil {
						tenant.TenantID = tid
						tenant.ObjectID = (((data["value"].([]interface{}))[0]).(map[string]interface{}))["objectId"].(string)
						response.Status = true
						response.Message = "Successfully retrieved tenant information."
						response.Data = tenant
					}
				} else {
					//tenant not found
					err = errors.New("Tenant not found.")
				}
			}
		}
	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
		response.Data = Tenant{}
	}

	return response
}

func (T TenantSvc) CreateTenant(tenant Tenant) {
	term.Write("Executing Method : Create a tenant.", term.Blank)
	response := AuthResponse{}
	var err error

	id_token := T.Context.Request().Header.Get("Securitytoken")
	if T.IsServiceReferral || id_token != "" {
		var access_token string
		access_token, err = azureapi.GetGraphApiToken()
		if err == nil {
			graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups?api-version=1.6"
			headers := make(map[string]string)
			headers["Authorization"] = "Bearer " + access_token
			headers["Content-Type"] = "application/json"

			jsonString := `{"displayName": "` + tenant.TenantID + `","mailNickname": "` + tenant.TenantID + `","mailEnabled": false,"securityEnabled": true,"description": "{\"Admin\":\"` + tenant.Admin + `\",\"Country\":\"` + tenant.Country + `\",\"Type\":\"` + tenant.Type + `\"}"}`

			err, _ = common.HTTP_POST(graphUrl, headers, []byte(jsonString), false)
			if err == nil {
				response.Status = true
				response.Message = "Tenant created successfully."
				response.Data = tenant
			}
		}
	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
		response.Data = Tenant{}
		b, _ := json.Marshal(response)
		T.ResponseBuilder().SetResponseCode(500).WriteAndOveride(b)
	} else {
		b, _ := json.Marshal(response)
		T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	}
}

func (T TenantSvc) UpdateTenant(tenant Tenant) {
	term.Write("Executing Method : Update Tenant.", term.Blank)
	response := AuthResponse{}
	response.Status = false
	response.Message = "Not implemented yet."
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(501).WriteAndOveride(b)
}

func (T TenantSvc) DeleteTenant(tid string) {
	term.Write("Executing Method : Delete Tenant.", term.Blank)
	response := AuthResponse{}
	response.Status = false
	response.Message = "Not implemented yet."
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
}

func (T TenantSvc) GetTenantUsers(tid string) AuthResponse {
	term.Write("Executing Method : Get Tenant Users", term.Blank)
	response := AuthResponse{}

	var err error
	id_token := T.Context.Request().Header.Get("Securitytoken")

	if id_token != "" {
		var access_token string
		access_token, err = azureapi.GetGraphApiToken()
		if err == nil {
			//get the tenant...
			tResp := T.GetTenant(tid)
			tenant := tResp.Data.(Tenant)
			if tResp.Status {
				graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups/" + tenant.ObjectID + "/members?api-version=1.6"
				headers := make(map[string]string)
				headers["Authorization"] = "Bearer " + access_token
				headers["Content-Type"] = "application/json"

				var body []byte
				err, body = common.HTTP_GET(graphUrl, headers, false)
				if err == nil {
					data := make(map[string]interface{})
					_ = json.Unmarshal(body, &data)

					var allUsers []User
					tenantsAsObjects := data["value"].([]interface{})

					for x := 0; x < len(tenantsAsObjects); x++ {
						singleObject := tenantsAsObjects[x].(map[string]interface{})
						user := User{}
						user.ObjectID = singleObject["objectId"].(string)
						user.EmailAddress = singleObject["otherMails"].([]interface{})[0].(string)
						user.Name = singleObject["displayName"].(string)
						user.Country = singleObject["country"].(string)
						user.Scopes = strings.Split(singleObject["jobTitle"].(string), "-")

						tenantString := ""
						if singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"] != nil {
							tenantString += singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"].(string)
						}
						if singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"] != nil {
							tenantString += "-" + singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"].(string)
						}
						if singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"] != nil {
							tenantString += "-" + singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"].(string)
						}
						if singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"] != nil {
							tenantString += "-" + singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"].(string)
						}
						if singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"] != nil {
							tenantString += "-" + singleObject["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"].(string)
						}

						alltenants := strings.Split(tenantString, "-")
						userTenant := make([]UserTenant, len(alltenants))
						for x := 0; x < len(alltenants); x++ {
							entry := alltenants[x]
							singleTenant := UserTenant{}
							if strings.Contains(entry, "default#") {
								singleTenant.IsDefault = true
								entry = strings.Replace(entry, "default#", "", -1)
							}
							if strings.Contains(entry, "admin#") {
								singleTenant.IsAdmin = true
								entry = strings.Replace(entry, "admin#", "", -1)
							}
							singleTenant.TenantID = entry
							userTenant[x] = singleTenant
						}

						user.Tenants = userTenant

						allUsers = append(allUsers, user)
					}
					response.Status = true
					response.Message = "Successfully retrieved all users for tenant."
					response.Data = allUsers
				}
			} else {
				err = errors.New(tResp.Message)
			}
		}
	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
		response.Data = Tenant{}
	}
	return response
}

func (T TenantSvc) AddUserToTenant(tid, Email string) AuthResponse {
	term.Write("Executing Method : Add user to Tenant", term.Blank)
	response := AuthResponse{}
	id_token := T.Context.Request().Header.Get("Securitytoken")
	var err error
	A := Auth{}
	A.RestService.Context = T.Context

	access_token, err := azureapi.GetGraphApiToken()
	if err != nil {
		response.Status = false
		response.Message = err.Error()
		return response
	}

	if T.IsServiceReferral || id_token != "" {
		//check if newuser, or invited registration or tenant invitation
		isNewUser := false
		isTenantInvite := false
		isInvitedRegistration := false
		tenantString := ""

		whichExtension := 0
		whichExtensionText := ""

		tData := ""
		oldData := "" //for rollback process

		userObjectID := ""

		if T.Context.Request().Header.Get("Invitetype") == "invitation" || T.Context.Request().Header.Get("Invitetype") == "subscription" {
			isTenantInvite = true
		}

		if T.IsServiceReferral {
			//get user and user id,
			graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/?api-version=1.6&$filter=otherMails/any" + url.QueryEscape("(o: o eq '"+Email+"')")
			headers := make(map[string]string)
			headers["Authorization"] = "Bearer " + access_token
			headers["Content-Type"] = "application/json"

			var body []byte
			err, body = common.HTTP_GET(graphUrl, headers, false)
			if err == nil {
				data := make(map[string]interface{})
				_ = json.Unmarshal(body, &data)

				userData := make(map[string]interface{})
				userData = data["value"].([]interface{})[0].(map[string]interface{})

				if userData["objectId"] != nil {

					userObjectID = userData["objectId"].(string)
					//extension_9239d4f1848b43dda66014d3c4f990b9_Tenant
					//check if user already available...

					if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"] != nil {
						tenantString += userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"].(string)
					}
					if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"] != nil {
						tenantString += "-" + userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"].(string)
					}
					if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"] != nil {
						tenantString += "-" + userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"].(string)
					}
					if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"] != nil {
						tenantString += "-" + userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"].(string)
					}
					if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"] != nil {
						tenantString += "-" + userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"].(string)
					}

					if tenantString == "" {
						isNewUser = true
					}

					if T.Context.Request().Header.Get("Nounce") != "defaultNonce" {
						isInvitedRegistration = true
					}

					if isNewUser {
						whichExtension = 0
						whichExtensionText = ""
					} else {
						//elect which extension should be updated
						if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"] != nil {
							whichExtension = 0
							whichExtensionText = ""
						}
						if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant1"] != nil {
							whichExtension = 1
							whichExtensionText = "1"
						}
						if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant2"] != nil {
							whichExtension = 2
							whichExtensionText = "2"
						}
						if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant3"] != nil {
							whichExtension = 3
							whichExtensionText = "3"
						}
						if userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant4"] != nil {
							whichExtension = 4
							whichExtensionText = "4"
						}

						if !isNewUser {
							if len(userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"+whichExtensionText].(string)) > 240 && whichExtension == 4 { //safe buffer for 256 char limit on field
								whichExtension = (-1)
								whichExtensionText = "invalid"
							} else if len(userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"+whichExtensionText].(string)) > 240 && whichExtension < 4 {
								whichExtension += 1
								whichExtensionText = strconv.Itoa(whichExtension)
							} else if len(userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"+whichExtensionText].(string)) <= 240 && whichExtension <= 4 {
								//nothing to be changed
							}

							tData = userData["extension_9239d4f1848b43dda66014d3c4f990b9_Tenant"+whichExtensionText].(string)
							oldData = tData //for rollbackprocess
						}
					}
				} else {
					err = errors.New("No user found with email : " + Email)
				}

			}
		} else {
			//get session..
			var sessionResponse AuthResponse
			sessionResponse = A.GetSession()
			if sessionResponse.Status {
				//correct request.. update user
				sessionResponse = sessionResponse.Data.(AuthResponse)
				userData := sessionResponse.Data.(map[string]interface{})

				userObjectID = userData["oid"].(string)

				//check if user already available...

				if userData["extension_Tenant"] != nil {
					tenantString += userData["extension_Tenant"].(string)
				}
				if userData["extension_Tenant1"] != nil {
					tenantString += "-" + userData["extension_Tenant1"].(string)
				}
				if userData["extension_Tenant2"] != nil {
					tenantString += "-" + userData["extension_Tenant2"].(string)
				}
				if userData["extension_Tenant3"] != nil {
					tenantString += "-" + userData["extension_Tenant3"].(string)
				}
				if userData["extension_Tenant4"] != nil {
					tenantString += "-" + userData["extension_Tenant4"].(string)
				}

				if userData["newUser"] != nil {
					isNewUser = true
				}
				if userData["nonce"].(string) != "defaultNonce" {
					isInvitedRegistration = true
				}

				if isNewUser {
					whichExtension = 0
					whichExtensionText = ""
				} else {
					//elect which extension should be updated
					if userData["extension_Tenant"] != nil {
						whichExtension = 0
						whichExtensionText = ""
					}
					if userData["extension_Tenant1"] != nil {
						whichExtension = 1
						whichExtensionText = "1"
					}
					if userData["extension_Tenant2"] != nil {
						whichExtension = 2
						whichExtensionText = "2"
					}
					if userData["extension_Tenant3"] != nil {
						whichExtension = 3
						whichExtensionText = "3"
					}
					if userData["extension_Tenant4"] != nil {
						whichExtension = 4
						whichExtensionText = "4"
					}

					if !isNewUser {
						if len(userData["extension_Tenant"+whichExtensionText].(string)) > 240 && whichExtension == 4 { //safe buffer for 256 char limit on field
							whichExtension = (-1)
							whichExtensionText = "invalid"
						} else if len(userData["extension_Tenant"+whichExtensionText].(string)) > 240 && whichExtension < 4 {
							whichExtension += 1
							whichExtensionText = strconv.Itoa(whichExtension)
						} else if len(userData["extension_Tenant"+whichExtensionText].(string)) <= 240 && whichExtension <= 4 {
							//nothing to be changed
						}

						tData = userData["extension_Tenant"+whichExtensionText].(string)
						oldData = tData //for rollbackprocess
					}
				}

			} else {
				err = errors.New(sessionResponse.Message)
				response.Status = false
				response.Message = err.Error()
				return response
			}
		}

		if !isNewUser && strings.Contains(tenantString, tid) {
			response.Status = false
			response.Message = "User already a member in this tenant."
			return response
		}

		if whichExtension >= 0 {
			//append the new tenant

			if isNewUser && !isInvitedRegistration {
				//normally registered new user
				tData += "default#admin#" + tid
			} else if isNewUser && isInvitedRegistration {
				//a new user came to sf from an invite
				tData += "default#" + tid
			} else if isTenantInvite {
				tData += "-" + tid
			} else { //remove this else when going live
				tData += "-" + tid
			}

			tData = strings.TrimPrefix(tData, "-")

			//update user.
			graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + userObjectID + "?api-version=1.6"
			headers := make(map[string]string)
			headers["Authorization"] = "Bearer " + access_token
			headers["Content-Type"] = "application/json"
			postString := `{"extension_9239d4f1848b43dda66014d3c4f990b9_Tenant` + whichExtensionText + `":"` + tData + `"}`

			err, _ = common.HTTP_PATCH(graphUrl, headers, []byte(postString), false)
			if err == nil {
				isRollBack := false
				getTenantResponse := T.GetTenant(tid)
				if getTenantResponse.Status { //no error
					//add user to the group
					tObjectID := getTenantResponse.Data.(Tenant).ObjectID
					graphUrl = "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups/" + tObjectID + "/$links/members?api-version=1.6"
					postString = `{"url": "https://graph.windows.net/smoothflowio.onmicrosoft.com/directoryObjects/` + userObjectID + `"}`
					err, _ = common.HTTP_POST(graphUrl, headers, []byte(postString), false)
					if err != nil {
						fmt.Println(err.Error())
						isRollBack = true
					} else {
						response.Status = true
						response.Message = "User assigned to tenant successfully."
					}
				} else {
					err = errors.New(getTenantResponse.Message)
					isRollBack = true
				}

				if isRollBack {
					//rollback user change
					fmt.Println("Rollbacking user change.")
					graphUrl = "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + userObjectID + "?api-version=1.6"
					postString = `{"extension_9239d4f1848b43dda66014d3c4f990b9_Tenant` + whichExtensionText + `":"` + oldData + `"}`
					_, _ = common.HTTP_PATCH(graphUrl, headers, []byte(postString), false)
				}
			}

		} else {
			err = errors.New("User has reached limits of joining new tenants..")
		}

	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
	}

	return response
}

func (T TenantSvc) DeleteUserFromTenant(tid, Email string) {
	term.Write("Executing Method : Delete Tenant.", term.Blank)
	response := AuthResponse{}
	response.Status = false
	response.Message = "Not implemented yet."
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(501).WriteAndOveride(b)
}

func (T TenantSvc) GetUserDefaultTenant(userid string) AuthResponse {
	term.Write("Executing Method : Get users default tenant", term.Blank)
	response := AuthResponse{}
	response.Status = false
	response.Message = "Not implemented yet."
	return response
}

func (T TenantSvc) SetUserDefaultTenant(userid, tid string) AuthResponse {
	term.Write("Executing Method : Set users default tenant", term.Blank)
	response := AuthResponse{}
	response.Status = false
	response.Message = "Not implemented yet."
	return response
}

/*
func (T TenantSvc) AddUserToTenant(tid, Email string) AuthResponse {
	term.Write("Executing Method : Add user to Tenant", term.Blank)
	response := AuthResponse{}
	id_token := T.Context.Request().Header.Get("Securitytoken")
	var err error
	A := Auth{}
	A.RestService.Context = T.Context

	if id_token != "" {
		//get session..
		var sessionResponse AuthResponse
		sessionResponse = A.GetSession()
		if sessionResponse.Status {
			//correct request.. update user
			sessionResponse = sessionResponse.Data.(AuthResponse)
			userData := sessionResponse.Data.(map[string]interface{})

			//check if user already available...
			tenantString := ""
			if userData["extension_Tenant"] != nil {
				tenantString += userData["extension_Tenant"].(string)
			}
			if userData["extension_Tenant1"] != nil {
				tenantString += "-" + userData["extension_Tenant1"].(string)
			}
			if userData["extension_Tenant2"] != nil {
				tenantString += "-" + userData["extension_Tenant2"].(string)
			}
			if userData["extension_Tenant3"] != nil {
				tenantString += "-" + userData["extension_Tenant3"].(string)
			}
			if userData["extension_Tenant4"] != nil {
				tenantString += "-" + userData["extension_Tenant4"].(string)
			}

			//check if newuser, or invited registration or tenant invitation
			isNewUser := false
			isTenantInvite := false
			isInvitedRegistration := false

			if userData["newUser"] != nil {
				isNewUser = true
			}
			if T.Context.Request().Header.Get("Invitetype") == "invitation" || T.Context.Request().Header.Get("Invitetype") == "subscription" {
				isTenantInvite = true
			}
			if userData["nonce"].(string) != "defaultNonce" {
				isInvitedRegistration = true
			}

			if isNewUser || !strings.Contains(tenantString, tid) {
				whichExtension := 0
				whichExtensionText := ""

				tData := ""
				oldData := "" //for rollback process

				if isNewUser {
					whichExtension = 0
					whichExtensionText = ""
				} else {
					//elect which extension should be updated
					if userData["extension_Tenant"] != nil {
						whichExtension = 0
						whichExtensionText = ""
					}
					if userData["extension_Tenant1"] != nil {
						whichExtension = 1
						whichExtensionText = "1"
					}
					if userData["extension_Tenant2"] != nil {
						whichExtension = 2
						whichExtensionText = "2"
					}
					if userData["extension_Tenant3"] != nil {
						whichExtension = 3
						whichExtensionText = "3"
					}
					if userData["extension_Tenant4"] != nil {
						whichExtension = 4
						whichExtensionText = "4"
					}

					if !isNewUser {
						if len(userData["extension_Tenant"+whichExtensionText].(string)) > 240 && whichExtension == 4 { //safe buffer for 256 char limit on field
							whichExtension = (-1)
							whichExtensionText = "invalid"
						} else if len(userData["extension_Tenant"+whichExtensionText].(string)) > 240 && whichExtension < 4 {
							whichExtension += 1
							whichExtensionText = strconv.Itoa(whichExtension)
						} else if len(userData["extension_Tenant"+whichExtensionText].(string)) <= 240 && whichExtension <= 4 {
							//nothing to be changed
						}

						tData = userData["extension_Tenant"+whichExtensionText].(string)
						oldData = tData //for rollbackprocess
					}
				}

				if whichExtension >= 0 {
					access_token, err := azureapi.GetGraphApiToken()
					if err == nil {
						//append the new tenant

						if isNewUser && !isInvitedRegistration {
							//normally registered new user
							tData += "default#admin#" + tid
						} else if isNewUser && isInvitedRegistration {
							//a new user came to sf from an invite
							tData += "default#" + tid
						} else if isTenantInvite {
							tData += "-" + tid
						} else { //remove this else when going live
							tData += "-" + tid
						}

						tData = strings.TrimPrefix(tData, "-")

						//update user.
						graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + userData["oid"].(string) + "?api-version=1.6"
						headers := make(map[string]string)
						headers["Authorization"] = "Bearer " + access_token
						headers["Content-Type"] = "application/json"
						postString := `{"extension_9239d4f1848b43dda66014d3c4f990b9_Tenant` + whichExtensionText + `":"` + tData + `"}`

						err, _ = common.HTTP_PATCH(graphUrl, headers, []byte(postString), false)
						if err == nil {
							isRollBack := false
							getTenantResponse := T.GetTenant(tid)
							if getTenantResponse.Status { //no error
								//add user to the group
								tObjectID := getTenantResponse.Data.(Tenant).ObjectID
								graphUrl = "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups/" + tObjectID + "/$links/members?api-version=1.6"
								postString = `{"url": "https://graph.windows.net/smoothflowio.onmicrosoft.com/directoryObjects/` + userData["oid"].(string) + `"}`
								err, _ = common.HTTP_POST(graphUrl, headers, []byte(postString), false)
								if err != nil {
									fmt.Println(err.Error())
									isRollBack = true
								} else {
									response.Status = true
									response.Message = "User assigned to tenant successfully."
								}
							} else {
								err = errors.New(getTenantResponse.Message)
								isRollBack = true
							}

							if isRollBack {
								//rollback user change
								fmt.Println("Rollbacking user change.")
								graphUrl = "https://graph.windows.net/smoothflowio.onmicrosoft.com/users/" + userData["oid"].(string) + "?api-version=1.6"
								postString = `{"extension_9239d4f1848b43dda66014d3c4f990b9_Tenant` + whichExtensionText + `":"` + oldData + `"}`
								_, _ = common.HTTP_PATCH(graphUrl, headers, []byte(postString), false)
							}
						}
					}
				} else {
					err = errors.New("User has reached limits of joining new tenants..")
				}
			} else {
				err = errors.New("User already a member in this tenant.")
			}

		} else {
			err = errors.New(sessionResponse.Message)
		}
	} else {
		err = errors.New("Securitytoken not found in header.")
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
	}

	return response
}
*/