import {createConfigNamespaceApi} from "../util/namespaceApi";

/** AI Tab 命名空间：设置面板注册项 save */
export const aiConfigApi = createConfigNamespaceApi<Config.IAI>({
    namespace: "ai",
    getConfig: () => window.scribli.config.ai,
    setConfig: (data) => {
        window.scribli.config.ai = data;
    },
    apiPath: "/api/setting/setAI",
});
