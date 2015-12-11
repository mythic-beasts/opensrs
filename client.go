package opensrs

import (
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type XCPClient struct {
	url        string
	username   string
	privateKey string
}

type NestedStringMap map[string]interface{}

func serializeItem(k string, v interface{}, w io.Writer) {
	w.Write([]byte(`<item key="`))
	xml.EscapeText(w, []byte(k))
	w.Write([]byte(`">`))
	switch i := v.(type) {
	case string:
		xml.EscapeText(w, []byte(i))
	case NestedStringMap:
		serializeMap(i, w)
	case []NestedStringMap:
		serializeNSMArray(i, w)
	case []string:
		serializeStringArray(i, w)
	}
	w.Write([]byte(`</item>`))
}

func serializeNSMArray(nsm []NestedStringMap, w io.Writer) {
	w.Write([]byte(`<dt_array>`))
	for i, v := range nsm {
		serializeItem(strconv.Itoa(i), v, w)
	}
	w.Write([]byte(`</dt_array>`))
}

func serializeStringArray(nsm []string, w io.Writer) {
	w.Write([]byte(`<dt_array>`))
	for i, v := range nsm {
		serializeItem(strconv.Itoa(i), v, w)
	}
	w.Write([]byte(`</dt_array>`))
}

func serializeMap(nsm NestedStringMap, w io.Writer) {
	w.Write([]byte(`<dt_assoc>`))
	for k, v := range nsm {
		serializeItem(k, v, w)
	}
	w.Write([]byte(`</dt_assoc>`))
}

func writeXMLMessage(nsm NestedStringMap, w io.Writer) {
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone='yes'?>
<!DOCTYPE OPS_envelope SYSTEM 'ops.dtd'>
<OPS_envelope>
<header>
 <version>0.9</version>
 </header>
 <body>
   <data_block>
`))
	serializeMap(nsm, w)

	w.Write([]byte(`</data_block>
</body>
</header>`))
}

func NewXCPClient(url string, username string, privateKey string) *XCPClient {
	xcp := XCPClient{
		url:        url,
		username:   username,
		privateKey: privateKey,
	}
	return &xcp
}

func (c *XCPClient) createSignature(xml string) string {
	x := md5.Sum([]byte(xml + c.privateKey))
	y := md5.Sum([]byte(fmt.Sprintf("%x", x) + c.privateKey))
	return fmt.Sprintf("%x", y)
}

func (c *XCPClient) doRequest(requestNSM NestedStringMap) (response *NestedStringMap, err error) {

	var xmlRequest bytes.Buffer
	writeXMLMessage(requestNSM, &xmlRequest)

	request, err := http.NewRequest("POST", c.url, &xmlRequest)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "text/xml")
	request.Header.Add("X-Username", c.username)
	request.Header.Add("X-Signature", c.createSignature(xmlRequest.String()))
	request.ContentLength = int64(len(xmlRequest.String()))

	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	nsm, err := c.xmlResponseToNSM(res.Body)
	return nsm, err
}

func (n NestedStringMap) getString(key string) (string, bool) {
	path := strings.Split(key, "/")
	last := path[len(path)-1]
	path = path[:len(path)-1]
	for _, item := range path {
		var ok bool
		n, ok = n[item].(NestedStringMap)
		if !ok {
			return "", false
		}
	}
	s, ok := n[last].(string)
	if !ok {
		return "", false
	}
	return s, true
}

func (n NestedStringMap) getInteger(key string) (int, bool) {
	s, ok := n.getString(key)
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return i, true
}

func (n NestedStringMap) getMap(key string) (*NestedStringMap, bool) {
	path := strings.Split(key, "/")
	for _, item := range path {
		var ok bool
		n, ok = n[item].(NestedStringMap)
		if !ok {
			return nil, false
		}
	}
	return &n, true
}

func (c *XCPClient) xmlResponseToNSM(xmlr io.Reader) (*NestedStringMap, error) {

	decoder := xml.NewDecoder(xmlr)

	var root *NestedStringMap
	var stack []*NestedStringMap
	var currentKey string
	var currentItem *NestedStringMap
	var charData string
	simple := false

	for {
		t, xmlerr := decoder.Token()
		if t == nil {
			break
		}
		if xmlerr != nil {
			return nil, xmlerr
		}
		switch se := t.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "dt_assoc", "dt_array":
				assoc := NestedStringMap{}
				if currentItem != nil {
					(*currentItem)[currentKey] = assoc
				}
				if root == nil {
					root = &assoc
				}
				stack = append(stack, &assoc)
				currentItem = &assoc
				simple = false

			case "item":
				charData = ""
				for _, k := range se.Attr {
					if k.Name.Local == "key" {
						currentKey = k.Value
					}
				}
				simple = true
			}
		case xml.EndElement:
			switch se.Name.Local {
			case "dt_assoc", "dt_array":
				stack = stack[:len(stack)-1]
				if len(stack) > 0 {
					currentItem = stack[len(stack)-1]
				} else {
					currentItem = nil
				}
			case "item":
				if simple {
					(*currentItem)[currentKey] = charData
				}
			}
			simple = false
		case xml.CharData:
			if simple {
				charData += string(se)
			}
		}
	}
	if root == nil {
		return nil, errors.New("No associative array found in response")
	}
	return root, nil
}
