package oxcmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opensvc/om3/core/client"
	"github.com/opensvc/om3/core/naming"
	"github.com/opensvc/om3/core/objectselector"
	"github.com/opensvc/om3/daemon/api"
)

type (
	CmdKeystoreRemove struct {
		OptsGlobal
		Key string
	}
)

func (t *CmdKeystoreRemove) Run(selector, kind string) error {
	ctx := context.Background()
	c, err := client.New(client.WithURL(t.Server))
	if err != nil {
		return err
	}
	paths, err := objectselector.New(
		selector,
		objectselector.WithClient(c),
	).Expand()
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := t.RunForPath(ctx, c, path); err != nil {
			return err
		}
	}
	return nil
}

func (t *CmdKeystoreRemove) RunForPath(ctx context.Context, c *client.T, path naming.Path) error {
	params := api.DeleteObjectKVStoreEntryParams{
		Key: t.Key,
	}
	response, err := c.DeleteObjectKVStoreEntryWithResponse(ctx, path.Namespace, path.Kind, path.Name, &params)
	if err != nil {
		return err
	}
	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusBadRequest:
		return fmt.Errorf("%s: %s", path, *response.JSON400)
	case http.StatusUnauthorized:
		return fmt.Errorf("%s: %s", path, *response.JSON401)
	case http.StatusForbidden:
		return fmt.Errorf("%s: %s", path, *response.JSON403)
	case http.StatusInternalServerError:
		return fmt.Errorf("%s: %s", path, *response.JSON500)
	default:
		return fmt.Errorf("%s: unexpected response: %s", path, response.Status())
	}
}
