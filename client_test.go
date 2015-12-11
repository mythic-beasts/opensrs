package opensrs

import (
	"bytes"
	"testing"
)

func testSerializeMap(t *testing.T, nsm NestedStringMap, ctrl string) {
	var buf bytes.Buffer
	serializeMap(nsm, &buf)
	if ctrl != buf.String() {
		t.Logf("Got: \n%s\n Expected: \n%s\n", buf.String(), ctrl)
		t.Fail()
	}
}

func TestSerialiseNSM(t *testing.T) {
	nsm := NestedStringMap{
		"foo": "bar",
	}
	ctrl := `<dt_assoc><item key="foo">bar</item></dt_assoc>`
	testSerializeMap(t, nsm, ctrl)

	nsm = NestedStringMap{
		"blort": NestedStringMap{
			"foo": "bar",
		},
	}
	ctrl = `<dt_assoc><item key="blort"><dt_assoc><item key="foo">bar</item></dt_assoc></item></dt_assoc>`
	testSerializeMap(t, nsm, ctrl)

	nsm = NestedStringMap{
		"blort": []NestedStringMap{
			NestedStringMap{
				"foo": "bar",
			},
		},
	}
	ctrl = `<dt_assoc><item key="blort"><dt_array><item key="0"><dt_assoc><item key="foo">bar</item></dt_assoc></item></dt_array></item></dt_assoc>`
	testSerializeMap(t, nsm, ctrl)

	nsm = NestedStringMap{
		"blort": []string{
			"fish",
			"soup",
		},
	}
	ctrl = `<dt_assoc><item key="blort"><dt_array><item key="0">fish</item><item key="1">soup</item></dt_array></item></dt_assoc>`
	testSerializeMap(t, nsm, ctrl)
}

func TestDeserialiseNSMSimple(t *testing.T) {
	c := XCPClient{}
	xml := `<dt_assoc><item key="foo">bar</item><item key="baz">37</item></dt_assoc>`
	nsm, err := c.xmlResponseToNSM(bytes.NewBufferString(xml))
	if err != nil {
		t.Error("Error decoding XML" + err.Error())
	}

	if v, ok := nsm.getString("foo"); !ok || v != "bar" {
		t.Error("Error getting string 'foo'")
	}
	if _, ok := nsm.getString("fish"); ok {
		t.Error("String 'fish' unexpectedly present")
	}
	if v, ok := nsm.getInteger("baz"); !ok || v != 37 {
		t.Error("Error getting int 'baz'")
	}

}

func TestDeserialiseNSMNested(t *testing.T) {
	c := XCPClient{}
	xml := `<dt_assoc>`
	xml += `<item key="blort"><dt_assoc><item key="foo">bar</item></dt_assoc></item>`
	xml += `<item key="wibble"><dt_array><item key="0">FirstItem</item><item key="1">SecondItem</item></dt_array></item>`
	xml += `</dt_assoc>`
	nsm, err := c.xmlResponseToNSM(bytes.NewBufferString(xml))
	if err != nil {
		t.Error("Error decoding XML" + err.Error())
	}

	if v, ok := nsm.getString("blort/foo"); !ok || v != "bar" {
		t.Error("Error getting string 'blort/foo'")
	}
	if _, ok := nsm.getString("blort/fish"); ok {
		t.Error("String 'fish' unexpectedly present")
	}
	if v, ok := nsm.getString("wibble/0"); !ok || v != "FirstItem" {
		t.Error("Error getting string 'wibble/0'")
	}
	if v, ok := nsm.getString("wibble/1"); !ok || v != "SecondItem" {
		t.Error("Error getting string 'wibble/1'")
	}

}
