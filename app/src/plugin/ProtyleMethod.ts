import { graphvizRender } from "../protyle/render/graphvizRender";
import { highlightRender } from "../protyle/render/highlightRender";
import { mathRender } from "../protyle/render/mathRender";
import { mermaidRender } from "../protyle/render/mermaidRender";
import { flowchartRender } from "../protyle/render/flowchartRender";
import { chartRender } from "../protyle/render/chartRender";
import { abcRender } from "../protyle/render/abcRender";
import { htmlRender } from "../protyle/render/htmlRender";
import { mindmapRender } from "../protyle/render/mindmapRender";
import { plantumlRender } from "../protyle/render/plantumlRender";
import { avRender } from "../protyle/render/av/render";

export class ProtyleMethod {
    public static graphvizRender = graphvizRender;
    public static highlightRender = highlightRender;
    public static mathRender = mathRender;
    public static mermaidRender = mermaidRender;
    public static flowchartRender = flowchartRender;
    public static chartRender = chartRender;
    public static abcRender = abcRender;
    public static mindmapRender = mindmapRender;
    public static plantumlRender = plantumlRender;
    public static avRender = avRender;
    public static htmlRender = htmlRender;
}
