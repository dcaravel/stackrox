//go:build sql_integration

package datastore

import (
	"context"
	"testing"

	dackboxTestUtils "github.com/stackrox/rox/central/dackbox/testutils"
	"github.com/stackrox/rox/central/role/resources"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/dackbox/edges"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/fixtures"
	sacTestUtils "github.com/stackrox/rox/pkg/sac/testutils"
	"github.com/stackrox/rox/pkg/scancomponent"
	"github.com/stackrox/rox/pkg/search"
	pgSearch "github.com/stackrox/rox/pkg/search/postgres"
	"github.com/stretchr/testify/suite"
)

var (
	imageScanOperatingSystem = "crime-stories"

	dontWaitForIndexing = false
	waitForIndexing     = true
)

func TestImageComponentEdgeDatastoreSAC(t *testing.T) {
	suite.Run(t, new(imageComponentEdgeDatastoreSACTestSuite))
}

type imageComponentEdgeDatastoreSACTestSuite struct {
	suite.Suite

	dackboxTestStore dackboxTestUtils.DackboxTestDataStore
	datastore        DataStore

	testContexts map[string]context.Context
}

func (s *imageComponentEdgeDatastoreSACTestSuite) SetupSuite() {
	var err error
	s.dackboxTestStore, err = dackboxTestUtils.NewDackboxTestDataStore(s.T())
	s.Require().NoError(err)
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		pool := s.dackboxTestStore.GetPostgresPool()
		s.datastore, err = GetTestPostgresDataStore(s.T(), pool)
		s.Require().NoError(err)
	} else {
		bleveIndex := s.dackboxTestStore.GetBleveIndex()
		dacky := s.dackboxTestStore.GetDackbox()
		s.datastore, err = GetTestRocksBleveDataStore(s.T(), bleveIndex, dacky)
		s.Require().NoError(err)
	}
	s.testContexts = sacTestUtils.GetNamespaceScopedTestContexts(context.Background(), s.T(), resources.Image)
}

func (s *imageComponentEdgeDatastoreSACTestSuite) TearDownSuite() {
	s.Require().NoError(s.dackboxTestStore.Cleanup(s.T()))
}

func (s *imageComponentEdgeDatastoreSACTestSuite) cleanImageToVulnerabilityGraph(waitForIndexing bool) {
	s.Require().NoError(s.dackboxTestStore.CleanImageToVulnerabilitiesGraph(waitForIndexing))
}

func getComponentID(component *storage.EmbeddedImageScanComponent, os string) string {
	return scancomponent.ComponentID(component.GetName(), component.GetVersion(), os)
}

func getEdgeID(image *storage.Image, component *storage.EmbeddedImageScanComponent, os string) string {
	imageID := image.GetId()
	componentID := getComponentID(component, os)
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		return pgSearch.IDFromPks([]string{imageID, componentID})
	}
	return edges.EdgeID{ParentID: imageID, ChildID: componentID}.ToString()
}

type edgeTestCase struct {
	contextKey        string
	expectedEdgeFound map[string]bool
}

