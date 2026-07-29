import {createConfigNamespaceApi} from "../util/namespaceApi";

/** 密钥库 Tab 命名空间：设置面板注册项 save */
export const secretsConfigApi = createConfigNamespaceApi<Config.ISecrets>({
    namespace: "secrets",
    getConfig: () => window.scribli.config.secrets,
    setConfig: (data) => {
        window.scribli.config.secrets = data;
    },
    apiPath: "/api/setting/setSecrets",
});

/** 变量库 Tab 命名空间：设置面板注册项 save */
export const variablesConfigApi = createConfigNamespaceApi<Config.IVariables>({
    namespace: "variables",
    getConfig: () => window.scribli.config.variables,
    setConfig: (data) => {
        window.scribli.config.variables = data;
    },
    apiPath: "/api/setting/setVariables",
});
