import {createConfigNamespaceApi} from "../util/namespaceApi";

export const searchConfigApi = createConfigNamespaceApi<Config.ISearch>({
    namespace: "search",
    getConfig: () => window.scribli.config.search,
    setConfig: (data) => {
        window.scribli.config.search = data;
    },
    apiPath: "/api/setting/setSearch",
});
