import React from 'react';
import { Link, withRouter } from 'react-router-dom';
import PropTypes from 'prop-types';
import URLService from 'modules/URLService';
import ReactRouterPropTypes from 'react-router-prop-types';

const PageNotFound = ({ match, location, resourceType }) => {
    const homeUrl = URLService.getURL(match, location)
        .base()
        .url();

    const resourceTypeName = (resourceType || 'resource').toLowerCase();
    return (
        <div className="text-center flex w-full justify-center items-center py-32 px-3 min-h-full bg-primary-200">
            <div>
                <h2 className="text-tertiary-800 mb-2">
                    {`Unfortunately, the ${resourceTypeName} you are looking for cannot be found.`}
                </h2>
                <p className="text-tertiary-800 mb-8">
                    {`It may have changed, did not exist, or no longer exists. Try using search from the dashboard to find what you're looking for.`}
                </p>
                <Link
                    className="p-4 text-uppercase text-base-100 focus:text-base-100 hover:text-base-100 bg-tertiary-700 hover:bg-tertiary-800 no-underline focus:bg-tertiary-800 inline-block text-center rounded-sm"
                    to={homeUrl}
                >
                    {`Go to dashboard`}
                </Link>
            </div>
        </div>
    );
};

PageNotFound.propTypes = {
    resourceType: PropTypes.string,
    match: ReactRouterPropTypes.match.isRequired,
    location: ReactRouterPropTypes.location.isRequired
};

PageNotFound.defaultProps = {
    resourceType: null
};

export default withRouter(PageNotFound);
