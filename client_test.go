package errplane

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type ClientConfigForTesting struct {
	AppId       string `json:"app_id"`
	Environment string `json:"env"`
	APIKey      string `json:"key"`
}

// make json file test_keys.json
// {app_id: "", env: "", key: ""}
func GetTestClient() *Client {
	var f *os.File
	var err error
	if f, err = os.Open("test_keys.json"); f == nil || err != nil {
		fmt.Printf("Couldn't open test config: %v\n", err)
		return nil
	}
	defer f.Close()

	var config *ClientConfigForTesting
	var dec = json.NewDecoder(f)
	// fmt.Println("File", f, "Decoder", dec)
	if err := dec.Decode(&config); err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return nil
	}

	return NewClient(config.AppId, config.Environment, config.APIKey)
}

func TestClient(t *testing.T) {
	c := GetTestClient()
	if c == nil {
		t.Fatal("no thingie")
	}

	err := c.Post((&Event{
		Name:    "test",
		Value:   100,
		Context: map[string]string{"foo": "bar"},
	}).Line())
	if err != nil {
		t.Errorf("Post failed: %v", err)
	}
}
