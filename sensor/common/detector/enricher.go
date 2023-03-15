package detector

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v3"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/generated/internalapi/central"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/booleanpolicy/augmentedobjs"
	"github.com/stackrox/rox/pkg/concurrency"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/expiringcache"
	"github.com/stackrox/rox/pkg/features"
	"github.com/stackrox/rox/pkg/images/types"
	"github.com/stackrox/rox/pkg/protoutils"
	registryTypes "github.com/stackrox/rox/pkg/registries/types"
	"github.com/stackrox/rox/pkg/set"
	"github.com/stackrox/rox/sensor/common/clusterid"
	"github.com/stackrox/rox/sensor/common/detector/metrics"
	"github.com/stackrox/rox/sensor/common/imagecacheutils"
	"github.com/stackrox/rox/sensor/common/registry"
	"github.com/stackrox/rox/sensor/common/scan"
	"github.com/stackrox/rox/sensor/common/store"
	"google.golang.org/grpc/status"
)

const (
	scanTimeout = 6 * time.Minute
)

type scanResult struct {
	action                 central.ResourceAction
	deployment             *storage.Deployment
	images                 []*storage.Image
	networkPoliciesApplied *augmentedobjs.NetworkPoliciesApplied
}

type imageChanResult struct {
	image        *storage.Image
	containerIdx int
}

type enricher struct {
	imageSvc       v1.ImageServiceClient
	scanResultChan chan scanResult

	serviceAccountStore store.ServiceAccountStore
	imageCache          expiringcache.Cache
	stopSig             concurrency.Signal
}

type cacheValue struct {
	signal concurrency.Signal
	image  *storage.Image
}

func (c *cacheValue) waitAndGet() *storage.Image {
	c.signal.Wait()
	return c.image
}

func scanImage(ctx context.Context, svc v1.ImageServiceClient, req *scanImageRequest) (*v1.ScanImageInternalResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, scanTimeout)
	defer cancel()

	internalReq := &v1.ScanImageInternalRequest{
		Image: req.containerImage,
	}
	if features.SourcedAutogeneratedIntegrations.Enabled() {
		internalReq.Source = &v1.ScanImageInternalRequest_Source{
			ClusterId:        req.clusterID,
			Namespace:        req.namespace,
			ImagePullSecrets: req.pullSecrets,
		}
	}

	return svc.ScanImageInternal(ctx, internalReq)
}

func scanImageLocal(ctx context.Context, svc v1.ImageServiceClient, req *scanImageRequest) (*v1.ScanImageInternalResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, scanTimeout)
	defer cancel()

	var img *storage.Image
	var err error
	if req.containerImage.GetIsClusterLocal() {
		img, err = scan.EnrichLocalImage(ctx, svc, req.containerImage)
	} else {
		// ForceLocalImageScanning must be enabled
		var reg registryTypes.Registry
		regStore := registry.Singleton()
		imgName := req.containerImage.GetName()

		reg, err = regStore.GetRegistryForImageInNamespace(imgName, req.namespace)
		if err != nil {
			// no registry was found, assume this image represents a registry that does not require authentication.
			// add the registry to regStore and use it for scanning going forward
			reg, err = regStore.UpsertNoAuthRegistry(ctx, req.namespace, imgName)
			if err != nil {
				return nil, err
			}
		}

		img, err = scan.EnrichLocalImageFromRegistry(ctx, svc, req.containerImage, reg)
	}

	return &v1.ScanImageInternalResponse{
		Image: img,
	}, err
}

type scanFunc func(ctx context.Context, svc v1.ImageServiceClient, req *scanImageRequest) (*v1.ScanImageInternalResponse, error)

func scanWithRetries(ctx context.Context, svc v1.ImageServiceClient, req *scanImageRequest, scan scanFunc) (*v1.ScanImageInternalResponse, error) {
	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = 5 * time.Second
	eb.Multiplier = 2
	eb.MaxInterval = 4 * time.Minute
	eb.MaxElapsedTime = 0 // Never stop the backoff, leave that decision to the parent context.

	eb.Reset()

outer:
	for {
		// We want to get the time spent in backoff without including the time it took to scan the image.
		timeSpentInBackoffSoFar := eb.GetElapsedTime()
		scannedImage, err := scan(ctx, svc, req)
		if err != nil {
			for _, detail := range status.Convert(err).Details() {
				// If the client is effectively rate-limited, backoff and try again.
				if _, isTooManyParallelScans := detail.(*v1.ScanImageInternalResponseDetails_TooManyParallelScans); isTooManyParallelScans {
					time.Sleep(eb.NextBackOff())
					continue outer
				}
			}

			return nil, err
		}

		metrics.ObserveTimeSpentInExponentialBackoff(timeSpentInBackoffSoFar)

		return scannedImage, nil
	}
}

func (c *cacheValue) scanAndSet(ctx context.Context, svc v1.ImageServiceClient, req *scanImageRequest) {
	defer c.signal.Signal()

	// Ask Central to scan the image if the image is not internal and local scanning is not forced
	// Otherwise, attempt to scan locally.
	scanImageFn := scanImage
	if req.containerImage.GetIsClusterLocal() || env.ForceLocalImageScanning.BooleanSetting() {
		scanImageFn = scanImageLocal
	}

	scannedImage, err := scanWithRetries(ctx, svc, req, scanImageFn)
	if err != nil {
		// Ignore the error and set the image to something basic,
		// so alerting can progress.
		c.image = types.ToImage(req.containerImage)
		return
	}

	c.image = scannedImage.GetImage()
}

func newEnricher(cache expiringcache.Cache, serviceAccountStore store.ServiceAccountStore) *enricher {
	return &enricher{
		scanResultChan:      make(chan scanResult),
		serviceAccountStore: serviceAccountStore,
		imageCache:          cache,
		stopSig:             concurrency.NewSignal(),
	}
}

