package main

//TODO: replace printf with basic logging
//TODO: replace hardcoded info with parameters
import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	hdrXForwardedURI  = "x-forwarded-uri"  //Originaly requested URL
	hdrXWebauthUser   = "X-WebAuth-User"   //Authenticated user id
	hdrXForwardedHost = "X-Forwarded-Host" //Original host
)

type reqStructure struct {
	URL        string              //Original url in the request that has to be authorized
	User       string              //Authenticated user (empty if none)
	Host       string              //Hostname in the original request
	Headers    map[string][]string //Headers in the original request
	Parameters map[string][]string //Queryparams in the original request
	Context    map[string][]string //Information retrieved from external sources
}

type contxtAttr struct {
	nameInPol    string   //Name of the attribute in the policy
	nameInContxt string   //Name of the attribute in the external system
	userInfo     userInfo //Info to identify the source of user information
}

type userInfo struct {
	pluginType string
	instanceID string
}

var paramOpaURL string                 // = "http://192.168.249.25:31756"
var paramMysql1ConnectionString string // = "root:password@tcp(192.168.249.25:32306)/authorization_info"

//Processes a request to Opaefik
func handler(w http.ResponseWriter, r *http.Request) {
	reqStr := newReqStructure(r)
	fmt.Printf("reqStr: \"%s\"\n", reqStr)
	polName := decidePolicy(reqStr)
	fmt.Printf("Policy: \"%s\"\n", polName)
	contxtAttrs := getContextInfo(polName)
	fmt.Printf("contxtAttrs: \"%s\"\n", contxtAttrs)
	reqStr.Context = getContextValues(reqStr, contxtAttrs)
	fmt.Printf("contxtVals: \"%s\"\n", reqStr.Context)
	reqJSON, err := json.Marshal(reqStr)
	if err != nil {
		fmt.Printf("Error '%s' parsing reqStr '%s'", err, reqStr)
		os.Exit(-1)
	}
	fmt.Printf("reqJSON: \"%s\"\n", reqJSON)
	policyResult, code := evaluatePolicy(reqJSON, polName)
	fmt.Printf("policyResult: \"%s\"\n", policyResult)
	fmt.Printf("policyResultCode: \"%s\"\n", code)
	w.WriteHeader(code)
	io.WriteString(w, policyResult)
}

//Returns the result of evaluating the request in Json format on the given policy
func evaluatePolicy(reqJSON []byte, polName string) (body string, code int) {
	code = http.StatusOK
	req, err := http.NewRequest("POST", paramOpaURL, bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	authResponse, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(authResponse))

	var f interface{}
	err = json.Unmarshal(authResponse, &f)
	m := f.(map[string]interface{})
	fmt.Println("\nmap ", m)
	fmt.Println("m[polName] ", m[polName])
	if m[polName] == true {
		code = http.StatusOK
		body = "Policy " + polName + " evaluated to OK"
	} else {
		code = http.StatusForbidden
		body = "Unauthorized applying policy " + polName
	}
	return
}

//Returns the set of context values from defined set of external sources
func getContextValues(r *reqStructure, attrs []contxtAttr) map[string][]string {
	result := make(map[string][]string)
	attrsToLookFor := make([]string, 0)
	reverseLookUp := make(map[string]string)
	for _, attInfo := range attrs {
		if attInfo.userInfo.pluginType == "mySQL" && attInfo.userInfo.instanceID == "1" {
			attrsToLookFor = append(attrsToLookFor, attInfo.nameInContxt)
			reverseLookUp[attInfo.nameInContxt] = attInfo.nameInPol
		} else {
			fmt.Println("Attribute makes referente to unknown UserInfo: ", attInfo)
		}
	}
	fmt.Printf("\nAttrsToLookFor: ", attrsToLookFor)
	fmt.Printf("\nReverseLookUp: ", reverseLookUp)
	if len(attrsToLookFor) > 0 {
		userInformation := getSQL1UserInfo(r.User)
		for _, v := range attrsToLookFor {
			fmt.Printf("\nUserInformation: ", userInformation)
			result[reverseLookUp[v]] = make([]string, 1)
			result[reverseLookUp[v]][0] = userInformation[v]
		}
	}
	return result
}