var (
	img1cmp1edge = getEdgeID(fixtures.GetImageSherlockHolmes1(), fixtures.GetEmbeddedImageComponent1x1(), imageScanOperatingSystem)
	img1cmp2edge = getEdgeID(fixtures.GetImageSherlockHolmes1(), fixtures.GetEmbeddedImageComponent1x2(), imageScanOperatingSystem)
	img1cmp3edge = getEdgeID(fixtures.GetImageSherlockHolmes1(), fixtures.GetEmbeddedImageComponent1s2x3(), imageScanOperatingSystem)
	img2cmp3edge = getEdgeID(fixtures.GetImageDoctorJekyll2(), fixtures.GetEmbeddedImageComponent1s2x3(), imageScanOperatingSystem)
	img2cmp4edge = getEdgeID(fixtures.GetImageDoctorJekyll2(), fixtures.GetEmbeddedImageComponent2x4(), imageScanOperatingSystem)
	img2cmp5edge = getEdgeID(fixtures.GetImageDoctorJekyll2(), fixtures.GetEmbeddedImageComponent2x5(), imageScanOperatingSystem)

	fullAccessMap = map[string]bool{
		img1cmp1edge: true,
		img1cmp2edge: true,
		img1cmp3edge: true,
		img2cmp3edge: true,
		img2cmp4edge: true,
		img2cmp5edge: true,
	}
	cluster1WithNamespaceAAccessMap = map[string]bool{
		img1cmp1edge: true,
		img1cmp2edge: true,
		img1cmp3edge: true,
		img2cmp3edge: false,
		img2cmp4edge: false,
		img2cmp5edge: false,
	}
	cluster2WithNamespaceBAccessMap = map[string]bool{
		img1cmp1edge: false,
		img1cmp2edge: false,
		img1cmp3edge: false,
		img2cmp3edge: true,
		img2cmp4edge: true,
		img2cmp5edge: true,
	}
	noAccessMap = map[string]bool{
		img1cmp1edge: false,
		img1cmp2edge: false,
		img1cmp3edge: false,
		img2cmp3edge: false,
		img2cmp4edge: false,
		img2cmp5edge: false,
	}

	testCases = []edgeTestCase{
		{
			contextKey:        sacTestUtils.UnrestrictedReadCtx,
			expectedEdgeFound: fullAccessMap,
		},
		{
			contextKey:        sacTestUtils.UnrestrictedReadWriteCtx,
			expectedEdgeFound: fullAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster1ReadWriteCtx,
			expectedEdgeFound: cluster1WithNamespaceAAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster1NamespaceAReadWriteCtx,
			expectedEdgeFound: cluster1WithNamespaceAAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster1NamespaceBReadWriteCtx,
			expectedEdgeFound: noAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster1NamespacesABReadWriteCtx,
			expectedEdgeFound: cluster1WithNamespaceAAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster1NamespacesBCReadWriteCtx,
			expectedEdgeFound: noAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster2ReadWriteCtx,
			expectedEdgeFound: cluster2WithNamespaceBAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster2NamespaceAReadWriteCtx,
			expectedEdgeFound: noAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster2NamespaceBReadWriteCtx,
			expectedEdgeFound: cluster2WithNamespaceBAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster2NamespacesABReadWriteCtx,
			expectedEdgeFound: cluster2WithNamespaceBAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster2NamespacesACReadWriteCtx,
			expectedEdgeFound: noAccessMap,
		},
		{
			contextKey:        sacTestUtils.Cluster3ReadWriteCtx,
			expectedEdgeFound: noAccessMap,
		},
		{
			contextKey: sacTestUtils.MixedClusterAndNamespaceReadCtx,
			// Has access to Cluster1 + NamespaceA as well as full access to Cluster2 (including NamespaceB).
			expectedEdgeFound: fullAccessMap,
		},
	}
)

func (s *imageComponentEdgeDatastoreSACTestSuite) TestExistsEdge() {
	// Inject the fixture graph and test for image1 to component1 edge
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(dontWaitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(dontWaitForIndexing)
	s.Require().NoError(err)

	targetEdgeID := img1cmp1edge
	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		exists, err := s.datastore.Exists(ctx, targetEdgeID)
		s.NoError(err)
		s.Equal(c.expectedEdgeFound[targetEdgeID], exists)
	}
}

func (s *imageComponentEdgeDatastoreSACTestSuite) TestGetEdge() {
	// Inject the fixtures graph and fetch the image1 to component1 edge
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(dontWaitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(dontWaitForIndexing)
	s.Require().NoError(err)

	targetEdgeID := img1cmp1edge
	expectedSrcID := fixtures.GetImageSherlockHolmes1().GetId()
	expectedDstID := getComponentID(fixtures.GetEmbeddedImageComponent1x1(), imageScanOperatingSystem)
	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		fetched, found, err := s.datastore.Get(ctx, targetEdgeID)
		s.NoError(err)
		if c.expectedEdgeFound[targetEdgeID] {
			s.True(found)
			s.Require().NotNil(fetched)
			s.Equal(expectedSrcID, fetched.GetImageId())
			s.Equal(expectedDstID, fetched.GetImageComponentId())
		} else {
			s.False(found)
			s.Nil(fetched)
		}
	}
}

