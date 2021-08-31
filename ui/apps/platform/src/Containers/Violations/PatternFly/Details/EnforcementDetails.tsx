import React, { ReactElement } from 'react';
import uniqBy from 'lodash/uniqBy';
import { Flex, FlexItem, Divider } from '@patternfly/react-core';

import LIFECYCLE_STAGES from 'constants/lifecycleStages';
import { ENFORCEMENT_ACTIONS } from 'constants/enforcementActions';
import { Alert } from '../types/violationTypes';
import Header from './Enforcement/Header';
import Explanation from './Enforcement/Explanation';

function getRuntimeEnforcementCount(processViolation) {
    return uniqBy(processViolation.processes, 'podId').length;
}

type EnforcementDetailsProps = {
    alert: Alert;
};

function EnforcementDetails({ alert }: EnforcementDetailsProps): ReactElement {
    const { lifecycleStage, processViolation, enforcement, policy } = alert;
    let enforcementCount = 0;
    if (lifecycleStage === LIFECYCLE_STAGES.RUNTIME) {
        if (enforcement?.action === ENFORCEMENT_ACTIONS.KILL_POD_ENFORCEMENT) {
            enforcementCount =
                enforcement && processViolation?.processes
                    ? getRuntimeEnforcementCount(processViolation)
                    : 0;
        } else {
            enforcementCount = 1;
        }
    } else if (lifecycleStage === LIFECYCLE_STAGES.DEPLOY) {
        enforcementCount = 1;
    }

    return (
        <Flex direction={{ default: 'column' }}>
            <FlexItem>
                <Header
                    lifecycleStage={alert.lifecycleStage}
                    enforcementCount={enforcementCount}
                    enforcementAction={enforcement?.action}
                />
                {enforcement && enforcementCount && (
                    <>
                        <Divider component="div" />
                        <Explanation
                            lifecycleStage={lifecycleStage}
                            enforcement={enforcement}
                            policyId={policy.id}
                        />
                    </>
                )}
            </FlexItem>
        </Flex>
    );
}

export default EnforcementDetails;
