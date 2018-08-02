package dao

import "github.com/gocql/gocql"

func GetUserPerms(
	session *gocql.Session,
	email string) []string {
	querystring := "select permission from user_permissions where email=?"
	iter := session.Query(querystring, email).Iter()
	var perms []string
	var perm string
	for iter.Scan(&perm) {
		perms = append(perms, perm)
	}
	return perms
}
