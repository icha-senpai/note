import {fetchPost} from "../../util/fetch";
import {processSync} from "../../dialog/processSystem";
import {
    refreshSyncModeRelatedItems,
    refreshSyncTabPanels,
} from "./syncUi";

export let syncTabElement: HTMLElement | undefined;

export const clearSyncTabElement = () => {
    syncTabElement = undefined;
};

export const mountSyncTabExtras = (root: HTMLElement) => {
    syncTabElement = root;
    refreshSyncTabPanels(root);
};

export const refreshSyncCloudSpaceGroup = (root: Element) => {
    refreshSyncTabPanels(root);
    const syncConfigElement = root.querySelector("#syncCloudList");
    if (syncConfigElement) {
        syncConfigElement.innerHTML = "";
        syncConfigElement.classList.add("fn__none");
    }
};

export const patchSyncConfig = (controlId: string, value: unknown) => {
    switch (controlId) {
        case "sync.provider": {
            const provider = value as Config.ISync["provider"];
            fetchPost("/api/sync/setSyncProvider", {provider}, () => {
                window.scribli.config.sync.provider = provider;
                if (syncTabElement) {
                    refreshSyncCloudSpaceGroup(syncTabElement);
                }
            });
            break;
        }
        case "sync.enabled": {
            const enabled = Boolean(value) as Config.ISync["enabled"];
            fetchPost("/api/sync/setSyncEnable", {enabled}, () => {
                window.scribli.config.sync.enabled = enabled;
                processSync();
            });
            break;
        }
        case "sync.generateConflictDoc": {
            const generateConflictDoc = Boolean(value) as Config.ISync["generateConflictDoc"];
            fetchPost("/api/sync/setSyncGenerateConflictDoc", {enabled: generateConflictDoc}, () => {
                window.scribli.config.sync.generateConflictDoc = generateConflictDoc;
            });
            break;
        }
        case "sync.mode": {
            const mode = value as Config.ISync["mode"];
            fetchPost("/api/sync/setSyncMode", {mode}, () => {
                window.scribli.config.sync.mode = mode;
                if (syncTabElement) {
                    refreshSyncModeRelatedItems(syncTabElement);
                }
            });
            break;
        }
        case "sync.interval": {
            const interval = value as Config.ISync["interval"];
            fetchPost("/api/sync/setSyncInterval", {interval}, () => {
                window.scribli.config.sync.interval = interval;
                processSync();
            });
            break;
        }
        case "sync.perception": {
            const perception = Boolean(value) as Config.ISync["perception"];
            fetchPost("/api/sync/setSyncPerception", {enabled: perception}, () => {
                window.scribli.config.sync.perception = perception;
                processSync();
            });
            break;
        }

        case "repo.indexRetentionDays": {
            const indexRetentionDays = value as Config.IRepo["indexRetentionDays"];
            fetchPost("/api/repo/setRepoIndexRetentionDays", {days: indexRetentionDays}, () => {
                window.scribli.config.repo.indexRetentionDays = indexRetentionDays;
            });
            break;
        }
        case "repo.retentionIndexesDaily": {
            const retentionIndexesDaily = value as Config.IRepo["retentionIndexesDaily"];
            fetchPost("/api/repo/setRetentionIndexesDaily", {indexes: retentionIndexesDaily}, () => {
                window.scribli.config.repo.retentionIndexesDaily = retentionIndexesDaily;
            });
            break;
        }
        default:
            console.warn(`[config] patchSyncConfig: unhandled controlId "${controlId}"`);
            break;
    }
};
