package solanaclient

import "encoding/json"

func JsonToBytes(js []byte) []byte {
	var ba []byte
	err := json.Unmarshal(js, &ba)
	if err != nil {
		return nil
	}
	return ba
}
