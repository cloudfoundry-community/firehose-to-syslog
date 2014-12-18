package token

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"regexp"
)

func makeRequest(uaa string, user string, password string) string {
	/*
		So, this is hacky deluxe, but it works.
		We will get redirected to a bogus redirect_uri with a uri parameter containing the token
		TODO, capture the redirect before gorequests try to carry it out.
		TODO, figure out if there is any library that can do this for us.
	*/
	_, _, err := gorequest.New().
		Post(uaa).
		Set("content-type", "application/x-www-form-urlencoded;charset=utf-8").
		Set("accept", "application/json;charset=utf-8").
		Query("client_id=cf").
		Query("response_type=token").
		Query("redirect_uri=https://A-bogus-url-that-wont-likely-ever-be-used.com/").
		Send(fmt.Sprintf("username=%s", user)).
		Send(fmt.Sprintf("password=%s", password)).
		Send("source=credentials").
		End()
	// This will contain a error string with the token in it
	return err[0].Error()
}

func metTokenFromRedirectURL(redirect string) string {
	/*
		redirect will be a error string like this:
		https://danmdsalkjdsalkjdsalkj.com/#access_token=eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI2ZTM3ODRjOC1hODdiLTRkNTUtYWEzNS1jOGI4OGE2Zjk3MTIiLCJzdWIiOiIyZDA1NTMwZS0wMjVhLTRiNGQtOTdjNi1lMDkyNWMyMTE5ZDIiLCJzY29wZSI6WyJwYXNzd29yZC53cml0ZSIsIm9wZW5pZCIsImNsb3VkX2NvbnRyb2xsZXIud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJkb3BwbGVyLmZpcmVob3NlIl0sImNsaWVudF9pZCI6ImNmIiwiY2lkIjoiY2YiLCJ1c2VyX2lkIjoiMmQwNTUzMGUtMDI1YS00YjRkLTk3YzYtZTA5MjVjMjExOWQyIiwidXNlcl9uYW1lIjoiZG9wcGxlciIsImVtYWlsIjoiZG9wcGxlciIsImlhdCI6MTQxODI5ODM1MiwiZXhwIjoxNDE4MzA1NTUyLCJpc3MiOiJodHRwczovL3VhYS5kZXYuY2Yuc3ByaW5nZXItc2JtLmNvbS9vYXV0aC90b2tlbiIsImF1ZCI6WyJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiLCJvcGVuaWQiLCJkb3BwbGVyIl19.ldE_S7wJ_dvskYzMiMWD3CaR1zpfJyg2J0hygK9_OntlH9pyFyuZcG9VrCyldG0Cj5oWmQrSxzLNvfIkXS42BtoXCuUYc4Z4LrrtEYGXwRSVKQPX9G8pPJOQAD3sn5oA6eUQr-4OkII-fsPgYCNIyO6L_CMIQr4Dw3FCrpGBNlg&token_type=bearer&expires_in=7199&scope=password.write%20openid%20cloud_controller.write%20cloud_controller.read%20doppler.firehose&jti=6e3784c8-a87b-4d55-aa35-c8b88a6f9712: dial tcp: lookup danmdsalkjdsalkjdsalkj.com: no such host
		we want
		eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI2ZTM3ODRjOC1hODdiLTRkNTUtYWEzNS1jOGI4OGE2Zjk3MTIiLCJzdWIiOiIyZDA1NTMwZS0wMjVhLTRiNGQtOTdjNi1lMDkyNWMyMTE5ZDIiLCJzY29wZSI6WyJwYXNzd29yZC53cml0ZSIsIm9wZW5pZCIsImNsb3VkX2NvbnRyb2xsZXIud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJkb3BwbGVyLmZpcmVob3NlIl0sImNsaWVudF9pZCI6ImNmIiwiY2lkIjoiY2YiLCJ1c2VyX2lkIjoiMmQwNTUzMGUtMDI1YS00YjRkLTk3YzYtZTA5MjVjMjExOWQyIiwidXNlcl9uYW1lIjoiZG9wcGxlciIsImVtYWlsIjoiZG9wcGxlciIsImlhdCI6MTQxODI5ODM1MiwiZXhwIjoxNDE4MzA1NTUyLCJpc3MiOiJodHRwczovL3VhYS5kZXYuY2Yuc3ByaW5nZXItc2JtLmNvbS9vYXV0aC90b2tlbiIsImF1ZCI6WyJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiLCJvcGVuaWQiLCJkb3BwbGVyIl19.ldE_S7wJ_dvskYzMiMWD3CaR1zpfJyg2J0hygK9_OntlH9pyFyuZcG9VrCyldG0Cj5oWmQrSxzLNvfIkXS42BtoXCuUYc4Z4LrrtEYGXwRSVKQPX9G8pPJOQAD3sn5oA6eUQr-4OkII-fsPgYCNIyO6L_CMIQr4Dw3FCrpGBNlg
	*/
	r, _ := regexp.Compile("#access_token=(.*)&token_type")
	token := r.FindStringSubmatch(redirect)[1]
	return fmt.Sprintf("bearer %s", token)
}

func GetToken(uaa string, user string, password string) string {
	err := makeRequest(uaa, user, password)
	return metTokenFromRedirectURL(err)
}
