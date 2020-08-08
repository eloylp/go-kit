package check_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	. "github.com/eloylp/kit/test/check"
)

func TestContains_success(t *testing.T) {
	body := []byte("This is a content shown.")
	chk := Contains("content")
	if err := chk(nil, body); err != nil {
		t.Fatalf("wanted non nil error, got %v ", err)
	}
}

func TestContains_fail(t *testing.T) {
	body := []byte("This is a content shown.")
	search := "NON_CONTAINS"
	chk := Contains(search)
	var err error
	if err = chk(nil, body); err == nil {
		t.Fatal("wanted error, got nil")
	}
	wantedErrMsg := fmt.Sprintf("contains: body does not contain %s", search)
	if err.Error() != wantedErrMsg {
		t.Fatalf("wanted error message: %q, got %q", wantedErrMsg, err.Error())
	}
}

func TestHasHeaders_ListErrors(t *testing.T) {
	th := http.Header{}
	th.Add(contentTypeHeader, "application/json")
	th.Add(locationHeader, "/users/1")
	returnTh := http.Header{}
	chk := HasHeaders(th)
	resp := &http.Response{
		Header: returnTh,
	}
	err := chk(resp, []byte(""))
	if err == nil {
		t.Fatal("wanted non nil err")
	}
	got := err.Error()
	if !strings.Contains(got, contentTypeHeader) {
		t.Errorf("wanted %s to be present in error", contentTypeHeader)
	}
	if !strings.Contains(got, locationHeader) {
		t.Errorf("wanted %s to be present in error", locationHeader)
	}
}

func TestHasHeaders_Success(t *testing.T) {
	th := http.Header{}
	th.Add(contentTypeHeader, "application/json")
	th.Add(locationHeader, "/users/1")
	returnTh := th
	chk := HasHeaders(th)
	resp := &http.Response{
		Header: returnTh,
	}
	err := chk(resp, []byte(""))
	if err != nil {
		t.Error("wanted nil error")
	}
}

func TestMatchesMD5_success(t *testing.T) {
	wanted := tuxFileMD5Sum
	chk := MatchesMD5(wanted)

	fileData := fileData(t, tuxFilePath)
	err := chk(nil, fileData)
	if err != nil {
		t.Error("no error wanted here.")
	}
}

func TestMatchesMD5_fail(t *testing.T) {
	wanted := "spurious"
	chk := MatchesMD5(wanted)
	fileData := fileData(t, tuxFilePath)
	err := chk(nil, fileData)
	if err == nil {
		t.Fatal("wanted error here. Got nil")
	}
	wantedErr := fmt.Sprintf("matchMD5: wanted %q is not %q", wanted, tuxFileMD5Sum)
	if err.Error() != wantedErr {
		t.Fatalf("wanted error was %q, got %q", wantedErr, err.Error())
	}
}

func fileData(t *testing.T, path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestContainsJSON_success(t *testing.T) {
	wantedJSON := `
	{
		"name" : "test",
		"type": "json"
	}
`
	chk := ContainsJSON(wantedJSON)
	err := chk(nil, []byte(`{"name":"test","type":"json"}`))
	if err != nil {
		t.Errorf("no error wanted here. Was: %q", err.Error())
	}
}

func TestContainsJSON_fail(t *testing.T) {
	wanted := `
	{
		"company" : "test",
		"type": "json"
	}
`
	chk := ContainsJSON(wanted)
	body := `{"name":"test","type":"json"}`
	err := chk(nil, []byte(body))
	if err == nil {
		t.Fatal("wanted error here. Got nil")
	}
	var wantedErr strings.Builder
	wantedErr.WriteString("containsJSON: wanted\n")
	wantedErr.WriteString(fmt.Sprintf("%q", removeSpacesAndTabs(wanted)))
	wantedErr.WriteString("is not\n")
	wantedErr.WriteString(fmt.Sprintf("%q", removeSpacesAndTabs(body)))
	if err.Error() != wantedErr.String() {
		t.Fatalf("wanted error was \n %q \n got \n %q", wantedErr.String(), err.Error())
	}
}
