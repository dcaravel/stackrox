package clair

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"bitbucket.org/stack-rox/apollo/generated/api/v1"
	"bitbucket.org/stack-rox/apollo/pkg/images"
	"bitbucket.org/stack-rox/apollo/pkg/logging"
	"bitbucket.org/stack-rox/apollo/pkg/scanners"
	"bitbucket.org/stack-rox/apollo/pkg/urlfmt"
	clairV1 "github.com/coreos/clair/api/v1"
)

const (
	requestTimeout = 10 * time.Second
)

var (
	log          = logging.LoggerForModule()
	errNotExists = errors.New("Layer does not exist")
)

type clair struct {
	client                *http.Client
	endpoint              string
	protoImageIntegration *v1.ImageIntegration
}

func newScanner(protoImageIntegration *v1.ImageIntegration) (*clair, error) {
	endpoint, ok := protoImageIntegration.Config["endpoint"]
	if !ok {
		return nil, errors.New("endpoint parameter must be defined for Clair")
	}

	endpoint, err := urlfmt.FormatURL(endpoint, true, false)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: requestTimeout,
	}
	scanner := &clair{
		client:                client,
		endpoint:              endpoint,
		protoImageIntegration: protoImageIntegration,
	}
	return scanner, nil
}

func (c *clair) sendRequest(method string, values url.Values, pathSegments ...string) ([]byte, int, error) {
	path, err := urlfmt.FullyQualifiedURL(c.endpoint, values, pathSegments...)
	if err != nil {
		return nil, -1, err
	}
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return nil, -1, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// Test initiates a test of the Clair Scanner which verifies that we have the proper scan permissions
func (c *clair) Test() error {
	_, code, err := c.sendRequest("GET", url.Values{}, "v1", "namespaces")
	if err != nil {
		return err
	} else if code != http.StatusOK {
		return fmt.Errorf("Received status code %v, but expected 200", code)
	}
	return nil
}

func (c *clair) retrieveLayerData(layer string) (*clairV1.LayerEnvelope, error) {
	v := url.Values{}
	v.Add("features", "true")
	v.Add("vulnerabilities", "true")
	body, status, err := c.sendRequest("GET", v, "v1", "layers", layer)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, errNotExists
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code %v: %v", status, string(body))
	}
	le := new(clairV1.LayerEnvelope)
	if err := json.Unmarshal(body, &le); err != nil {
		return nil, err
	}
	return le, nil
}

func (c *clair) getLastScanFromV2Metadata(image *v1.Image) (*clairV1.LayerEnvelope, bool) {
	if image.GetMetadata().GetV2() == nil {
		return nil, false
	}
	v2 := image.GetMetadata().GetV2()
	layerEnvelope, err := c.retrieveLayerData(v2.GetDigest())
	if err == nil {
		return layerEnvelope, true
	} else if err != errNotExists {
		log.Error(err)
	}

	// V2 Metadata has latest image layer last so look for images in Clair from most recent to least recent
	layers := v2.GetLayers()
	for i := len(layers) - 1; i >= 0; i-- {
		layerEnvelope, err := c.retrieveLayerData(layers[i])
		if err == nil {
			return layerEnvelope, true
		} else if err != errNotExists {
			log.Error(err)
		}
	}
	return nil, false
}

func (c *clair) getLastScanFromV1Metadata(image *v1.Image) (*clairV1.LayerEnvelope, bool) {
	digest := images.NewDigest(image.GetMetadata().GetRegistrySha()).Digest()
	layerEnvelope, err := c.retrieveLayerData(digest)
	if err == nil {
		return layerEnvelope, true
	} else if err != errNotExists {
		log.Error(err)
	}
	for _, layer := range image.GetMetadata().GetFsLayers() {
		layerEnvelope, err := c.retrieveLayerData(layer)
		if err == nil {
			return layerEnvelope, true
		} else if err != errNotExists {
			log.Error(err)
		}
	}
	return nil, false
}

// GetLastScan retrieves the most recent scan
func (c *clair) GetLastScan(image *v1.Image) (*v1.ImageScan, error) {
	if image == nil || image.GetName().GetRemote() == "" || image.GetName().GetTag() == "" {
		return nil, nil
	}
	le, found := c.getLastScanFromV2Metadata(image)
	if !found {
		le, found = c.getLastScanFromV1Metadata(image)
	}
	if le == nil || le.Layer == nil {
		return nil, fmt.Errorf("No scan data found for image %s", image.GetName().GetFullName())
	}
	return convertLayerToImageScan(image, le), nil
}

// Match decides if the image is contained within this scanner
func (c *clair) Match(image *v1.Image) bool {
	return true
}

func (c *clair) Global() bool {
	return len(c.protoImageIntegration.GetClusters()) == 0
}

func init() {
	scanners.Registry["clair"] = func(integration *v1.ImageIntegration) (scanners.ImageScanner, error) {
		scan, err := newScanner(integration)
		return scan, err
	}
}
