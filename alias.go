package main

import "strings"

const aliasesKey string = "aliases"

type User string

func GetUserByAlias(user string) User {
	result := Redis.HGet(aliasesKey, strings.ToLower(user))
	if result.Err() != nil {
		return User(user)
	}

	return User(result.Val())
}

func RefreshAliases() error {
	entity, err := DevDialog.EntitiesFindByIdRequest("fdb3a3e8-3963-4d11-b83c-5c97f7fd2976")
	if err != nil {
		return err
	}

	result := Redis.Del(aliasesKey)
	if result.Err() != nil {
		return result.Err()
	}

	for _, entry := range entity.Entries {
		for _, syn := range entry.Synonyms {
			r := Redis.HSet(aliasesKey, strings.ToLower(syn), entry.Value)
			if r.Err() != nil {
				return r.Err()
			}
		}
	}

	return nil
}