func (s *imageComponentEdgeDatastoreSACTestSuite) TestGetBatch() {
	// Inject the fixtures graph and fetch the image1 to component1 and image2 to component 4 edges
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(dontWaitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(dontWaitForIndexing)
	s.Require().NoError(err)

	targetEdge1ID := img1cmp1edge
	expectedSrc1ID := fixtures.GetImageSherlockHolmes1().GetId()
	expectedDst1ID := getComponentID(fixtures.GetEmbeddedImageComponent1x1(), imageScanOperatingSystem)
	targetEdge2ID := img2cmp4edge
	expectedSrc2ID := fixtures.GetImageDoctorJekyll2().GetId()
	expectedDst2ID := getComponentID(fixtures.GetEmbeddedImageComponent2x4(), imageScanOperatingSystem)
	toFetch := []string{targetEdge1ID, targetEdge2ID}
	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		fetched, err := s.datastore.GetBatch(ctx, toFetch)
		s.NoError(err)
		expectedFetchedSize := 0
		if c.expectedEdgeFound[targetEdge1ID] {
			expectedFetchedSize++
		}
		if c.expectedEdgeFound[targetEdge2ID] {
			expectedFetchedSize++
		}
		fetchedMatches := 0
		s.Equal(expectedFetchedSize, len(fetched))
		for _, edge := range fetched {
			if edge.GetId() == targetEdge1ID {
				fetchedMatches++
				s.Equal(expectedSrc1ID, edge.GetImageId())
				s.Equal(expectedDst1ID, edge.GetImageComponentId())
			}
			if edge.GetId() == targetEdge2ID {
				fetchedMatches++
				s.Equal(expectedSrc2ID, edge.GetImageId())
				s.Equal(expectedDst2ID, edge.GetImageComponentId())
			}
		}
		s.Equal(expectedFetchedSize, fetchedMatches)
	}
}

func (s *imageComponentEdgeDatastoreSACTestSuite) TestCount() {
	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("graph.Context wrapper missing in ImageComponentEdge searcher",
			"to enable Search test case in non-postgres mode")
	}
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(waitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(waitForIndexing)
	s.Require().NoError(err)

	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		expectedCount := 0
		for _, visible := range c.expectedEdgeFound {
			if visible {
				expectedCount++
			}
		}
		count, err := s.datastore.Count(ctx)
		s.NoError(err)
		s.Equal(expectedCount, count)
	}
}

func (s *imageComponentEdgeDatastoreSACTestSuite) TestSearch() {
	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("graph.Context wrapper missing in ImageComponentEdge searcher",
			"to enable Search test case in non-postgres mode")
	}
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(waitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(waitForIndexing)
	s.Require().NoError(err)

	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		expectedCount := 0
		for _, visible := range c.expectedEdgeFound {
			if visible {
				expectedCount++
			}
		}
		results, err := s.datastore.Search(ctx, search.EmptyQuery())
		s.NoError(err)
		s.Equal(expectedCount, len(results))
		for _, r := range results {
			s.True(c.expectedEdgeFound[r.ID])
		}
	}
}

func (s *imageComponentEdgeDatastoreSACTestSuite) TestSearchEdges() {
	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("graph.Context wrapper missing in ImageComponentEdge searcher",
			"to enable Search test case in non-postgres mode")
	}
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(waitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(waitForIndexing)
	s.Require().NoError(err)

	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		expectedCount := 0
		for _, visible := range c.expectedEdgeFound {
			if visible {
				expectedCount++
			}
		}
		results, err := s.datastore.SearchEdges(ctx, search.EmptyQuery())
		s.NoError(err)
		s.Equal(expectedCount, len(results))
		for _, r := range results {
			s.True(c.expectedEdgeFound[r.GetId()])
		}
	}

}

func (s *imageComponentEdgeDatastoreSACTestSuite) TestSearchRawEdges() {
	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("graph.Context wrapper missing in ImageComponentEdge searcher",
			"to enable Search test case in non-postgres mode")
	}
	err := s.dackboxTestStore.PushImageToVulnerabilitiesGraph(waitForIndexing)
	defer s.cleanImageToVulnerabilityGraph(waitForIndexing)
	s.Require().NoError(err)

	for _, c := range testCases {
		ctx := s.testContexts[c.contextKey]
		expectedCount := 0
		for _, visible := range c.expectedEdgeFound {
			if visible {
				expectedCount++
			}
		}
		results, err := s.datastore.SearchRawEdges(ctx, search.EmptyQuery())
		s.NoError(err)
		s.Equal(expectedCount, len(results))
		for _, r := range results {
			s.True(c.expectedEdgeFound[r.GetId()])
		}
	}

}
