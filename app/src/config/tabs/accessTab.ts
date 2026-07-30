import type {SettingTabBuilder} from "../setting/builder";
import {fetchPost, fetchSyncPost} from "../../util/fetch";
import {Dialog} from "../../dialog";
import {confirmDialog} from "../../dialog/confirmDialog";
import {Constants} from "../../constants";
import {isBrowser, isMobile} from "../../util/functions";
import {showMessage} from "../../dialog/message";
/// #if !BROWSER
import {shell} from "electron";
/// #endif
import {saveExportFile} from "../../protyle/util/compatibility";
import {openByMobile} from "../../editor/openLink";
import {genConfigItemMainHtml} from "../render/fragments";
import {renderPublishAuthAccounts, savePublish, sendAccessSetting, updatePublishConfig} from "./accessRuntime";
import {sendAppSetting} from "./appRuntime";
import zxcvbn = require("zxcvbn");

const getPasswordStrength = (password: string) => {
    const score = zxcvbn(password).score;
    if (score <= 1) {
        return "weak";
    }
    if (score === 2) {
        return "medium";
    }
    return "strong";
};

const updatePasswordStrength = (element: HTMLElement, password: string) => {
    if (!password) {
        element.classList.add("fn__none");
        return;
    }
    const strength = getPasswordStrength(password);
    element.classList.remove("fn__none");
    element.setAttribute("data-strength", strength);
    element.textContent = window.scribli.languages[`passwordStrength${strength[0].toUpperCase()}${strength.slice(1)}`];
};

const confirmWeakPassword = (password: string, confirm: () => void) => {
    if (getPasswordStrength(password) !== "weak") {
        confirm();
        return;
    }
    confirmDialog("⚠️ " + window.scribli.languages.weakPasswordConfirmTitle, window.scribli.languages.weakPasswordConfirmTip, confirm);
};

const registerAccessAuthGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("authentication", window.scribli.languages.authentication);
    const onWeb = isBrowser();

    if (!window.scribli.config.readonly && !onWeb) {
        group.button({
            id: "authCode",
            title: window.scribli.languages.about5,
            desc: window.scribli.languages.about6,
            label: window.scribli.languages.config,
            icon: "iconLock",
            afterMount: mountAuthCodeButton,
        });
    }
    if (window.scribli.config.accessAuthCode && !onWeb) {
        group.switch("system.lockScreenMode", {
            title: window.scribli.languages.about7,
            desc: window.scribli.languages.about8,
            save: (value) => sendAppSetting("system.lockScreenMode", value),
        });
    }
    group.text("api.token", {
        title: window.scribli.languages.about13,
        desc: window.scribli.languages.about14.replace("${token}", window.scribli.config.api.token),
        save: (value) => sendAccessSetting("api.token", value),
        afterMount: bindApiTokenInput,
    });
};

