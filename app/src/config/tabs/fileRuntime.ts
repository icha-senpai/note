import {createConfigNamespaceApi} from "../util/namespaceApi";

export const fileConfigApi = createConfigNamespaceApi<Config.IFileTree>({
    namespace: "fileTree",
    getConfig: () => window.scribli.config.fileTree,
    setConfig: (data) => {
        window.scribli.config.fileTree = data;
    },
    apiPath: "/api/setting/setFiletree",
});
