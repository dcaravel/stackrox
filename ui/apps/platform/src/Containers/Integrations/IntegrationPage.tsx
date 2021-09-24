import React, { ReactElement } from 'react';
import { Link } from 'react-router-dom';
import {
    Button,
    ButtonVariant,
    Flex,
    FlexItem,
    PageSection,
    Title,
    Breadcrumb,
    BreadcrumbItem,
    Divider,
} from '@patternfly/react-core';

import { integrationsPath } from 'routePaths';
import { getIntegrationLabel } from 'Containers/Integrations/utils/integrationUtils';
import PageTitle from 'Components/PageTitle';
import BreadcrumbItemLink from 'Components/BreadcrumbItemLink';
import useIntegrationPermissions from './hooks/useIntegrationPermissions';
import usePageState from './hooks/usePageState';

export type IntegrationPageProps = {
    title: string;
    children: ReactElement | ReactElement[];
};

function IntegrationPage({ title, children }: IntegrationPageProps): ReactElement {
    const permissions = useIntegrationPermissions();
    const {
        pageState,
        params: { source, type, id },
    } = usePageState();
    const typeLabel = getIntegrationLabel(source, type);

    const integrationsListPath = `${integrationsPath}/${source}/${type}`;
    const integrationEditPath = `${integrationsPath}/${source}/${type}/edit/${id as string}`;

    return (
        <>
            <PageTitle title={title} />
            <PageSection variant="light">
                <div className="pf-u-mb-sm">
                    <Breadcrumb>
                        <BreadcrumbItemLink to={integrationsPath}>Integrations</BreadcrumbItemLink>
                        <BreadcrumbItemLink to={integrationsListPath}>
                            {typeLabel}
                        </BreadcrumbItemLink>
                        <BreadcrumbItem isActive>{title}</BreadcrumbItem>
                    </Breadcrumb>
                </div>
                <Flex>
                    <FlexItem>
                        <Title headingLevel="h1">{`${
                            pageState === 'VIEW_DETAILS' ? '' : 'Configure'
                        } ${typeLabel} Integration`}</Title>
                    </FlexItem>
                    {pageState === 'VIEW_DETAILS' && permissions[source].write && (
                        <FlexItem align={{ default: 'alignRight' }}>
                            <Button
                                variant={ButtonVariant.secondary}
                                component={(props) => <Link {...props} to={integrationEditPath} />}
                            >
                                Edit
                            </Button>
                        </FlexItem>
                    )}
                </Flex>
            </PageSection>
            <Divider component="div" />
            {children}
        </>
    );
}

export default IntegrationPage;
