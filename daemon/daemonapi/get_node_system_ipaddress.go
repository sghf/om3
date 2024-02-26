package daemonapi

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/opensvc/om3/core/clusternode"
	"github.com/opensvc/om3/core/object"
	"github.com/opensvc/om3/daemon/api"
)

func (a *DaemonAPI) GetNodeSystemIPAddress(ctx echo.Context, nodename api.InPathNodeName) error {
	if a.localhost == nodename {
		return a.getLocalNodeSystemIPAddress(ctx)
	} else if !clusternode.Has(nodename) {
		return JSONProblemf(ctx, http.StatusBadRequest, "Invalid parameters", "%s is not a cluster node", nodename)
	} else {
		return a.getPeerNodeSystemIPAddress(ctx, nodename)
	}
}

func (a *DaemonAPI) getPeerNodeSystemIPAddress(ctx echo.Context, nodename string) error {
	c, err := newProxyClient(ctx, nodename)
	if err != nil {
		return JSONProblemf(ctx, http.StatusInternalServerError, "New client", "%s: %s", nodename, err)
	} else if !clusternode.Has(nodename) {
		return JSONProblemf(ctx, http.StatusBadRequest, "Invalid nodename", "field 'nodename' with value '%s' is not a cluster node", nodename)
	}
	if resp, err := c.GetNodeSystemIPAddressWithResponse(ctx.Request().Context(), nodename); err != nil {
		return JSONProblemf(ctx, http.StatusInternalServerError, "Request peer", "%s: %s", nodename, err)
	} else if len(resp.Body) > 0 {
		return ctx.JSONBlob(resp.StatusCode(), resp.Body)
	}
	return nil
}

func (a *DaemonAPI) getLocalNodeSystemIPAddress(ctx echo.Context) error {
	n, err := object.NewNode()
	if err != nil {
		return JSONProblemf(ctx, http.StatusInternalServerError, "New node", "%s", err)
	}
	data, err := n.LoadSystem()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return JSONProblemf(ctx, http.StatusNotFound, "Load system cache", "waiting for cached value: %s", err)
		} else {
			return JSONProblemf(ctx, http.StatusInternalServerError, "Load system cache", "%s", err)
		}
	}
	items := make(api.IPAddressItems, 0)
	for key, value := range data.LAN {
		for i := 0; i < len(value); i++ {
			items = append(items, api.IPAddressItem{
				Kind: "IPAddressItem",
				Data: api.IPAddress{
					Mac:            key,
					Address:        value[i].Address,
					FlagDeprecated: value[i].FlagDeprecated,
					Intf:           value[i].Intf,
					Mask:           value[i].Mask,
					Type:           value[i].Type,
				},
				Meta: api.NodeMeta{
					Node: a.localhost,
				},
			})
		}
	}

	return ctx.JSON(http.StatusOK, api.IPAddressList{Kind: "IPAddressList", Items: items})
}