import {createConfigNamespaceApi} from "../util/namespaceApi";

export const aiConfigApi = createConfigNamespaceApi<Config.IAI>({
    namespace: "ai",
    getConfig: () => window.scribli.config.ai,
    setConfig: (data) => {
        window.scribli.config.ai = data;
    },
    apiPath: "/api/setting/setAI",
});
