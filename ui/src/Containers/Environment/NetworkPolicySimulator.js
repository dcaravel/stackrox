import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Panel from 'Components/Panel';
import * as Icon from 'react-feather';
import gettingStarted from 'images/getting-started.svg';
import ReactDropzone from 'react-dropzone';
import Loader from 'Components/Loader';
import { actions as notificationActions } from 'reducers/notifications';
import { connect } from 'react-redux';

import SendNotificationSection from 'Containers/Environment/SendNotificationSection';
import NetworkPolicySimulatorSuccessView from './NetworkPolicySimulatorSuccessView';
import NetworkPolicySimulatorErrorView from './NetworkPolicySimulatorErrorView';

class NetworkPolicySimulator extends Component {
    static propTypes = {
        onClose: PropTypes.func.isRequired,
        onYamlUpload: PropTypes.func.isRequired,
        yamlUploadState: PropTypes.string.isRequired,
        addToast: PropTypes.func.isRequired,
        removeToast: PropTypes.func.isRequired,
        errorMessage: PropTypes.string.isRequired,
        yamlFile: PropTypes.shape({
            content: PropTypes.string,
            name: PropTypes.string
        })
    };

    static defaultProps = {
        yamlFile: null
    };

    state = {
        showGetStartedSection: true,
        showDragAndDrop: true
    };

    onClose = () => {
        this.props.onClose();
    };

    onDrop = acceptedFiles => {
        acceptedFiles.forEach(file => {
            if (file && !file.name.includes('.yaml')) {
                this.showToast();
                return;
            }
            const reader = new FileReader();
            reader.onload = () => {
                const fileAsBinaryString = reader.result;
                this.props.onYamlUpload({ content: fileAsBinaryString, name: file.name });
            };
            reader.readAsBinaryString(file);
        });
    };
    showToast = () => {
        const errorMessage = 'Invalid file type. Try again.';
        this.props.addToast(errorMessage);
        setTimeout(this.props.removeToast, 500);
    };
    hideGetStartedSection = () => this.setState({ showGetStartedSection: false });

    toggleDragAndDrop = showDragAndDrop => {
        this.setState({ showDragAndDrop });
    };

    renderGettingStarted = () => (
        <section className="bg-white shadow text-base-600 border border-base-200 m-3 flex flex-col flex-no-shrink">
            <div className="p-3 border-b border-base-300 text-primary-600 cursor-pointer flex justify-between space-between">
                <h1 className="text-base text-base-600 font-600">Getting Started</h1>
                <Icon.X className="h-4 w-4 text-base-500" onClick={this.hideGetStartedSection} />
            </div>
            <div className="pt-3 pr-3 pl-3 self-center">
                <img alt="" src={gettingStarted} />
            </div>
            <div className="m-3 border-t border-dashed border-base-300 pt-3 leading-loose font-500">
                The network simulator allows you to quickly preview your environment under different
                policy configurations. After proper configuration, notify and share the YAML file
                with your team. To get started, upload a YAML file below.
            </div>
        </section>
    );

    renderLoadingSection = () => (
        <section className="m-3 flex flex-1 border border-dashed border-base-300 bg-white">
            <div className="flex flex-col flex-1 font-500 uppercase">
                <Loader message="Processing Network Policies" />
            </div>
        </section>
    );

    renderDragAndDrop = message => (
        <section className="bg-white m-3 flex flex-1 border border-dashed border-base-300 cursor-pointer">
            <ReactDropzone
                onDrop={this.onDrop}
                className="flex flex-1 flex-col self-center uppercase p-5"
            >
                <div
                    className="h-16 w-16 self-center rounded-full flex items-center justify-center"
                    style={{ background: '#faecd2', color: '#b39357' }}
                >
                    <Icon.Upload className="h-10 w-10" strokeWidth="1.5px" />
                </div>

                <div className="text-center pt-6">{message}</div>
            </ReactDropzone>
        </section>
    );

    renderProcessingView = () => {
        const { yamlUploadState } = this.props;
        if (yamlUploadState !== 'REQUEST') return null;
        return <div className="flex flex-col flex-1">{this.renderLoadingSection()}</div>;
    };

    renderUploadView = () => {
        const { yamlUploadState } = this.props;
        if (yamlUploadState !== 'INITIAL') return null;
        const uploadMessage = 'Click to upload or drop network policy yaml inside.';
        return (
            <div className="flex flex-col flex-1">
                {this.state.showGetStartedSection && this.renderGettingStarted()}
                {this.renderDragAndDrop(uploadMessage)}
            </div>
        );
    };

    renderSuccessView = () => {
        const { yamlUploadState } = this.props;
        if (yamlUploadState !== 'SUCCESS') return null;

        const uploadMessage = 'Simulate another set of policies';
        return (
            <div className="flex flex-col flex-1">
                {this.state.showDragAndDrop && (
                    <div className="h-1/5">{this.renderDragAndDrop(uploadMessage)}</div>
                )}
                <NetworkPolicySimulatorSuccessView
                    yamlFile={this.props.yamlFile}
                    onCollapse={this.toggleDragAndDrop}
                />
                <SendNotificationSection />
            </div>
        );
    };

    renderErrorView = () => {
        const { yamlUploadState } = this.props;
        if (yamlUploadState !== 'ERROR') return null;
        const uploadMessage = 'Simulate another set of policies';
        return (
            <div className="flex flex-col flex-1">
                <div className="h-1/5">{this.renderDragAndDrop(uploadMessage)}</div>
                <NetworkPolicySimulatorErrorView
                    yamlFile={this.props.yamlFile}
                    errorMessage={this.props.errorMessage}
                />
            </div>
        );
    };
    renderSidePanel() {
        const { yamlFile } = this.props;
        const colorType = this.props.yamlUploadState === 'ERROR' ? 'danger' : 'success';
        const header = 'Network Policy Simulator';
        return (
            <Panel
                className="border-r-0"
                header={header}
                onClose={this.onClose}
                closeButtonClassName={`bg-${colorType}-500 hover:bg-${colorType}-500`}
                closeButtonIconColor="text-white"
            >
                {!yamlFile && this.renderUploadView()}
                {yamlFile && this.renderProcessingView()}
                {yamlFile && this.renderErrorView()}
                {yamlFile && this.renderSuccessView()}
            </Panel>
        );
    }

    render() {
        return (
            <div className="h-full absolute pin-r pin-b w-2/5 pt-1 pb-1 pr-1 shadow-md">
                {this.renderSidePanel()}
            </div>
        );
    }
}

const mapDispatchToProps = {
    addToast: notificationActions.addNotification,
    removeToast: notificationActions.removeOldestNotification
};

export default connect(null, mapDispatchToProps)(NetworkPolicySimulator);
