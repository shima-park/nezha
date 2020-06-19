package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/shima-park/nezha/pkg/rpc/proto"
	"gopkg.in/yaml.v2"
)

func GetJSON(url string, ret interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status code: %d", resp.StatusCode)
	}

	return handleBody(resp.Body, ret)
}

func PostJSON(url string, data, ret interface{}) error {
	param, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(param))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status code: %d", resp.StatusCode)
	}

	return handleBody(resp.Body, ret)
}

func PostYaml(url string, data, ret interface{}) error {
	param, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(param))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status code: %d", resp.StatusCode)
	}

	return handleBody(resp.Body, ret)
}

func handleBody(r io.Reader, ret interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	res := proto.Result{
		Data: ret,
	}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	if res.Code != 0 {
		return errors.New(res.Msg)
	}

	return nil
}

func normalizeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	return strings.TrimSuffix(url, "/")
}
