package opensrs

import (
	"errors"
	"time"
)

type GetInfoResponse struct {
	ExpiryDate time.Time
}

func (c *XCPClient) GetInfo(domain string) (*GetInfoResponse, error) {
	nsm := NestedStringMap{
		"protocol": "XCP",
		"object":   "DOMAIN",
		"action":   "GET",
		"attributes": NestedStringMap{
			"domain": domain,
			"type":   "all_info",
		},
	}
	res, err := c.doRequest(nsm)

	if err != nil {
		return nil, err
	}

	expiryString, found := res.getString("attributes/expiredate")
	if !found {
		return nil, errors.New("Expiry date not found in info response")
	}

	expiry, err := time.Parse("2006-01-02 15:04:05", expiryString)
	if err != nil {
		return nil, err
	}

	gir := GetInfoResponse{
		ExpiryDate: expiry,
	}
	return &gir, nil
}
