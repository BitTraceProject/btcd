package tests

import (
	"crypto/tls"
	"io"
	"net/http"
	"strconv"
	"testing"
)

func TestGetTargetGeight(t *testing.T) {
	h, err := getNewTargetHeight()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(h)
}

func getNewTargetHeight() (int32, error) {
	// get target height
	// get https://blockchain.info/q/getblockcount
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}
	resp, err := c.Get("https://blockchain.info/q/getblockcount") // mainchain
	if err != nil {
		return 0, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err // TODO add default height
	}
	height, err := strconv.ParseInt(string(data), 10, 32)
	return int32(height), nil
}
