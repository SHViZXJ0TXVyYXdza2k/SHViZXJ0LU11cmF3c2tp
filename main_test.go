package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestHttp(t *testing.T) {
	go main()
	c := &http.Client{}
	url := "http://localhost:8080/api/objects/abc"
	// Create test with wrong header
	payload := strings.NewReader("Dane testowe")
	req, _ := http.NewRequest("PUT", url, payload)
	res, _ := c.Do(req)

	if res.StatusCode != 400 {
		t.Error("Able to create without setting content")
	}
	// Create
	payload = strings.NewReader("Dane testowe")
	req, _ = http.NewRequest("PUT", url, payload)
	req.Header.Add("content-type", "application/json")
	res, _ = c.Do(req)
	if res.StatusCode != 201 {
		t.Error("Cannot create an object")
	}

	// Read
	res, _ = c.Get(url)
	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "{\"Content-Type\":\"application/json\",\"Data\":\"Dane testowe\"}" || res.StatusCode != 200 {
		t.Error("Cannot read an object")
	}

	// Update
	payload = strings.NewReader("Nowe dane testowe")
	req, _ = http.NewRequest("PUT", url, payload)
	req.Header.Add("content-type", "application/json")
	c.Do(req)
	res, _ = c.Get(url)

	body2, _ := ioutil.ReadAll(res.Body)
	if string(body) == string(body2) {
		t.Error("Cannot update an object")
	}

	//Delete
	req, _ = http.NewRequest("DELETE", url, payload)
	res, _ = c.Do(req)
	if res.StatusCode != 200 {
		t.Error("Cannot delete an object")
	}
	res, _ = c.Do(req)
	if res.StatusCode != 404 {
		t.Error("Able to delete non existing object")
	}
	req, _ = http.NewRequest("DELETE", "http://localhost:8080/api/objects/abc-", payload)
	res, _ = c.Do(req)
	if res.StatusCode != 400 {
		t.Error("Able to delete with wrong syntax")
	}

	//Read wrong syntax
	res, _ = c.Get("http://localhost:8080/api/objects/abc-")
	if res.StatusCode != 400 {
		t.Error("Able to read with wrong syntax")
	}
	res, _ = c.Get("http://localhost:8080/api/objects/abcd")
	if res.StatusCode != 404 {
		fmt.Println(res.StatusCode)
		t.Error("Able to read with non existing object")
	}

	//Create limitations
	bigurl := "http://localhost:8080/api/objects/abcabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabcdefhabac"
	req, _ = http.NewRequest("PUT", bigurl, payload)
	req.Header.Add("content-type", "application/json")
	res, _ = c.Do(req)
	if res.StatusCode != 400 {
		t.Error("Able to create an object with looooooong name (:")
	}

	var b strings.Builder
	b.Grow(2048)
	for i := 0; i < 2048; i++ {
		b.WriteByte(0)
	}
	s := b.String()
	bigpayload := strings.NewReader(s)

	req, _ = http.NewRequest("PUT", url, bigpayload)
	req.Header.Add("content-type", "application/json")
	res, _ = c.Do(req)
	if res.StatusCode != 413 {
		t.Error("Able to create too large object")
	}

	// List keys
	payload = strings.NewReader("Dummy data")
	req, _ = http.NewRequest("PUT", "http://localhost:8080/api/objects/key1", payload)
	req.Header.Add("content-type", "application/json")
	c.Do(req)

	payload = strings.NewReader("Dummy data")
	req, _ = http.NewRequest("PUT", "http://localhost:8080/api/objects/Key2", payload)
	req.Header.Add("content-type", "application/json")
	c.Do(req)

	payload = strings.NewReader("Dummy data")
	req, _ = http.NewRequest("PUT", "http://localhost:8080/api/objects/KEY3", payload)
	req.Header.Add("content-type", "application/json")

	c.Do(req)
	res, _ = c.Get("http://localhost:8080/api/objects/")
	body, _ = ioutil.ReadAll(res.Body)
	fmt.Println(string(body))

	defer res.Body.Close()
}
