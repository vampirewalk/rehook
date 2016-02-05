package main

import (
	"fmt"
	"github.com/keybase/go-keychain"
)

func queryToken(service, account string) (token string, err error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(service)
	query.SetAccount(account)
	//query.SetAccessGroup(accessGroup)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		// Error
		return "", err
	} else if len(results) != 1 {
		// Not found
		return "", fmt.Errorf("Not Found")
	} else {
		password := string(results[0].Data)
		return password, nil
	}
}
