package multikube

import (
	"strings"
	"io/ioutil"
	"net/http"
	"errors"
	"fmt"
)

type Multikube struct {
	Config *Config
	Cache *Cache
}

type APIErrorResponse struct {
	Code int
	Err error
}

func (a APIErrorResponse) Error() string {
	return a.Err.Error()
}

func (a APIErrorResponse) Status() int {
	return a.Code
}

func handleErr(w http.ResponseWriter, err error) {
	if err != nil {
		switch e := err.(type) {
		case APIErrorResponse:
			http.Error(w, e.Error(), e.Status())
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func newErrf(s string, f ...interface{}) error {
	return errors.New(fmt.Sprintf(s, f...))
}

func newErr(s string) error {
	return errors.New(s)
}

func get(url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func put(url string, data []byte) ([]byte, error) {

	body := strings.NewReader(string(data))

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func post(url string, data []byte) ([]byte, error) {

	body := strings.NewReader(string(data))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func delete(url string) ([]byte, error) {

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func New() *Multikube {
	return &Multikube{
		Config: SetupConfig(),
		Cache: &Cache{},
	}
}