func (e *enricher) getImageFromCache(key string) (*storage.Image, bool) {
	value, _ := e.imageCache.Get(key).(*cacheValue)
	if value == nil {
		return nil, false
	}
	return value.waitAndGet(), true
}

func (e *enricher) runScan(req *scanImageRequest) imageChanResult {
	// Cache key is either going to be image full name or image ID.
	// In case of image full name, we can skip. In case of image ID, we should make sure to check if the image's name
	// is equal / contained in the images `Names` field.
	key := imagecacheutils.GetImageCacheKey(req.containerImage)

	// If the container image says that the image is not pullable, don't even bother trying to scan
	if req.containerImage.GetNotPullable() {
		return imageChanResult{
			image:        types.ToImage(req.containerImage),
			containerIdx: req.containerIdx,
		}
	}

	// forceEnrichImageWithSignatures will be set to true in case we have an image where a cached value exists for the
	// digest, but the name has not been added to the "Names" field. In this case, we will force a
	// re-scan of the image, which should only fetch & verify signatures (since we already have a scan
	// result associated, this should not matter.
	var forceEnrichImageWithSignatures bool

	img, ok := e.getImageFromCache(key)
	if ok {
		log.Debugf("cache hit for image %q", req.containerImage.GetName())
		// If the container image name is already within the cached images names, we can short-circuit.
		if protoutils.SliceContains(req.containerImage.GetName(), img.GetNames()) {
			log.Debugf("returning cached entry for image %q", req.containerImage.GetName())
			return imageChanResult{
				image:        img,
				containerIdx: req.containerIdx,
			}
		}
		log.Debugf("skipping cached entry for image %q", req.containerImage.GetName())
		// We found an image that is already in cache (i.e. with the same digest), but the image name is different.
		// Ensuring we have a fully enriched image (especially regarding image signatures), we need to make sure to
		// scan this image once more. This should result in the signatures + signature verification being re-done.
		forceEnrichImageWithSignatures = true
	}

	newValue := &cacheValue{
		signal: concurrency.NewSignal(),
	}
	value := e.imageCache.GetOrSet(key, newValue).(*cacheValue)
	if forceEnrichImageWithSignatures || newValue == value {
		value.scanAndSet(concurrency.AsContext(&e.stopSig), e.imageSvc, req)
	}

	result := imageChanResult{
		image:        value.waitAndGet(),
		containerIdx: req.containerIdx,
	}

	ns := ""
	if req != nil {
		ns = req.namespace
	}

	fn := ""
	if result.image != nil && result.image.GetName() != nil {
		fn = result.image.GetName().GetFullName()
	} else {
		log.Debugf("image full name is nil, %+v", img)
	}

	cacheSize := ""
	if e.imageCache != nil {
		cacheSize = fmt.Sprintf("%d", len(e.imageCache.GetAll()))
	}

	log.Debugf("image scan has finished from %q, img %q, cache len %q", ns, fn, cacheSize)
	return result
}

type scanImageRequest struct {
	containerIdx         int
	containerImage       *storage.ContainerImage
	clusterID, namespace string
	pullSecrets          []string
}

func (e *enricher) runImageScanAsync(imageChan chan<- imageChanResult, req *scanImageRequest) {
	go func() {
		// unguarded send (push to channel outside a select) is allowed because the imageChan is a buffered channel of exact size
		imageChan <- e.runScan(req)
	}()
}

func (e *enricher) getImages(deployment *storage.Deployment) []*storage.Image {
	imageChan := make(chan imageChanResult, len(deployment.GetContainers()))

	var pullSecrets []string
	if features.SourcedAutogeneratedIntegrations.Enabled() {
		pullSecretsSet := set.NewStringSet(e.serviceAccountStore.GetImagePullSecrets(deployment.GetNamespace(), deployment.GetServiceAccount())...)
		pullSecretsSet.AddAll(deployment.GetImagePullSecrets()...)
		pullSecrets = pullSecretsSet.AsSlice()
	}
	for idx, container := range deployment.GetContainers() {
		e.runImageScanAsync(imageChan, &scanImageRequest{
			containerIdx:   idx,
			containerImage: container.GetImage(),
			clusterID:      clusterid.Get(),
			namespace:      deployment.GetNamespace(),
			pullSecrets:    pullSecrets,
		})
	}
	images := make([]*storage.Image, len(deployment.GetContainers()))
	for i := 0; i < len(deployment.GetContainers()); i++ {
		imgResult := <-imageChan

		// This will ensure that when we change the Name of the image
		// that it will not cause a potential race condition
		// cloning the full object is too expensive and also unnecessary
		image := *imgResult.image
		// Overwrite the image Name as a workaround to the fact that we fetch the image by ID
		// The ID may actually have many names that refer to it. e.g. busybox:latest and busybox:1.31 could have the
		// exact same ID
		image.Name = deployment.Containers[imgResult.containerIdx].GetImage().GetName()
		images[imgResult.containerIdx] = &image
	}
	return images
}

func (e *enricher) blockingScan(deployment *storage.Deployment, netpolApplied *augmentedobjs.NetworkPoliciesApplied, action central.ResourceAction) {
	select {
	case <-e.stopSig.Done():
		return
	case e.scanResultChan <- scanResult{
		action:                 action,
		deployment:             deployment,
		images:                 e.getImages(deployment),
		networkPoliciesApplied: netpolApplied,
	}:
	}
}

func (e *enricher) outputChan() <-chan scanResult {
	return e.scanResultChan
}

func (e *enricher) stop() {
	e.stopSig.Signal()
}
