package opensrs

import (
	"errors"
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
	res, err := c.doRequest(nsm)
	dsr := []DSRecord{}
	if err != nil {
		return dsr, err
	}
	records, ok := res.getMap("attributes/dnssec")
	if !ok {
		return dsr, errors.New("'attributes' not found in response message")
	}
	for k, _ := range *records {
		record, ok := records.getMap(k)
		if !ok {
			return dsr, errors.New("Could not find DS record in response")
		}
		r := DSRecord{}
		r.KeyTag, _ = record.getInteger("key_tag")
		r.Algorithm, _ = record.getInteger("algorithm")
		r.DigestType, _ = record.getInteger("digest_type")
		r.Digest, _ = record.getString("digest")
		dsr = append(dsr, r)
	}

	return dsr, nil
}
