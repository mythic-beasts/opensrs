package opensrs

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type XCPClient struct {
	url        string
	username   string
	privateKey string
}

type NestedStringMap map[string]interface{}

func serializeItem(k string, v interface{}) string {
	s := fmt.Sprintf(`<item key="%s">`, k)
	switch i := v.(type) {
	case string:
		s += i
	case NestedStringMap:
		s += serializeMap(i)
	case []NestedStringMap:
		s += serializeArray(i)
	}
	s += `</item>`
	return s
}

func serializeArray(nsm []NestedStringMap) string {
	s := `<dt_array>`
	for i, v := range nsm {
		s += serializeItem(strconv.Itoa(i), v)
	}
	s += `</dt_array>`
	return s
}

func serializeMap(nsm NestedStringMap) string {
	s := `<dt_assoc>`
	for k, v := range nsm {
		s += serializeItem(k, v)
	}
	s += `</dt_assoc>`
	return s
}

func xmlMessage(nsm NestedStringMap) string {
	s := `<?xml version="1.0" encoding="UTF-8" standalone='yes'?>
<!DOCTYPE OPS_envelope SYSTEM 'ops.dtd'>
<OPS_envelope>
<header>
 <version>0.9</version>
 </header>
 <body>
   <data_block>
`
	s += serializeMap(nsm)

	s += `</data_block>
</body>
</header>`
	return s
}

func NewXCPClient(url string, username string, privateKey string) XCPClient {
	xcp := XCPClient{
		url:        url,
		username:   username,
		privateKey: privateKey,
	}
	return xcp
}

func (c *XCPClient) createSignature(xml string) string {
	x := md5.Sum([]byte(xml + c.privateKey))
	y := md5.Sum([]byte(fmt.Sprintf("%x", x) + c.privateKey))
	fmt.Printf("MD5 %x\n", y)
	return fmt.Sprintf("%x", y)
}

func (c *XCPClient) doRequest(xmlrequest string) (response *NestedStringMap, err error) {

	request, err := http.NewRequest("POST", c.url, strings.NewReader(xmlrequest))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "text/xml")
	request.Header.Add("X-Username", c.username)
	request.Header.Add("X-Signature", c.createSignature(xmlrequest))
	request.ContentLength = int64(len(xmlrequest))
	request.Header.Write(os.Stdout)

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
			case "dt_assoc":
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
			case "dt_assoc":
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
	fmt.Println(*root)
	return root, nil
}
