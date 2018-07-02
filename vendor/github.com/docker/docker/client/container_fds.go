package client // import "github.com/docker/docker/client"

import (
	"context"
//	"encoding/json"
	"net/url"
	"strconv"
	"github.com/docker/docker/container"
	"github.com/docker/docker/api/types"
//	"github.com/docker/docker/api/types/filters"
)

// ContainerList returns the list of containers in the docker host.
func (cli *Client) ContainerFDS(ctx context.Context, options types.ContainerFDSOptions) ([]container.Container,error) {
	query := url.Values{}

	if options.Policy != 0 {
		query.Set("policy", strconv.Itoa(options.Policy))
		
	}

	println("In the func of ContainerFDS")

	//_, err := cli.post(ctx, "/containers/fds", query, nil,nil)
	resp, err := cli.get(ctx,"/containers/resource",query,nil)
	if err != nil {
		return err
	}else {
		println("POST OK!")
	}
	var containers []container.Container
	err = json.NewDecoder(resp.body).Decode(&containers)
	ensureReaderClosed(resp)
	return containers, err


}