//Implementation of a SQL connector, which maps to an specific database schema
func getSQL1UserInfo(user string) map[string]string {
	result := make(map[string]string)
	fmt.Printf("Getting user information\n")
	//TODO Pass configuration as parameters
	attrUsername := "Username"
	attrName := "Name"
	attrHouse := "House"
	attrGroup := "Group"
	db, err := sql.Open("mysql", paramMysql1ConnectionString)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	selDB, err := db.Query("select USERS.id " + attrUsername + ", USERS.name " + attrName + ", USERS.house " + attrHouse + ", GROUPS.name '" + attrGroup + "' from USERS JOIN USERS_GROUPS JOIN GROUPS on USERS.id = USERS_GROUPS.id_user and GROUPS.id = USERS_GROUPS.id_group WHERE USERS.id = '" + user + "';")
	if err != nil {
		panic(err.Error())
	}
	defer selDB.Close()
	if selDB.Next() { // We asume there is only one result (or 0) based on the user id
		var username, name, house, group string
		err = selDB.Scan(&username, &name, &house, &group)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("username: \"%s\", name: \"%s\", house: \"%s\", group: \"%s\"\n", username, name, house, group)
		result[attrUsername] = username
		result[attrName] = name
		result[attrHouse] = house
		result[attrGroup] = group
	}
	return result
}

//Returns the set of attributes that are required for the evaluation of given policy
func getContextInfo(polName string) []contxtAttr {
	attrsToGet := make([]contxtAttr, 0)
	//This configuration should be dynamic and read from a configuration component
	switch polName {
	case "authz_public":
	case "authz_authenticated_any":
	case "authz_mac_vharkonen":
	case "authz_group_mentat":
		attrsToGet = append(attrsToGet, contxtAttr{"group", "Group", userInfo{"mySQL", "1"}})
	case "authz_house_atreides":
		attrsToGet = append(attrsToGet, contxtAttr{"house", "House", userInfo{"mySQL", "1"}})
	}
	return attrsToGet
}

//Returns the identifier of the policy that must be applied in this request
func decidePolicy(r *reqStructure) string {
	//This is just a simple implementation. Deciding the policy to use may be as complex as required.
	result := "default"
	if len(r.URL) > 0 {
		result = r.URL[1:len(r.URL)]
		result = strings.Replace(result, "/", "_", -1)
		result = "authz_" + result
	}
	return result
}

//Prints the contents of the original HTTP request that is going to be processed
func printRequest(r *http.Request) {
	fmt.Printf("Headers:\n")
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			fmt.Printf(" \"%s\" - \"%s\"\n", name, h)
		}
	}
	u, _ := url.Parse(r.Header.Get(hdrXForwardedURI))
	fmt.Printf("Original EscapedPath:%s\n", u.EscapedPath())
	fmt.Printf("Original Hostname:%s\n", u.Hostname())
	fmt.Printf("Original IsAbs:%t\n", u.IsAbs())
	fmt.Printf("Original Port:%s\n", u.Port())
	fmt.Printf("Original RequestURI:%s\n", u.RequestURI())
	fmt.Printf("Original String:%s\n", u.String())
	fmt.Printf("Original parameters:\n")
	for param, values := range u.Query() {
		for _, val := range values {
			fmt.Printf(" \"%s\" - \"%s\"\n", param, val)
		}
	}
}

//Translates the original request information from "Traefic-specific" format to Opaefik format.
func newReqStructure(r *http.Request) *reqStructure {
	var result reqStructure
	u, _ := url.Parse(r.Header.Get(hdrXForwardedURI))
	result.URL = u.EscapedPath()
	result.User = r.Header.Get(hdrXWebauthUser)
	result.Host = r.Header.Get(hdrXForwardedHost)
	result.Headers = r.Header
	result.Parameters = u.Query()
	return &result
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: opaefik <paramOpaURL> <paramMysql1ConnectionString>")
		fmt.Println("paramOpaURL: URL to connect to Open Policy Agent.")
		fmt.Println("paramMysql1ConnectionString: SQL connection string to the Mysql database with authorization information.")
		fmt.Println("Example: opaefik \"http://192.168.249.25:31756\" \"root:password@tcp(192.168.249.25:32306)/authorization_info\"")
		os.Exit(-1)
	}
	paramOpaURL = os.Args[1]
	paramMysql1ConnectionString = os.Args[2]
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

