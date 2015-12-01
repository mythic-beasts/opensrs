package opensrs

import (
	"fmt"
)

type DSRecord struct {
	KeyTag     int
	Algorithm  int
	DigestType int
	Digest     string
}

func (c *XCPClient) GetDSRecords(domain string) ([]DSRecord, error) {
	nsm := NestedStringMap{
		"protocol": "XCP",
		"object":   "DOMAIN",
		"action":   "GET",
		"attributes": NestedStringMap{
			"domain": domain,
			"type":   "dnssec",
		},
	}
	xml := xmlMessage(nsm)
	res, err := c.doRequest(xml)
	if err != nil {
		return []DSRecord{}, err
	}

	fmt.Println(res)

	return []DSRecord{}, nil
}
