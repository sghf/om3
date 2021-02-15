package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"opensvc.com/opensvc/core/types"
)

// DaemonStatus fetchs the daemon status structure from the agent api
func (a API) DaemonStatus() (types.DaemonStatus, error) {
	var ds types.DaemonStatus
	resp, err := a.Requester.Get("daemon_status")
	if err != nil {
		fmt.Println(err)
		return ds, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ds, err
	}
	body = bytes.TrimRight(body, "\x00")
	err = json.Unmarshal(body, &ds)
	if err != nil {
		fmt.Println(err)
		return ds, err
	}
	fmt.Println(ds)
	return ds, nil
}