const mountAuthCodeButton = (root: HTMLElement) => {
    root.querySelector("#authCode")?.addEventListener("click", () => {
        const dialog = new Dialog({
            title: window.scribli.languages.about5,
            content: `<div class="b3-dialog__content">
    <input class="b3-text-field fn__block" placeholder="${window.scribli.languages.about5}" value="${window.scribli.config.accessAuthCode}">
    <div class="b3-label__text">${window.scribli.languages.about6}</div>
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
            width: isMobile() ? "92vw" : "520px",
        });
        const inputElement = dialog.element.querySelector("input") as HTMLInputElement;
        const btnsElement = dialog.element.querySelectorAll(".b3-button");
        dialog.element.setAttribute("data-key", Constants.DIALOG_ACCESSAUTHCODE);
        dialog.bindInput(inputElement, () => {
            (btnsElement[1] as HTMLButtonElement).click();
        });
        inputElement.select();
        btnsElement[0].addEventListener("click", () => {
            dialog.destroy();
        });
        btnsElement[1].addEventListener("click", () => {
            fetchPost("/api/system/setAccessAuthCode", {accessAuthCode: inputElement.value});
        });
    });
};

const bindApiTokenInput = (root: HTMLElement) => {
    const tokenElement = root.querySelector<HTMLInputElement>(`#${CSS.escape("api.token")}`);
    let tokenFocused = false;
    tokenElement?.addEventListener("focus", () => {
        tokenFocused = true;
    });
    tokenElement?.addEventListener("blur", () => {
        tokenFocused = false;
    });
    tokenElement?.addEventListener("mousedown", (event) => {
        if (!tokenFocused) {
            event.preventDefault();
            tokenElement.select();
        }
    });
};

const registerAccessServerGroup = (tab: SettingTabBuilder) => {
    const hideOnWeb = isBrowser();
    if (hideOnWeb) {
        return;
    }
    const group = tab.group("server", window.scribli.languages.configGroupServer);

    group.switch("system.networkServe", {
        title: window.scribli.languages.about11,
        desc: window.scribli.languages.about12,
        save: (value) => sendAppSetting("system.networkServe", value),
    });
    if (window.scribli.config.system.networkServe) {
        group.switch("system.networkServeTLS", {
            title: window.scribli.languages.networkServeTLS,
            desc: `${window.scribli.languages.networkServeTLSTip}<div class="fn__hr--small"></div>${window.scribli.languages.networkServeTLSTip2}`,
            save: (value) => sendAppSetting("system.networkServeTLS", value),
        });
    }
    if (window.scribli.config.system.networkServe && window.scribli.config.system.networkServeTLS) {
        group.button({
            id: "exportCACert",
            title: window.scribli.languages.exportCACert,
            desc: window.scribli.languages.exportCACertTip,
            label: window.scribli.languages.export,
            icon: "iconUpload",
            afterMount: (root) => {
                root.querySelector("#exportCACert")?.addEventListener("click", () => {
                    fetchPost("/api/system/exportTLSCACert", {}, (response) => {
                        void saveExportFile(response.data.path);
                    });
                });
            },
        });
        group.button({
            id: "exportCABundle",
            title: window.scribli.languages.exportCABundle,
            desc: window.scribli.languages.exportCABundleTip,
            label: window.scribli.languages.export,
            icon: "iconUpload",
            afterMount: (root) => {
                root.querySelector("#exportCABundle")?.addEventListener("click", () => {
                    fetchPost("/api/system/exportTLSCABundle", {}, (response) => {
                        void saveExportFile(response.data.path);
                    });
                });
            },
        });
        group.button({
            id: "importCABundle",
            title: window.scribli.languages.importCABundle,
            desc: window.scribli.languages.importCABundleTip,
            label: window.scribli.languages.import,
            icon: "iconDownload",
            afterMount: (root) => {
                root.querySelector("#importCABundle")?.addEventListener("click", () => {
                    const input = document.createElement("input");
                    input.type = "file";
                    input.accept = ".zip";
                    input.onchange = () => {
                        if (input.files && input.files[0]) {
                            const formData = new FormData();
                            formData.append("file", input.files[0]);
                            fetch("/api/system/importTLSCABundle", {
                                method: "POST",
                                body: formData,
                            }).then(res => res.json()).then((response) => {
                                if (response.code === 0) {
                                    showMessage(window.scribli.languages.importCABundleSuccess);
                                } else {
                                    showMessage(response.msg, 6000, "error");
                                }
                            });
                        }
                    };
                    input.click();
                });
            },
        });
    }
    group.stack({
        key: "localServer",
        keywords: [
            window.scribli.languages.about2,
            window.scribli.languages.about3,
            window.scribli.languages.about4,
            window.scribli.languages.about18,
        ],
        afterMount: (root) => {
            root.querySelector("#openLocalServer")?.addEventListener("click", () => {
                const url = `http://127.0.0.1:${location.port}`;
                /// #if !BROWSER
                void shell.openExternal(url);
                /// #else
                openByMobile(url);
                /// #endif
            });
        },
    }, (stack) => {
        stack.title(window.scribli.languages.about2);
        stack.button({
            id: "openLocalServer",
            label: window.scribli.languages.about4,
            icon: "iconLink",
        });
        stack.desc(window.scribli.languages.about3.replace("${port}", location.port));
        stack.desc((() => {
            const parts: string[] = [];
            for (const serverAddr of window.scribli.config.serverAddrs) {
                if (!serverAddr.trim()) {
                    break;
                }
                parts.push(`<code class="fn__code">${serverAddr}</code>`);
            }
            return parts.join(" ");
        })());
        stack.desc(window.scribli.languages.about18);
    });
};

const registerAccessPublishGroup = (tab: SettingTabBuilder) => {
    const group = tab.group("publish", window.scribli.languages.configGroupPublish);

    group.switch("publish.enable", {
        title: window.scribli.languages.publishService,
        desc: window.scribli.languages.publishServiceTip,
        save: (value) => sendAccessSetting("publish.enable", value),
    });
    group.number("publish.port", {
        title: window.scribli.languages.publishServicePort,
        desc: window.scribli.languages.publishServicePortTip,
        min: 0,
        max: 65535,
        save: (value) => sendAccessSetting("publish.port", value),
    });
    group.slot({
        key: "publishAddresses",
        keywords: [
            window.scribli.languages.publishServiceAddresses,
            window.scribli.languages.publishServiceAddressesTip,
            window.scribli.languages.publishServiceNotStarted,
        ],
        html: () => `<div class="b3-label config-item">
    <div class="fn__flex config-wrap">
        ${genConfigItemMainHtml(window.scribli.languages.publishServiceAddresses, window.scribli.languages.publishServiceAddressesTip)}
        <div class="fn__space"></div>
    </div>
    <div id="publishAddresses" class="b3-label__text"></div>
</div>`,
        afterMount: () => {
            fetchPost("/api/setting/getPublish", {}, (response: IWebSocketData) => {
                updatePublishConfig(true, response);
            });
        },
    });
    group.switch("publish.auth.enable", {
        title: window.scribli.languages.publishServiceAuth,
        desc: window.scribli.languages.publishServiceAuthTip,
        save: (value) => sendAccessSetting("publish.auth.enable", value),
    });
    group.button({
        id: "publishAuthAccountAdd",
        title: window.scribli.languages.publishServiceAuthAccounts,
        desc: window.scribli.languages.publishServiceAuthAccountsTip,
        label: window.scribli.languages.publishServiceAuthAccountAdd,
        icon: "iconAdd",
        afterMount: (root) => {
            root.querySelector("#publishAuthAccountAdd")?.addEventListener("click", () => {
                window.scribli.config.publish.auth.accounts.push({
                    username: "",
                    password: "",
                    memo: "",
                });
                renderPublishAuthAccounts();
            });
        },
    });
    group.slot({
        key: "publishAuthAccounts",
        keywords: [
            window.scribli.languages.userName,
            window.scribli.languages.password,
            window.scribli.languages.memo,
            window.scribli.languages.delete,
        ],
        html: () => '<div class="b3-label config-item"><div class="fn__flex-1" id="publishAuthAccounts" style="overflow: visible;"></div></div>',
        afterMount: mountPublishAuthAccounts,
    });
};

const mountPublishAuthAccounts = (root: HTMLElement) => {
    const publishAuthAccounts = root.querySelector("#publishAuthAccounts");
    publishAuthAccounts?.addEventListener("change", (event) => {
        const input = event.target as HTMLInputElement;
        if (input.tagName !== "INPUT" || !input.dataset.name) {
            return;
        }
        const li = input.closest("li");
        if (li) {
            const index = parseInt(li.dataset.index);
            const name = input.dataset.name as keyof Config.IPublishAuthAccount;
            if (name in window.scribli.config.publish.auth.accounts[index]) {
                window.scribli.config.publish.auth.accounts[index][name] = input.value;
                savePublish(false);
            }
        }
    });
    publishAuthAccounts?.addEventListener("click", (event) => {
        const target = event.target as Element;
        const li = target.closest('[data-action="remove"]')?.closest("li");
        if (li) {
            const index = parseInt(li.dataset.index);
            window.scribli.config.publish.auth.accounts.splice(index, 1);
            savePublish(true);
            return;
        }
        const togglePassword = target.closest('.b3-form__icona-icon[data-action="togglePassword"]');
        if (togglePassword) {
            const isEye = togglePassword.firstElementChild.getAttribute("xlink:href") === "#iconEye";
            togglePassword.firstElementChild.setAttribute("xlink:href", isEye ? "#iconEyeoff" : "#iconEye");
            togglePassword.previousElementSibling.setAttribute("type", isEye ? "text" : "password");
        }
    });
};

export const registerAccessTab = (tab: SettingTabBuilder) => {
    registerAccessAuthGroup(tab);
    registerEncryptedNotebookGroup(tab);
    registerAccessServerGroup(tab);
    registerAccessPublishGroup(tab);
};

const registerEncryptedNotebookGroup = (tab: SettingTabBuilder) => {
    if (window.scribli.config.readonly) {
        return;
    }
    const group = tab.group("encryptedNotebook", window.scribli.languages.encryptedNotebook);
    group.slot({
        key: "encryptedNotebookStatus",
        keywords: [
            window.scribli.languages.encryptedNotebook,
            window.scribli.languages.enableEncryptedNotebook,
            window.scribli.languages.masterPassword,
            window.scribli.languages.changeMasterPassword,
        ],
        html: () =>
            `<label class="fn__flex b3-label config-item">
	    ${genConfigItemMainHtml(window.scribli.languages.enableEncryptedNotebook, window.scribli.languages.encryptedNotebookTip + "<br><span class=\"ft__error\">" + window.scribli.languages.encryptedNotebookRiskTip + "</span><br>" + window.scribli.languages.featurePreview)}
    <span class="fn__space"></span>
    <input class="b3-switch fn__flex-center" id="encryptedNotebookSwitch" type="checkbox">
</label>
<div class="b3-label config-item fn__none" id="encryptedNotebookMigrationAlert">
    <div class="ft__error">${window.scribli.languages.masterPasswordMigrationPending}</div>
</div>
<div class="b3-label config-item fn__none" id="encryptedNotebookActions">
    <div class="fn__flex fn__flex-center config-wrap">
        <div class="fn__flex-1"></div>
        <div class="fn__flex fn__flex-center" id="encryptedNotebookEnabledActions">
            <button class="b3-button b3-button--outline fn__flex-center fn__size200" id="changeMasterPasswordBtn">
                <svg class="svg"><use xlink:href="#iconLock"></use></svg>
                ${window.scribli.languages.changeMasterPassword}
            </button>
            <span class="fn__space"></span>
            <button class="b3-button b3-button--outline fn__flex-center fn__size200" id="exportCryptoBackupBtn">
                <svg class="svg"><use xlink:href="#iconDownload"></use></svg>
                ${window.scribli.languages.exportNotebookCryptoBackup}
            </button>
            <span class="fn__space"></span>
        </div>
        <button class="b3-button b3-button--outline fn__flex-center fn__size200" id="importCryptoBackupBtn">
            <svg class="svg"><use xlink:href="#iconUpload"></use></svg>
            ${window.scribli.languages.importNotebookCryptoBackup}
        </button>
    </div>
</div>`,
        afterMount: mountEncryptedNotebook,
    });
    group.number("notebookCrypto.autoLockMinutes", {
        title: window.scribli.languages.encryptedNotebookAutoLock,
        desc: window.scribli.languages.encryptedNotebookAutoLockDesc,
        min: 0,
        save: (value) => {
            fetchPost("/api/notebook/setNotebookCryptoAutoLock", {autoLockMinutes: value});
        },
    });
};

const mountEncryptedNotebook = (root: HTMLElement) => {
    const switchElement = root.querySelector("#encryptedNotebookSwitch") as HTMLInputElement;
    const actionsElement = root.querySelector("#encryptedNotebookActions");
    const enabledActionsElement = root.querySelector("#encryptedNotebookEnabledActions");
    const importCryptoBackupBtnElement = root.querySelector("#importCryptoBackupBtn");
    const migrationAlertElement = root.querySelector("#encryptedNotebookMigrationAlert");
    const refresh = () => {
        fetchPost("/api/notebook/getEncryptedNotebookStatus", {}, (response) => {
            const enabled = response.data.enabled;
            switchElement.checked = enabled;
            window.scribli.config.notebookCrypto.enabled = enabled;
            enabledActionsElement.classList.toggle("fn__none", !enabled);
            importCryptoBackupBtnElement.classList.toggle("fn__none", enabled);
            actionsElement.classList.remove("fn__none");
            migrationAlertElement.classList.toggle("fn__none", !response.data.migrationPending);
        });
    };
    refresh();

    actionsElement.querySelector("#changeMasterPasswordBtn")?.addEventListener("click", () => {
        openChangeMasterPasswordDialog(refresh);
    });

    actionsElement.querySelector("#exportCryptoBackupBtn")?.addEventListener("click", () => {
        fetchPost("/api/notebook/exportNotebookCryptoBackup", {}, async (response) => {
            if (response.code === -1) {
                showMessage(response.msg, 6000, "error");
                return;
            }
            await saveExportFile(response.data.file);
            showMessage(window.scribli.languages.exportNotebookCryptoBackupTip);
        });
    });

    actionsElement.querySelector("#importCryptoBackupBtn")?.addEventListener("click", () => {
        const fileInput = document.createElement("input");
        fileInput.type = "file";
        fileInput.accept = ".json,application/json";
        fileInput.onchange = () => {
            const file = fileInput.files?.[0];
            if (!file) {
                return;
            }
            const passwordDialog = new Dialog({
                title: window.scribli.languages.masterPassword,
                content: `<div class="b3-dialog__content">
    <input type="password" placeholder="${window.scribli.languages.masterPassword}" class="b3-text-field fn__block">
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button>
    <div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
                width: "520px",
            });
            const pwdInput = passwordDialog.element.querySelector(".b3-text-field") as HTMLInputElement;
            passwordDialog.element.querySelector(".b3-button--cancel")?.addEventListener("click", () => {
                passwordDialog.destroy();
            });
            passwordDialog.element.querySelector(".b3-button--text")?.addEventListener("click", () => {
                const password = pwdInput.value.trim();
                if (!password) {
                    showMessage(window.scribli.languages.masterPassword);
                    return;
                }
                const formData = new FormData();
                formData.append("file", file);
                formData.append("password", password);
                fetch("/api/notebook/importNotebookCryptoBackup", {
                    method: "POST",
                    body: formData,
                }).then((res) => res.json()).then((response: IWebSocketData) => {
                    if (response.code === -1) {
                        showMessage(response.msg, 6000, "error");
                        return;
                    }
                    showMessage(window.scribli.languages.importNotebookCryptoBackupTip);
                    passwordDialog.destroy();
                    refresh();
                });
            });
            pwdInput.focus();
        };
        fileInput.click();
    });

    switchElement.addEventListener("change", () => {
        if (switchElement.checked) {
            openEnableEncryptedDialog(refresh, refresh);
        } else {
            fetchPost("/api/notebook/getEncryptedNotebookStatus", {}, (response) => {
                if (response.data.count > 0) {
                    showMessage(window.scribli.languages.encryptedNotebookDisableTip.replace("${x}", response.data.count), 4000);
                    switchElement.checked = true;
                } else if (response.data.hasHistoryDependency) {
                    showMessage(window.scribli.languages["_kernel"]["323"], 6000, "error");
                    switchElement.checked = true;
                } else {
                    fetchSyncPost("/api/notebook/disableEncryptedNotebooks", {}).then((res: IWebSocketData) => {
                        if (res.code === -1) {
                            switchElement.checked = true;
                            return;
                        }
                        showMessage(window.scribli.languages.encryptedNotebookDisabled);
                        refresh();
                    });
                }
            });
        }
    });
};

const openEnableEncryptedDialog = (onSuccess: () => void, onCancel: () => void) => {
    const dialog = new Dialog({
        title: "🔐 " + window.scribli.languages.setMasterPassword,
        content: `<div class="b3-dialog__content">
    <input type="password" placeholder="${window.scribli.languages.masterPassword}" class="b3-text-field fn__block">
    <div class="password-strength fn__none"></div>
    <div class="fn__hr"></div>
    <input type="password" placeholder="${window.scribli.languages.confirmMasterPassword}" class="b3-text-field fn__block">
    <div class="fn__hr--b"></div>
    <label class="b3-label__text"><input type="checkbox" id="encRiskConfirm"> ${window.scribli.languages.encryptedNotebookRiskTip}</label>
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text" disabled>${window.scribli.languages.confirm}</button>
</div>`,
        width: isMobile() ? "92vw" : "520px",
        destroyCallback: onCancel,
    });
    const btnsElement = dialog.element.querySelectorAll(".b3-button");
    const inputs = dialog.element.querySelectorAll("input");
    const confirmBtn = btnsElement[1] as HTMLButtonElement;
    const riskCheckbox = dialog.element.querySelector("#encRiskConfirm") as HTMLInputElement;
    const passwordStrength = dialog.element.querySelector(".password-strength") as HTMLElement;
    (inputs[0] as HTMLInputElement).focus();
    inputs[0].addEventListener("input", () => updatePasswordStrength(passwordStrength, inputs[0].value));
    riskCheckbox.addEventListener("change", () => {
        confirmBtn.disabled = !riskCheckbox.checked;
    });
    btnsElement[0].addEventListener("click", () => {
        dialog.destroy();
    });
    confirmBtn.addEventListener("click", () => {
        const pwd1 = (inputs[0] as HTMLInputElement).value;
        const pwd2 = (inputs[1] as HTMLInputElement).value;
        if (!pwd1) {
            showMessage(window.scribli.languages.masterPassword);
            return;
        }
        if (pwd1 !== pwd2) {
            showMessage(window.scribli.languages.passwordNoMatch);
            return;
        }
        confirmWeakPassword(pwd1, async () => {
            const response = await fetchSyncPost("/api/notebook/enableEncryptedNotebooks", {password: pwd1});
            if (response.code === 0) {
                showMessage(window.scribli.languages.encryptedNotebookEnabled);
                dialog.destroy();
                onSuccess();
            }
        });
    });
};

const openChangeMasterPasswordDialog = (onChanged?: () => void) => {
    const dialog = new Dialog({
        title: "🔐 " + window.scribli.languages.changeMasterPassword,
        content: `<div class="b3-dialog__content">
    <input type="password" placeholder="${window.scribli.languages.oldMasterPassword}" class="b3-text-field fn__block">
    <div class="fn__hr"></div>
    <input type="password" placeholder="${window.scribli.languages.newMasterPassword}" class="b3-text-field fn__block">
    <div class="password-strength fn__none"></div>
    <div class="fn__hr"></div>
    <input type="password" placeholder="${window.scribli.languages.confirmMasterPassword}" class="b3-text-field fn__block">
</div>
<div class="b3-dialog__action">
    <button class="b3-button b3-button--cancel">${window.scribli.languages.cancel}</button><div class="fn__space"></div>
    <button class="b3-button b3-button--text">${window.scribli.languages.confirm}</button>
</div>`,
        width: isMobile() ? "92vw" : "520px",
    });
    const btnsElement = dialog.element.querySelectorAll(".b3-button");
    const inputs = dialog.element.querySelectorAll("input");
    const passwordStrength = dialog.element.querySelector(".password-strength") as HTMLElement;
    inputs[1].addEventListener("input", () => updatePasswordStrength(passwordStrength, inputs[1].value));
    btnsElement[0].addEventListener("click", () => {
        dialog.destroy();
    });
    btnsElement[1].addEventListener("click", () => {
        const oldPwd = (inputs[0] as HTMLInputElement).value;
        const newPwd = (inputs[1] as HTMLInputElement).value;
        const confirmPwd = (inputs[2] as HTMLInputElement).value;
        if (!oldPwd || !newPwd) {
            return;
        }
        if (newPwd !== confirmPwd) {
            showMessage(window.scribli.languages.passwordNoMatch);
            return;
        }
        confirmWeakPassword(newPwd, async () => {
            const response = await fetchSyncPost("/api/notebook/changeMasterPassword", {
                oldPassword: oldPwd,
                newPassword: newPwd
            });
            if (response.code === 0) {
                showMessage(window.scribli.languages.changeMasterPasswordSuccessTip);
                dialog.destroy();
            } else {
                showMessage(response.msg, 6000, "error");
                onChanged?.();
            }
        });
    });
};
