import type {Tab} from "../Tab";
import {App} from "../../index";
import {Model} from "../Model";

export class Inbox extends Model {
    constructor(app: App, tab: Tab | Element) {
        super({app});

        const element = tab instanceof Element ? tab : tab.panelElement;
        element.classList.add("fn__flex-column", "file-tree", "sy__inbox", "dockPanel");
        element.innerHTML = `<div class="block__icons">
    <div class="block__logo fn__flex-1">${window.scribli.languages.inbox}</div>
</div>
<div class="fn__flex-1">
    <ul class="b3-list b3-list--background">
        <li class="b3-list--empty">${window.scribli.languages.inboxTip}</li>
    </ul>
</div>`;
    }
}
