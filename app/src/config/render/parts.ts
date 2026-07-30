import type {SettingControl} from "../setting/control";

export type RowPart =
    | {
        kind: "title";
        text: string;
    }
    | {
        kind: "desc";
        text: string;
    }
    | SettingControl;

export const isSettingControl = (part: RowPart): part is SettingControl =>
    "readConfig" in part && "readValue" in part;

type SwitchQuerySwitchItem = Extract<SettingControl, {kind: "switch"}> & {
    label: string;
    icon?: string;
};

type SwitchQueryNumberItem = Extract<SettingControl, {kind: "number"}> & {
    label: string;
};

export type SwitchQueryItem = SwitchQuerySwitchItem | SwitchQueryNumberItem;

export type StackLeft =
    | {kind: "title"; text: string}
    | {kind: "desc"; text: string}
    | Extract<SettingControl, {kind: "textBlock"}>;

export type StackRight =
    | {kind: "button"; id: string; label: string; icon: string}
    | Extract<SettingControl, {kind: "switch" | "number" | "select"}>;

export type StackLine = {
    left: StackLeft;
    right?: StackRight;
};
