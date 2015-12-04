# mythic-beasts.com/opensrs

Go package for interacting with the OpenSRS domain management interface via XCP.

## Example

    xcp := opensrs.NewXCPClient("https://horizon.opensrs.net:55443/", "username", "privateKey")
    res, _ := xcp.GetInfo("example.com")
    fmt.Println("Expiry date" + res.ExpiryDate.Format("2006-01-02T15:04:05"))

## Supported methods

At present only the following methods are supported:

* GetInfo - response object is incompletely populated
* GetDSRecords
* SetDSRecords



