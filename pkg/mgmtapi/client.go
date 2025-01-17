package mgmtapi

import (
	"context"

	"github.com/k8ssandra/cass-operator/pkg/httphelper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NewManagementClient returns a new instance for management-api go-client
func NewManagementClient(ctx context.Context, client client.Client) (httphelper.NodeMgmtClient, error) {
	logger := log.FromContext(ctx)

	// We don't support authentication yet, so always use insecure
	provider := &httphelper.InsecureManagementApiSecurityProvider{}
	protocol := provider.GetProtocol()

	httpClient, err := provider.BuildHttpClient(client, ctx)
	if err != nil {
		logger.Error(err, "error in BuildManagementApiHttpClient")
		return httphelper.NodeMgmtClient{}, err
	}

	return httphelper.NodeMgmtClient{
		Client:   httpClient,
		Log:      logger,
		Protocol: protocol,
	}, nil
}
