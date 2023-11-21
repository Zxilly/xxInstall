// +build windows

package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows"
)

func isRoot() bool {
	var sid *windows.SID

	err := windows.AllocateAndInitializeSid(
			&windows.SECURITY_NT_AUTHORITY,
			2,
			windows.SECURITY_BUILTIN_DOMAIN_RID,
			windows.DOMAIN_ALIAS_RID_ADMINS,
			0, 0, 0, 0, 0, 0,
			&sid)
	if err != nil {
			log.Fatalf("SID Error: %s", err)
			return
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
			log.Fatalf("Token Membership Error: %s", err)
			return
	}

	return member
}