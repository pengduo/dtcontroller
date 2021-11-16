package vmsdk

import (
	"context"
	"log"
	"net/url"

	"github.com/vmware/govmomi"
)

func Vmclient(ctx context.Context, vURL string, username string, password string) (client *govmomi.Client, err error) {
	u, err := url.Parse(vURL)
	if err != nil {
		log.Panicln(err.Error())
		return client, err
	}
	u.User = url.UserPassword(username, password)
	c, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		log.Panicln(err.Error())
		return client, err
	}
	return c, nil
}
