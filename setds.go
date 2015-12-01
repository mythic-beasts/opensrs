package opensrs

import (
	"errors"
	"fmt"
	"strconv"
)

func (c *XCPClient) SetDSRecords(domain string, dsr []DSRecord) error {
	records := []NestedStringMap{}
	for _, r := range dsr {
		records = append(records, NestedStringMap{
			"algorithm":   strconv.Itoa(r.Algorithm),
			"key_tag":     strconv.Itoa(r.KeyTag),
			"digest_type": strconv.Itoa(r.DigestType),
			"digest":      r.Digest,
		})
	}

	nsm := NestedStringMap{
		"protocol": "XCP",
		"object":   "domain",
		"action":   "modify",
		"attributes": NestedStringMap{
			"domain": domain,
			"dnssec": records,
			"data":   "dnssec",
		},
	}

	xml := xmlMessage(nsm)
	res, err := c.doRequest(xml)
	if err != nil {
		return err
	}
	s, _ := res.getString("is_success")
	if s != "1" {
		msg, _ := res.getString("response_text")
		return errors.New("Error: " + msg)
	}

	fmt.Println(res)

	return nil
}
