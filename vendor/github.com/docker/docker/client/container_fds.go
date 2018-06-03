package client // import "github.com/docker/docker/client"

import (
	"context"
//	"encoding/json"
	"net/url"
	"strconv"

	"github.com/docker/docker/api/types"
//	"github.com/docker/docker/api/types/filters"
)

// ContainerList returns the list of containers in the docker host.
func (cli *Client) ContainerFDS(ctx context.Context, options types.ContainerFDSOptions) (error) {
	query := url.Values{}

	if options.Policy != 0 {
		query.Set("policy", strconv.Itoa(options.Policy))
		
	}

	println("In the func of ContainerFDS")

	_, err := cli.post(ctx, "/containers/fds", query, nil,nil)
	if err != nil {
		return err
	}

return nil


}
