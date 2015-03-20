package tokenrequest

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"

	"github.com/RangelReale/osincli"
	"github.com/golang/glog"

	"github.com/openshift/origin/pkg/auth/server/login"
)

const (
	RequestTokenEndpoint = "/token/request"
	DisplayTokenEndpoint = "/token/display"
)

type endpointDetails struct {
	originOAuthClient *osincli.Client
}

type Endpoints interface {
	Install(mux login.Mux, paths ...string)
}

func NewEndpoints(originOAuthClient *osincli.Client) Endpoints {
	return &endpointDetails{originOAuthClient}
}

// Install registers the request token endpoints into a mux. It is expected that the
// provided prefix will serve all operations
func (endpoints *endpointDetails) Install(mux login.Mux, paths ...string) {
	for _, prefix := range paths {
		mux.HandleFunc(path.Join(prefix, RequestTokenEndpoint), endpoints.requestToken)
		mux.HandleFunc(path.Join(prefix, DisplayTokenEndpoint), endpoints.displayToken)
	}
}

// requestToken works for getting a token in your browser and seeing what your token is
func (endpoints *endpointDetails) requestToken(w http.ResponseWriter, req *http.Request) {
	authReq := endpoints.originOAuthClient.NewAuthorizeRequest(osincli.CODE)
	oauthURL := authReq.GetAuthorizeUrlWithParams("")

	http.Redirect(w, req, oauthURL.String(), http.StatusFound)
}

func (endpoints *endpointDetails) displayToken(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := tokenData{RequestURL: "request"}

	authorizeReq := endpoints.originOAuthClient.NewAuthorizeRequest(osincli.CODE)
	authorizeData, err := authorizeReq.HandleRequest(req)
	if err != nil {
		data.Error = fmt.Sprintf("Error handling auth request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderToken(w, data)
		return
	}

	accessReq := endpoints.originOAuthClient.NewAccessRequest(osincli.AUTHORIZATION_CODE, authorizeData)
	accessData, err := accessReq.GetToken()
	if err != nil {
		data.Error = fmt.Sprintf("Error getting token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderToken(w, data)
		return
	}

	jsonBytes, err := json.MarshalIndent(accessData.ResponseData, "", "   ")
	if err != nil {
		data.Error = fmt.Sprintf("Error marshalling json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderToken(w, data)
		return
	}

	data.OAuthJSON = string(jsonBytes)
	data.AccessToken = accessData.AccessToken
	renderToken(w, data)
}

func renderToken(w io.Writer, data tokenData) {
	if err := tokenTemplate.Execute(w, data); err != nil {
		glog.Errorf("Unable to render token template: %v", err)
	}
}

type tokenData struct {
	Error       string
	OAuthJSON   string
	AccessToken string
	RequestURL  string
}

// TODO: allow template to be read from an external file
var tokenTemplate = template.Must(template.New("tokenTemplate").Parse(`
<style>
	body    { font-family: sans-serif; font-size: 12pt; margin: 2em 5%; background-color: #F9F9F9; }
	pre     { padding-left: 1em; border-left: .25em solid #eee; }
	a       { color: #00f; text-decoration: none; }
	a:hover { text-decoration: underline; }
</style>

{{ if .Error }}
  {{ .Error }}
{{ else }}
  <h3>Here is your brand new OAuth access token:</h3>
  <pre>{{.OAuthJSON}}</pre>
  
  <h3>How do I use this token?</h3>
  <pre>osc --token={{.AccessToken}} &hellip;</pre>
  <pre>curl -H "Authorization: Bearer {{.AccessToken}}" &hellip;</pre>
  
  <h3>How do I delete this token when I'm done?</h3>
  <pre>osc delete oauthaccesstoken {{.AccessToken}}</pre>
  <pre>curl -X DELETE &hellip;/osapi/v1beta1/oAuthAccessTokens/{{.AccessToken}}</pre>
{{ end }}

<br><br>
<a href="{{.RequestURL}}">Request another token</a>
`))
