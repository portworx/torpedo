package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {

	url := "https://ip-10-13-8-200.pwx.dev.purestorage.com/v3-public/localProviders/local?action=login"
	method := "POST"

	payload := strings.NewReader(`{
    "username": "admin",
    "password": "Password@123"
}`)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// Decode the response JSON into a Go struct
	//var data map[string]interface{}
	//err = json.NewDecoder(res.Body).Decode(&data)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//token := data["token"]
	//fmt.Printf("The value of the field is %v\n", token)

	body, err := ioutil.ReadAll(res.Body)
	fmt.Println("After...")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Body is --- " + string(body))

}
