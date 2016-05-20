package opensrs

import (
	"errors"
	"fmt"
)

func (c *XCPClient) GetLockState(domain string) (bool, error) {
	nsm := NestedStringMap{
		"protocol": "XCP",
		"object":   "DOMAIN",
		"action":   "GET",
		"attributes": NestedStringMap{
			"domain": domain,
			"type":   "status",
		},
	}
	res, err := c.doRequest(nsm)

	if err != nil {
		return false, err
	}
	fmt.Println(*res)
	lockString, found := res.getString("attributes/lock_state")
	if !found {
		return false, errors.New("lock_state not found in response")
	}

	var isLocked bool
	if lockString == "0" {
		isLocked = false
	} else if lockString == "1" {
		isLocked = true
	} else {
		return false, errors.New("Cannot parse lock_state")
	}
	return isLocked, nil
}

func (c *XCPClient) SetLockState(domain string, state bool) error {
	stateStr := "0"
	if state {
		stateStr = "1"
	}

	nsm := NestedStringMap{
		"protocol": "XCP",
		"object":   "domain",
		"action":   "modify",
		"attributes": NestedStringMap{
			"domain":     domain,
			"lock_state": stateStr,
			"data":       "status",
		},
	}

	res, err := c.doRequest(nsm)
	if err != nil {
		return err
	}
	s, _ := res.getString("is_success")
	if s != "1" {
		msg, _ := res.getString("response_text")
		return errors.New("Error: " + msg)
	}

	return nil
}
