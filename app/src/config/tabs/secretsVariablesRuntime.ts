import {createConfigNamespaceApi} from "../util/namespaceApi";

export const secretsConfigApi = createConfigNamespaceApi<Config.ISecrets>({
    namespace: "secrets",
    getConfig: () => window.scribli.config.secrets,
    setConfig: (data) => {
        window.scribli.config.secrets = data;
    },
    apiPath: "/api/setting/setSecrets",
});

export const variablesConfigApi = createConfigNamespaceApi<Config.IVariables>({
    namespace: "variables",
    getConfig: () => window.scribli.config.variables,
    setConfig: (data) => {
        window.scribli.config.variables = data;
    },
    apiPath: "/api/setting/setVariables",
});
