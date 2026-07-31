# Generated Runtime Assets

Scribli ships a small set of prebuilt browser assets under `app/stage/protyle/js` because the editor loads them lazily at runtime. Treat these files as generated/vendor outputs: do not hand-edit minified bundles. Regenerate them from pinned upstream packages or from local source, then copy the exact output into `app/stage/protyle/js`.

## Scope

Only regenerate assets that Scribli actually loads. The current runtime consumers are the editor renderers, PDF viewer, graph dock, image preview/export tools, HTML export, and the service worker.

| Asset path | Runtime consumer | Source package or build source | Current pin | Regeneration path |
| --- | --- | --- | --- | --- |
| `app/stage/protyle/js/lute/lute.min.js` | `app/src/index.ts`, `app/src/window/index.ts`, HTML export, emoji preview, service worker | Local Go source in `third_party/forks/lute` | Scribli source tree | Build with GopherJS using `GOOS=js`, `GOARCH=ecmascript`, JavaScript tags, then copy the generated `lute.min.js`. Recreate this as a repo script before updating the checked-in file. |
| `app/stage/protyle/js/protyle-html.js` | HTML export, service worker | Scribli frontend source | `Constants.SCRIBLI_VERSION` | Produced by the frontend build pipeline. Regenerate with the app build, not by editing `app/stage` manually. |
| `app/stage/protyle/js/abcjs/abcjs-basic-min.js` | `app/src/protyle/render/abcRender.ts` | `abcjs` package, `dist/abcjs-basic-min.js` | `6.5.0` | Copy from the pinned package tarball. `abcjs-basic-min.min.js` is not referenced by the app and should not be regenerated unless a consumer is added. |
| `app/stage/protyle/js/echarts/echarts.min.js` | `chartRender.ts`, `mindmapRender.ts` | `echarts` package, `dist/echarts.min.js` | `5.3.2` | Copy from the pinned package tarball. Keep the query string in renderer code aligned with the file. |
| `app/stage/protyle/js/echarts/echarts-gl.min.js` | `chartRender.ts` | `echarts-gl` package, `dist/echarts-gl.min.js` | `2.0.9` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/flowchart.js/flowchart.min.js` | `flowchartRender.ts` | `raphael` plus `flowchart.js` | `raphael@2.3.0`, `flowchart.js@1.18.0` | Rebuild as a browser bundle from the pinned packages with Raphael available to Flowchart, then minify. The npm `flowchart.js` tarball ships source, not a ready `release/flowchart.min.js`. |
| `app/stage/protyle/js/graphviz/viz.js` | `graphvizRender.ts` | `@viz-js/viz`, `lib/viz-standalone.js` | `3.11.0` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/highlight.js/highlight.min.js` | `highlightRender.ts` | `highlight.js` browser build | Bundle reports `11.11.2` / git `f273f007f8` | Rebuild from the matching Highlight.js source commit. The npm registry currently has `11.11.1`, not `11.11.2`, so do not replace this with a blind package copy. |
| `app/stage/protyle/js/highlight.js/third-languages.js` | `highlightRender.ts` | Third-party Highlight.js language packages | `highlightjs-solidity@2.0.5`, `highlightjs-solidity@2.0.6` for Yul | Rebuild by concatenating the needed third-party language bundles only. Do not include languages Scribli does not register. |
| `app/stage/protyle/js/highlight.js/styles/*.min.css` | `setCodeTheme()` in `app/src/protyle/render/util.ts` | `highlight.js` styles | Kept with the checked-in Highlight.js browser bundle | Copy the needed styles from the same Highlight.js source/package used for `highlight.min.js`. If Scribli later limits the code-theme picker, prune unused styles then. |
| `app/stage/protyle/js/html-to-image.min.js` | Diagram/image export and preview | `html-to-image`, `dist/html-to-image.js` | `1.11.13` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/katex/katex.min.js` | `mathRender.ts` | `katex`, `dist/katex.min.js` | `0.16.9` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/katex/katex.min.css` and `app/stage/protyle/js/katex/fonts/**` | `mathRender.ts` and KaTeX CSS | `katex`, `dist/katex.min.css`, `dist/fonts/**` | `0.16.9` | Copy from the same pinned package as `katex.min.js`. |
| `app/stage/protyle/js/katex/mhchem.min.js` | `mathRender.ts` | `katex`, `dist/contrib/mhchem.min.js` | `0.16.9` | Copy from the pinned KaTeX package. |
| `app/stage/protyle/js/mathjax/tex-svg-full.js` | `app/src/protyle/preview/index.ts` | MathJax browser component | `3.1.2` in current bundle | Copy from the pinned MathJax package/component. This is only needed for the standalone math preview path. |
| `app/stage/protyle/js/mermaid/mermaid.min.js` | `mermaidRender.ts` | `mermaid`, `dist/mermaid.min.js` | `11.13.0` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/mermaid/mermaid-zenuml.min.js` | `mermaidRender.ts` | Mermaid ZenUML external diagram package | `0.2.2` | Build/copy from the pinned package, then apply the existing compatibility patch: remove strict-mode wrapping and expose `window.zenuml`. Replace the Chinese note in the generated output during regeneration. |
| `app/stage/protyle/js/mermaid/icons.json` | `mermaidRender.ts` | Iconify icon data used by Mermaid | `11.11.0` query string in current renderer | Generate or copy only the icon packs registered by Scribli's Mermaid configuration. Do not keep the full icon universe if the registered packs are narrower. |
| `app/stage/protyle/js/pdf/pdf.min.mjs` | PDF viewer templates | `pdfjs-dist`, `legacy/build/pdf.min.mjs` | `4.7.76` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/pdf/pdf.worker.min.mjs` | `app/src/asset/pdf/viewer.js` | `pdfjs-dist`, `legacy/build/pdf.worker.min.mjs` | `4.7.76` | Copy from the pinned package tarball. |
| `app/stage/protyle/js/pdf/pdf.sandbox.min.mjs` | PDF scripting support | `pdfjs-dist`, `legacy/build/pdf.sandbox.min.mjs` | `4.7.76` | Copy from the pinned package tarball if PDF scripting remains enabled. |
| `app/stage/protyle/js/pdf/cmaps/**` and `app/stage/protyle/js/pdf/standard_fonts/**` | PDF.js CMap and font support | `pdfjs-dist`, `cmaps/**`, `standard_fonts/**` | `4.7.76` | Copy from the same pinned PDF.js package. |
| `app/stage/protyle/js/plantuml/plantuml-encoder.min.js` | `plantumlRender.ts` | `plantuml-encoder`, `dist/plantuml-encoder.min.js` | `1.4.0` | Copy from the pinned package tarball. The bundle includes its compression helpers; do not hand-roll it. |
| `app/stage/protyle/js/viewerjs/viewer.js` | `app/src/protyle/preview/image.ts` | `viewerjs`, `dist/viewer.min.js` | `1.11.7` | Copy from the pinned package tarball and keep the staged filename as `viewer.js` because that is what the runtime loader requests. Viewer styling used by the app lives in `app/src/assets/scss/viewerjs/_viewer.scss` and is built with app CSS. |
| `app/stage/protyle/js/vis/vis-network.min.js` | `app/src/layout/dock/Graph.ts` | `vis-network`, `dist/vis-network.min.js` | `9.1.13` | Copy from the pinned package tarball. |

## Preferred Workflow

1. Run `powershell -ExecutionPolicy Bypass -File scripts/regenerate-protyle-vendor-assets.ps1` for the direct-copy package assets.
2. Pin every package version in that script; do not use floating `latest`.
3. Copy only the files listed above into `app/stage/protyle/js`.
4. Apply the Mermaid ZenUML compatibility patch in script form, not by manually editing the generated bundle.
5. Rebuild the frontend with `cd app; pnpm build`.
6. Smoke-test editor features that load these assets: Markdown rendering, code highlighting, math, Mermaid, flowchart, Graphviz, ECharts, PlantUML, PDF preview, image preview, graph dock, and HTML export.

## Cleanup Candidates

`app/stage/protyle/js/abcjs/abcjs-basic-min.min.js` is present but no current Scribli source path references it. Delete it only after confirming no exported/static HTML path expects that name.

`app/stage/protyle/js/highlight.js/styles/**` is large because the code-theme picker can reference many Highlight.js themes. Prune this only together with the settings UI list so theme selection cannot point at missing CSS.

`app/stage/protyle/js/mermaid/icons.json` is the largest support file in this group. It should be regenerated from only the icon packs Scribli registers, but do that in the same change as Mermaid regeneration so diagram rendering is tested once.
