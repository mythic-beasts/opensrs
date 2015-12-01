package opensrs

import (
	"fmt"
	"time"
)

type GetInfoResponse struct {
	ExpiryDate time.Time
}

func (c *XCPClient) GetInfo(domain string) (string, error) {
	nsm := NestedStringMap{
		"protocol": "XCP",
		"object":   "DOMAIN",
		"action":   "GET",
		"attributes": NestedStringMap{
			"domain": domain,
			"type":   "all_info",
		},
	}
	xml := xmlMessage(nsm)
	res, err := c.doRequest(xml)
	if err != nil {
		return "", err
	}

	expiry, _ := res.getString("attributes/expiredate")
	fmt.Println("Expiry: " + expiry)
	return "", nil
}
