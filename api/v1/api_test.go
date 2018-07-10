package v1_test

import (
	"testing"
	"io/ioutil"
	"net/http/httptest"
	"gitlab.com/amimof/multikube/api/v1"
)

func TestApi(t *testing.T) {
	api := v1.NewAPI()
	req := httptest.NewRequest("GET", "/api/v1/namespaces/", nil)
	w := httptest.NewRecorder()
	api.Router.ServeHTTP(w, req)
	
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Content-Type: %s", resp.Header.Get("Content-Type"))
	t.Logf("Body: %s", string(body))

}