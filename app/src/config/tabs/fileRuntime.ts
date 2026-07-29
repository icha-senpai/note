import {createConfigNamespaceApi} from "../util/namespaceApi";

/** 文档 Tab 命名空间：设置面板注册项 save */
export const fileConfigApi = createConfigNamespaceApi<Config.IFileTree>({
    namespace: "fileTree",
    getConfig: () => window.scribli.config.fileTree,
    setConfig: (data) => {
        window.scribli.config.fileTree = data;
    },
    apiPath: "/api/setting/setFiletree",
});
