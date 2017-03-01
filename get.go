package main

import (
	"encoding/json"
	stderr "errors"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

var errInvalidToken = stderr.New("无效token")

func getMagic(uri string) (string, string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", "", errors.Wrap(err, "Response error from :"+uri)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.Errorf("StatusCode(%d)!= %d", resp.StatusCode, http.StatusOK)
	}
	var result = struct {
		Result string
		Token  string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", "", errors.Wrap(err, "JSON decode error")
	}

	return result.Result, result.Token, nil
}

func checkResult(uri, key, token string) error {

	uri = fmt.Sprintf("%s/?key=%s&token=%s", uri, key, token)

	resp, err := http.Get(uri)
	if err != nil {
		return errors.Wrap(err, "Response error from :"+uri)
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var result = struct {
		Info string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return errors.Wrap(err, "JSON decode error")
	}

	return errors.Errorf("%d %s", resp.StatusCode, result.Info)
}
