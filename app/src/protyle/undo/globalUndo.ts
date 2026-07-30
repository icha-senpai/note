import {Constants} from "../../constants";
import {fetchPost} from "../../util/fetch";
import {confirmDialog} from "../../dialog/confirmDialog";
import {showMessage} from "../../dialog/message";
import {waitForPendingTransactions} from "../util/transactionQueue";
import {getActiveTab} from "../../layout/tabUtil";

interface IUndoStateMirror {
    canUndo: boolean;
    canRedo: boolean;
}

const undoStateMirror = new Map<string, IUndoStateMirror>();
let isUndoing = false;

export const markMirror = (rootID: string, state: Partial<IUndoStateMirror>) => {
    const cur = undoStateMirror.get(rootID) || {canUndo: false, canRedo: false};
    undoStateMirror.set(rootID, {...cur, ...state});
};

export const getMirror = (rootID: string): IUndoStateMirror => {
    return undoStateMirror.get(rootID) || {canUndo: false, canRedo: false};
};

export const syncMirrorFromBroadcast = (undoState: { [rootID: string]: { canUndo: boolean; canRedo: boolean } }) => {
    if (!undoState) {
        return;
    }
    Object.entries(undoState).forEach(([rootID, state]) => {
        undoStateMirror.set(rootID, {canUndo: !!state.canUndo, canRedo: !!state.canRedo});
    });
};

export const initMirror = (rootID: string) => {
    if (!rootID) {
        return;
    }
    fetchPost("/api/transactions/undoState", {rootID}, (response) => {
        const data = response.data;
        if (data) {
            undoStateMirror.set(rootID, {canUndo: !!data.canUndo, canRedo: !!data.canRedo});
        }
    });
};

export const refreshUndoButtons = (protyle: IProtyle) => {
    if (!protyle.block?.rootID) {
        return;
    }
    const state = getMirror(protyle.block.rootID);
    if (protyle.breadcrumb) {
        const parent = protyle.breadcrumb.element.parentElement;
        const undoElement = parent.querySelector('[data-type="undo"]') as HTMLElement;
        const redoElement = parent.querySelector('[data-type="redo"]') as HTMLElement;
        if (undoElement) {
            if (state.canUndo) {
                undoElement.removeAttribute("disabled");
            } else {
                undoElement.setAttribute("disabled", "disabled");
            }
        }
        if (redoElement) {
            if (state.canRedo) {
                redoElement.removeAttribute("disabled");
            } else {
                redoElement.setAttribute("disabled", "disabled");
            }
        }
    }
};

export const getActiveProtyle = (): IProtyle => {
    const activeTab = getActiveTab();
    const model = activeTab?.model;
    if (model && (model as any).editor?.protyle) {
        return (model as any).editor.protyle;
    }
    const allProtyle = window.scribli?.blockPanels || [];
    for (const panel of allProtyle) {
        if (!panel.element || !document.activeElement || !panel.element.contains(document.activeElement)) {
            continue;
        }
        const editor = panel.editors.find(item => item.protyle.element.contains(document.activeElement));
        if (editor?.protyle) {
            return editor.protyle;
        }
    }
    return undefined;
};

const resolveRootNames = async (rootIDs: string[]): Promise<string[]> => {
    const names: string[] = [];
    for (const id of rootIDs) {
        await new Promise<void>((resolve) => {
            fetchPost("/api/filetree/getHPathByID", {id}, (response: IWebSocketData) => {
                if (response.code === 0 && response.data) {
                    names.push(response.data as string);
                } else {
                    names.push(id);
                }
                resolve();
            });
        });
    }
    return names;
};

const focusRootIDs = (rootIDs: string[], focusBlockId?: string) => {
    const protyle = getActiveProtyle();
    if (protyle && rootIDs.includes(protyle.block?.rootID) && focusBlockId) {
        const target = protyle.wysiwyg.element.querySelector(`[data-node-id="${focusBlockId}"]`);
        if (target) {
            const rect = target.getBoundingClientRect();
            if (rect.bottom < 0 || rect.top > window.innerHeight) {
                target.scrollIntoView({behavior: "smooth", block: "center"});
            }
        }
    }
};

