import { graphvizRender } from "./render/graphvizRender";
import { highlightRender } from "./render/highlightRender";
import { mathRender } from "./render/mathRender";
import { mermaidRender } from "./render/mermaidRender";
import { flowchartRender } from "./render/flowchartRender";
import { chartRender } from "./render/chartRender";
import { abcRender } from "./render/abcRender";
import { htmlRender } from "./render/htmlRender";
import { mindmapRender } from "./render/mindmapRender";
import { plantumlRender } from "./render/plantumlRender";
import "../assets/scss/export.scss";

class Protyle {
    public static graphvizRender = graphvizRender;
    public static highlightRender = highlightRender;
    public static mathRender = mathRender;
    public static mermaidRender = mermaidRender;
    public static flowchartRender = flowchartRender;
    public static chartRender = chartRender;
    public static abcRender = abcRender;
    public static mindmapRender = mindmapRender;
    public static plantumlRender = plantumlRender;
    public static htmlRender = htmlRender;
}

window.Protyle = Protyle;

export default Protyle;
