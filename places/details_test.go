package places

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestDetailsCallDo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(handler))
	ts2.Close()

	for _, test := range []struct {
		Name      string
		PlaceID   string
		Language  string
		Extension string
		URL       string
		Want      error
	}{
		{
			Name:      "OK Response",
			PlaceID:   "ChIJLU7jZClu5kcR4PcOOO6p3I0",
			Language:  "en",
			Extension: "review_summary",
			URL:       ts.URL,
			Want:      nil,
		},
		{
			Name:    "Invalid Request",
			PlaceID: "invalid_request",
			URL:     ts.URL,
			Want:    errors.New("INVALID_REQUEST"),
		},
		{
			Name:    "Non-OK Status",
			PlaceID: "notok",
			URL:     ts.URL,
			Want:    errors.New("bad resp 400: "),
		},
		{
			Name:    "Invalid JSON",
			PlaceID: "invalid_json",
			URL:     ts.URL,
			Want:    errors.New("json: cannot unmarshal string into Go value of type places.DetailsResponse"),
		},
		{
			Name:    "communication problem",
			PlaceID: "wrong",
			URL:     ts2.URL,
			Want: errors.New(
				"Get " + ts2.URL + "/details/json?key=testkey&placeid=wrong: dial tcp " + ts2.Listener.Addr().String() + ": getsockopt: connection refused",
			),
		},
	} {
		var service = Service{
			client: http.DefaultClient,
			key:    "testkey",
			url:    test.URL,
		}

		var details = DetailsCall{
			service:    &service,
			placeID:    test.PlaceID,
			Language:   test.Language,
			Extensions: test.Extension,
		}
		_, got := details.Do()

		if got != test.Want {
			if got.Error() != test.Want.Error() {
				t.Errorf("DetailsCall{}.Do() %v = %#v, want %#v",
					test.Name, got, test.Want)
			}
		}
	}
}

func handler(writer http.ResponseWriter, reader *http.Request) {
	uri := reader.URL.RequestURI()

	if uri == "/details/json?extensions=review_summary&key=testkey&language=en&placeid=ChIJLU7jZClu5kcR4PcOOO6p3I0" {
		fmt.Fprintf(writer, readResponse("ok"))
		return
	}

	if uri == "/details/json?key=testkey&placeid=invalid_request" {
		fmt.Fprintf(writer, readResponse("invalid_request"))
		return
	}

	if uri == "/details/json?key=testkey&placeid=invalid_json" {
		fmt.Fprintf(writer, readResponse("invalid_json"))
		return
	}

	if uri == "/details/json?key=testkey&placeid=notok" {
		writer.WriteHeader(400)
		fmt.Fprintf(writer, "")
		return
	}
}

func readResponse(responseType string) string {
	absPath, err := filepath.Abs("../data/" + responseType + ".json")
	if err != nil {
		panic(err)
	}

	response, err := ioutil.ReadFile(absPath)
	if err != nil {
		panic(err)
	}

	return string(response)
}