export const requestUndo = async (protyle: IProtyle) => {
    if (!protyle || isUndoing) {
        return;
    }
    const rootID = protyle.block?.rootID;
    if (!rootID) {
        return;
    }

    const state = getMirror(rootID);
    if (!state.canUndo) {
        return;
    }

    isUndoing = true;
    await waitForPendingTransactions(protyle);

    let peekMutatedRootIDs: string[] = [];
    await new Promise<void>((resolve) => {
        fetchPost("/api/transactions/undoState", {rootID}, (response) => {
            if (response.data?.peekMutatedRootIDs) {
                peekMutatedRootIDs = response.data.peekMutatedRootIDs;
            }
            resolve();
        });
    });

    if (peekMutatedRootIDs.length > 1) {
        const names = await resolveRootNames(peekMutatedRootIDs);
        const blockInput = (e: Event) => {
            e.stopImmediatePropagation();
            e.preventDefault();
        };
        protyle.wysiwyg.element.addEventListener("keydown", blockInput, true);
        protyle.wysiwyg.element.addEventListener("beforeinput", blockInput, true);
        const confirmed = await new Promise<boolean>((resolve) => {
            confirmDialog(`⚠️ ${window.scribli.languages.undo}`,
                `${window.scribli.languages.undoCrossDocConfirm}<div style="margin-top: 8px;">${names.map(n => `• ${n}`).join("<br>")}</div>`,
                () => resolve(true),
                () => resolve(false));
        });
        protyle.wysiwyg.element.removeEventListener("keydown", blockInput, true);
        protyle.wysiwyg.element.removeEventListener("beforeinput", blockInput, true);
        if (!confirmed) {
            isUndoing = false;
            return;
        }
    }

    fetchPost("/api/transactions/undo", {
        rootID,
        app: Constants.SCRIBLI_APPID,
        session: protyle.id,
    }, (response) => {
        isUndoing = false;
        const data = response.data;
        if (!data) {
            return;
        }
        if (data.failed) {
            if (data.msg) {
                showMessage(data.msg);
            }
            return;
        }
        if (!data.undoOperations || data.undoOperations.length === 0) {
            markMirror(rootID, {canUndo: !!data.canUndo, canRedo: !!data.canRedo});
            refreshUndoButtons(protyle);
            return;
        }
        markMirror(rootID, {canUndo: !!data.canUndo, canRedo: !!data.canRedo});
        const mutatedRootIDs: string[] = data.mutatedRootIDs || [];
        if (mutatedRootIDs.length > 1) {
            refreshUndoButtons(protyle);
        } else {
            protyle.undo.renderLocal(protyle, data.doOperations);
            refreshUndoButtons(protyle);
            const focusBlockId = data.doOperations?.find((op: IOperation) => op.id)?.id;
            focusRootIDs(mutatedRootIDs, focusBlockId);
        }
    });
};

export const requestRedo = async (protyle: IProtyle) => {
    if (!protyle || isUndoing) {
        return;
    }
    const rootID = protyle.block?.rootID;
    if (!rootID) {
        return;
    }

    const state = getMirror(rootID);
    if (!state.canRedo) {
        return;
    }

    isUndoing = true;
    await waitForPendingTransactions(protyle);
    fetchPost("/api/transactions/redo", {
        rootID,
        app: Constants.SCRIBLI_APPID,
        session: protyle.id,
    }, (response) => {
        isUndoing = false;
        const data = response.data;
        if (!data) {
            return;
        }
        if (data.failed) {
            if (data.msg) {
                showMessage(data.msg);
            }
            return;
        }
        if (!data.doOperations || data.doOperations.length === 0) {
            markMirror(rootID, {canUndo: !!data.canUndo, canRedo: !!data.canRedo});
            refreshUndoButtons(protyle);
            return;
        }
        markMirror(rootID, {canUndo: !!data.canUndo, canRedo: !!data.canRedo});
        const mutatedRootIDs: string[] = data.mutatedRootIDs || [];
        if (mutatedRootIDs.length > 1) {
            refreshUndoButtons(protyle);
        } else {
            protyle.undo.renderLocal(protyle, data.doOperations);
            refreshUndoButtons(protyle);
            const focusBlockId = data.doOperations?.find((op: IOperation) => op.id)?.id;
            focusRootIDs(mutatedRootIDs, focusBlockId);
        }
    });
};
