import React, { useEffect, useCallback } from 'react';
import ReactRouterPropTypes from 'react-router-prop-types';
import { generatePath } from 'react-router-dom';
import { useQuery } from '@apollo/client';
import { HashLink } from 'react-router-hash-link';

import PageHeader from 'Components/PageHeader';
import SearchFilterInput from 'Components/SearchFilterInput';
import entityTypes, { searchCategories } from 'constants/entityTypes';
import workflowStateContext from 'Containers/workflowStateContext';
import { SEARCH_OPTIONS_QUERY } from 'queries/search';
import { clustersBasePath, clustersPathWithParam, integrationsPath } from 'routePaths';
import useURLSearch from 'hooks/useURLSearch';
import parseURL from 'utils/URLParser';

import ClustersTablePanel from './ClustersTablePanel';
import ClustersSidePanel from './ClustersSidePanel';

const ClustersPage = ({
    history,
    location: { pathname, search },
    match: {
        params: { clusterId: selectedClusterId },
    },
}) => {
    const { searchFilter, setSearchFilter } = useURLSearch();
    const workflowState = parseURL({ pathname, search });

    // Handle changes to the currently selected deployment.
    const setSelectedClusterId = useCallback(
        (newCluster) => {
            const newClusterId = newCluster?.id || newCluster || '';
            const newWorkflowState = newClusterId
                ? workflowState.pushRelatedEntity(entityTypes.CLUSTER, newClusterId)
                : workflowState.pop();

            const newUrl = newWorkflowState.toUrl();

            history.push(newUrl);
        },
        [workflowState, history]
    );

    const searchQueryOptions = {
        variables: {
            categories: [searchCategories.CLUSTER],
        },
    };
    const { data: searchData } = useQuery(SEARCH_OPTIONS_QUERY, searchQueryOptions);
    const searchOptions = (searchData && searchData.searchOptions) || [];

    // When the selected cluster changes, update the URL.
    useEffect(() => {
        const newPath = selectedClusterId
            ? generatePath(clustersPathWithParam, { clusterId: selectedClusterId })
            : clustersBasePath;
        history.push({
            pathname: newPath,
            search,
        });
    }, [history, search, selectedClusterId]);
    const headerText = 'Clusters';
    const subHeaderText = 'Resource list';

    const pageHeader = (
        <PageHeader header={headerText} subHeader={subHeaderText}>
            <div className="flex flex-1 items-center justify-end">
                <SearchFilterInput
                    className="w-full"
                    searchFilter={searchFilter}
                    searchOptions={searchOptions}
                    searchCategory="CLUSTERS"
                    placeholder="Add one or more filters"
                    handleChangeSearchFilter={setSearchFilter}
                />
                <div className="flex items-center ml-4 mr-3">
                    <HashLink
                        to={`${integrationsPath}#token-integrations`}
                        className="no-underline btn btn-base flex-shrink-0"
                    >
                        Manage Tokens
                    </HashLink>
                </div>
            </div>
        </PageHeader>
    );

    return (
        <workflowStateContext.Provider value={workflowState}>
            <section className="flex flex-1 flex-col h-full">
                <div className="flex flex-1 flex-col">
                    {pageHeader}
                    <div className="flex flex-1 relative">
                        <ClustersTablePanel
                            selectedClusterId={selectedClusterId}
                            setSelectedClusterId={setSelectedClusterId}
                            searchOptions={searchOptions}
                        />
                        <ClustersSidePanel
                            selectedClusterId={selectedClusterId}
                            setSelectedClusterId={setSelectedClusterId}
                        />
                    </div>
                </div>
            </section>
        </workflowStateContext.Provider>
    );
};

ClustersPage.propTypes = {
    history: ReactRouterPropTypes.history.isRequired,
    location: ReactRouterPropTypes.location.isRequired,
    match: ReactRouterPropTypes.match.isRequired,
};

export default ClustersPage;
