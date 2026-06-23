// Package html renders a Document as a single self-contained, interactive HTML
// report built on vis-network.
//
// The output has zero external dependencies — no CDN, no web fonts, no network
// calls — so it works in air-gapped environments. The vis-network library is
// vendored and embedded into the binary (see assets/), then inlined into each
// report as a base64 data: URI; the scan data rides along as a JSON island and
// is rendered into an interactive dependency graph client-side.
package html

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/Aakhri-Pastaa/infrasight/internal/output"
)

//go:embed assets/vis-network.min.js
var visNetworkJS []byte

const (
	dataToken = "/*__INFRASIGHT_DATA__*/"
	libToken  = "__VIS_NETWORK_B64__"
)

// Render writes a self-contained interactive HTML report for the document.
func Render(w io.Writer, doc output.Document) error {
	// MarshalIndent escapes <, > and & (SetEscapeHTML is on by default), so the
	// payload is safe to embed inside a <script type="application/json"> island.
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	lib := base64.StdEncoding.EncodeToString(visNetworkJS)

	out := strings.Replace(page, libToken, lib, 1)
	out = strings.Replace(out, dataToken, string(data), 1)
	_, err = io.WriteString(w, out)
	return err
}

const page = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>InfraSight Graph</title>
<style>
:root{
  --bg:#0d1117; --panel:#161b22; --panel2:#1c2330; --fg:#e6edf3; --muted:#8b98a8;
  --border:#2b3543; --accent:#4f9cf9;
  --ok:#3fb950; --warn:#d29922; --crit:#f85149;
}
@media (prefers-color-scheme: light){
  :root{ --bg:#f6f8fa; --panel:#fff; --panel2:#f0f3f6; --fg:#1f2328; --muted:#636c76;
         --border:#d0d7de; --accent:#0969da; }
}
*{box-sizing:border-box}
html,body{height:100%;margin:0}
body{background:var(--bg);color:var(--fg);display:flex;flex-direction:column;
  font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;}
header{padding:12px 18px;border-bottom:1px solid var(--border);background:var(--panel);
  display:flex;align-items:center;gap:12px;flex:0 0 auto}
header h1{margin:0;font-size:16px;letter-spacing:.3px;white-space:nowrap}
header .sub{color:var(--muted);font-size:12px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.ro{font-size:11px;color:var(--ok);border:1px solid var(--ok);border-radius:6px;padding:1px 7px}
#app{flex:1;display:flex;min-height:0}
#sidebar{width:300px;flex:0 0 auto;border-right:1px solid var(--border);background:var(--panel);
  overflow:auto;padding:14px}
#net{flex:1;min-width:0;background:var(--bg)}
#detail{width:330px;flex:0 0 auto;border-left:1px solid var(--border);background:var(--panel);
  overflow:auto;padding:14px;display:none}
#detail.open{display:block}
.cards{display:flex;flex-wrap:wrap;gap:8px;margin-bottom:12px}
.card{background:var(--panel2);border:1px solid var(--border);border-radius:8px;padding:8px 10px;flex:1;min-width:64px}
.card .n{font-size:18px;font-weight:600}
.card .l{color:var(--muted);font-size:10px;text-transform:uppercase;letter-spacing:.5px}
.crit .n{color:var(--crit)} .warn .n{color:var(--warn)} .ok .n{color:var(--ok)}
h2{font-size:12px;text-transform:uppercase;letter-spacing:.5px;color:var(--muted);margin:16px 0 8px}
input[type=search],select{width:100%;padding:7px 10px;border-radius:7px;border:1px solid var(--border);
  background:var(--bg);color:var(--fg);margin-bottom:8px;font-size:13px}
.row{display:flex;gap:8px;align-items:center;margin-bottom:8px}
.row label{font-size:13px;color:var(--muted)}
button{background:var(--panel2);color:var(--fg);border:1px solid var(--border);border-radius:7px;
  padding:6px 10px;cursor:pointer;font-size:13px}
button:hover{border-color:var(--accent)}
.chip{display:flex;align-items:center;gap:8px;padding:5px 4px;cursor:pointer;border-radius:6px;font-size:13px}
.chip:hover{background:var(--panel2)}
.chip input{margin:0}
.chip .c{margin-left:auto;color:var(--muted);font-size:12px}
.banner{background:var(--panel2);border-left:3px solid var(--crit);border-radius:6px;
  padding:8px 10px;margin-bottom:8px;font-size:12px}
.banner.warn{border-left-color:var(--warn)}
.banner b{display:block;margin-bottom:3px;font-size:11px;text-transform:uppercase;letter-spacing:.5px}
.banner div{color:var(--muted)}
.legend{display:flex;gap:14px;margin-top:6px;font-size:12px;color:var(--muted)}
.legend span{display:flex;align-items:center;gap:5px}
.dot{width:9px;height:9px;border-radius:50%}
.dot.healthy{background:var(--ok)} .dot.warning{background:var(--warn)} .dot.critical{background:var(--crit)}
#detail .dlabel{font-size:16px;font-weight:600;margin:0 0 2px;word-break:break-word}
#detail .dtype{font-size:12px;color:var(--muted);margin-bottom:10px}
#detail .badge{display:inline-block;font-size:11px;padding:1px 7px;border-radius:6px;border:1px solid var(--border);margin-right:6px}
table.meta{width:100%;border-collapse:collapse;margin:8px 0;font-size:12px}
table.meta td{padding:4px 6px;border-top:1px solid var(--border);vertical-align:top;word-break:break-word}
table.meta td.k{color:var(--muted);white-space:nowrap;width:38%}
.neighbors a{display:block;color:var(--accent);text-decoration:none;font-size:13px;padding:3px 0;cursor:pointer}
.close{float:right;border:none;background:none;color:var(--muted);font-size:18px;cursor:pointer;padding:0}
.vis-tooltip{position:absolute;white-space:pre !important;font-family:ui-monospace,SFMono-Regular,Menlo,monospace !important;
  font-size:12px !important;color:var(--fg) !important;background:var(--panel) !important;
  border:1px solid var(--border) !important;border-radius:8px !important;padding:8px 10px !important;
  box-shadow:0 4px 16px rgba(0,0,0,.35) !important;max-width:360px;pointer-events:none;z-index:9}
.hint{color:var(--muted);font-size:12px;margin-top:18px}
</style>
</head>
<body>
<header>
  <h1>InfraSight</h1>
  <span class="ro">READ-ONLY</span>
  <div class="sub" id="subtitle"></div>
</header>
<div id="app">
  <aside id="sidebar">
    <div class="cards" id="cards"></div>
    <div id="findings"></div>

    <input type="search" id="q" placeholder="Search nodes…" autocomplete="off">

    <div class="row">
      <label for="layout">Layout</label>
      <select id="layout">
        <option value="force">Force-directed</option>
        <option value="hier">Hierarchical</option>
      </select>
    </div>
    <div class="row">
      <input type="checkbox" id="physics" checked>
      <label for="physics">Physics</label>
      <button id="fit" style="margin-left:auto">Fit view</button>
    </div>

    <h2>Types</h2>
    <div id="types"></div>
    <div class="legend">
      <span><i class="dot healthy"></i>healthy</span>
      <span><i class="dot warning"></i>warning</span>
      <span><i class="dot critical"></i>critical</span>
    </div>
    <p class="hint">Click a node for details · double-click to focus · drag to pan · scroll to zoom.</p>
  </aside>

  <div id="net"></div>

  <aside id="detail"></aside>
</div>

<script src="data:text/javascript;base64,__VIS_NETWORK_B64__"></script>
<script type="application/json" id="infrasight-data">
/*__INFRASIGHT_DATA__*/
</script>
<script>
(function(){
  var doc = JSON.parse(document.getElementById('infrasight-data').textContent);
  var netEl = document.getElementById('net');
  if (typeof vis === 'undefined' || !vis.Network) {
    netEl.innerHTML = '<p style="padding:20px;color:#f85149">vis-network failed to load.</p>';
    return;
  }
  var dark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;

  // ---- header + cards ----
  var s = doc.scan || {}, sum = doc.summary || {}, h = sum.healthSummary || {};
  document.getElementById('subtitle').textContent =
    (s.hostname||'host') + '  ·  ' + (s.modules||[]).join(', ') + '  ·  ' +
    (s.durationMs||0) + 'ms  ·  v' + (s.version||'?');
  var cards = [
    {l:'Nodes', n:sum.totalNodes||0, c:''},
    {l:'Edges', n:sum.totalEdges||0, c:''},
    {l:'Warn', n:h.warning||0, c:'warn'},
    {l:'Crit', n:h.critical||0, c:'crit'}
  ];
  document.getElementById('cards').innerHTML = cards.map(function(c){
    return '<div class="card '+c.c+'"><div class="n">'+c.n+'</div><div class="l">'+c.l+'</div></div>';
  }).join('');

  var fhtml='';
  if((doc.security||[]).length){
    fhtml += '<div class="banner"><b>Security ('+doc.security.length+')</b>'+
      doc.security.slice(0,12).map(function(f){
        return '<div>['+esc(f.severity)+'] '+esc(f.title)+(f.detail?' — '+esc(f.detail):'')+'</div>';
      }).join('')+'</div>';
  }
  if((sum.criticals||[]).length) fhtml += banner('Critical', sum.criticals, '');
  if((sum.warnings||[]).length)  fhtml += banner('Warnings', sum.warnings, 'warn');
  document.getElementById('findings').innerHTML = fhtml;
  function banner(t, items, cls){
    return '<div class="banner '+cls+'"><b>'+t+'</b>'+
      items.slice(0,8).map(function(x){return '<div>'+esc(x)+'</div>';}).join('')+'</div>';
  }

  // ---- node styling ----
  var HEALTH = {
    healthy:  {bg:'rgba(63,185,80,.18)',  bd:'#3fb950'},
    warning:  {bg:'rgba(210,153,34,.20)', bd:'#d29922'},
    critical: {bg:'rgba(248,81,73,.20)',  bd:'#f85149'},
    '':       {bg: dark?'rgba(110,118,129,.18)':'rgba(140,152,168,.18)', bd: dark?'#6e7681':'#8b98a8'}
  };
  function shapeFor(t){
    return ({HARDWARE:'box',OS:'star',SERVICE:'hexagon',PROCESS:'dot',CONTAINER:'square',
             DATABASE:'diamond',PORT:'triangle',CERTIFICATE:'triangleDown',WEBSITE:'ellipse',
             NETWORK:'ellipse',CLOUD_RESOURCE:'database'})[t] || 'dot';
  }
  function sizeFor(t){
    return ({OS:30,HARDWARE:22,SERVICE:22,DATABASE:22,CONTAINER:20,WEBSITE:20,PROCESS:16,
             PORT:14,CERTIFICATE:16})[t] || 9;
  }
  var rawById = {};
  var visNodes = doc.nodes.map(function(n){
    rawById[n.id] = n;
    var hc = HEALTH[n.health||''] || HEALTH[''];
    return {
      id:n.id, label:n.label + (n.version ? '\n'+n.version : ''), group:n.type,
      shape:shapeFor(n.type), size:sizeFor(n.type),
      color:{ background:hc.bg, border:hc.bd,
              highlight:{background:hc.bg,border:'#4f9cf9'}, hover:{background:hc.bg,border:'#4f9cf9'} },
      title:tooltip(n)
    };
  });
  var visEdges = doc.edges.map(function(e,i){
    return { id:'e'+i, from:e.source, to:e.target, title:e.relation, arrows:'to' };
  });
  // Returns a PLAIN-TEXT tooltip. It is assigned to vis-network's node title
  // property, which vis renders via textContent (not innerHTML), so a raw '<'
  // or '&' is shown literally and cannot inject markup. Do NOT run this through
  // esc() — that would double-escape and display a literal "&lt;". If you ever
  // switch to an HTML title (an element), you MUST escape every interpolated
  // value then. (Avoid backticks here: this whole template is a Go raw string.)
  function tooltip(n){
    var lines = [n.type + '   ' + (n.health||'unknown')];
    lines.push(n.label + (n.version ? '  '+n.version : ''));
    var m = n.metadata||{}; Object.keys(m).slice(0,8).forEach(function(k){ lines.push('  '+k+': '+m[k]); });
    return lines.join('\n');
  }

  var nodesDS = new vis.DataSet(visNodes);
  var edgesDS = new vis.DataSet(visEdges);

  // ---- type filters (PACKAGE off by default to avoid a hairball) ----
  var counts = {}; doc.nodes.forEach(function(n){ counts[n.type]=(counts[n.type]||0)+1; });
  var typeList = Object.keys(counts).sort();
  var active = {}; typeList.forEach(function(t){ active[t] = (t !== 'PACKAGE'); });
  var search = '';
  var visible = new Set();

  document.getElementById('types').innerHTML = typeList.map(function(t){
    return '<label class="chip"><input type="checkbox" data-t="'+t+'"'+(active[t]?' checked':'')+'>'+
      t+'<span class="c">'+counts[t]+'</span></label>';
  }).join('');
  document.getElementById('types').addEventListener('change', function(e){
    var t = e.target.getAttribute('data-t'); if(!t) return;
    active[t] = e.target.checked; recompute();
  });

  var nodesView = new vis.DataView(nodesDS, { filter:function(node){
    if(!active[node.group]) return false;
    if(search){
      var n = rawById[node.id];
      var hay = (n.label+' '+(n.version||'')+' '+JSON.stringify(n.metadata||{})).toLowerCase();
      if(hay.indexOf(search) < 0) return false;
    }
    return true;
  }});
  var edgesView = new vis.DataView(edgesDS, { filter:function(e){
    return visible.has(e.from) && visible.has(e.to);
  }});
  function recompute(){
    nodesView.refresh();
    visible = new Set(nodesView.getIds());
    edgesView.refresh();
  }
  recompute();

  // ---- network ----
  var fontColor = dark ? '#e6edf3' : '#1f2328';
  var edgeColor = dark ? '#46506080' : '#9aa7b8';
  var options = {
    autoResize:true,
    nodes:{ borderWidth:2, font:{ color:fontColor, size:13 } },
    edges:{ width:1.2, color:{ color:edgeColor, highlight:'#4f9cf9', hover:'#4f9cf9' },
            arrows:{ to:{ enabled:true, scaleFactor:0.55 } },
            smooth:{ enabled:true, type:'dynamic' } },
    interaction:{ hover:true, tooltipDelay:120, multiselect:false, navigationButtons:false, keyboard:false },
    physics:{ enabled:true, stabilization:{ enabled:true, iterations:250, fit:true },
              barnesHut:{ gravitationalConstant:-9000, springLength:130, springConstant:0.04,
                          damping:0.45, avoidOverlap:0.2 } },
    layout:{ improvedLayout:true }
  };
  var network = new vis.Network(netEl, { nodes:nodesView, edges:edgesView }, options);

  network.on('selectNode', function(p){ showDetail(p.nodes[0]); });
  network.on('doubleClick', function(p){ if(p.nodes.length) network.focus(p.nodes[0], {scale:1.4, animation:true}); });

  // ---- controls ----
  var physicsChk = document.getElementById('physics');
  document.getElementById('q').addEventListener('input', function(e){ search = e.target.value.trim().toLowerCase(); recompute(); });
  document.getElementById('fit').addEventListener('click', function(){ network.fit({animation:true}); });
  physicsChk.addEventListener('change', function(){
    if(document.getElementById('layout').value === 'force')
      network.setOptions({ physics:{ enabled: physicsChk.checked } });
  });
  document.getElementById('layout').addEventListener('change', function(e){
    if(e.target.value === 'hier'){
      network.setOptions({ layout:{ hierarchical:{ enabled:true, direction:'UD',
        sortMethod:'directed', shakeTowards:'roots', levelSeparation:120, nodeSpacing:150 } },
        physics:{ enabled:false } });
    } else {
      network.setOptions({ layout:{ hierarchical:false }, physics:{ enabled: physicsChk.checked } });
    }
    network.fit({animation:true});
  });

  // ---- detail panel ----
  var detail = document.getElementById('detail');
  function showDetail(id){
    var n = rawById[id]; if(!n) return;
    var rows = '';
    var m = n.metadata||{}; Object.keys(m).forEach(function(k){
      rows += '<tr><td class="k">'+esc(k)+'</td><td>'+esc(String(m[k]))+'</td></tr>';
    });
    var nb = network.getConnectedNodes(id).map(function(x){
      var r = rawById[x]; return '<a data-id="'+esc(x)+'">'+esc(r?r.label:x)+'</a>';
    }).join('') || '<div style="color:var(--muted);font-size:12px">none in view</div>';

    detail.innerHTML =
      '<button class="close" title="Close">×</button>'+
      '<p class="dlabel">'+esc(n.label)+'</p>'+
      '<div class="dtype">'+esc(n.id)+'</div>'+
      '<div><span class="badge">'+esc(n.type)+'</span>'+
        '<span class="badge">'+esc(n.status||'')+'</span>'+
        (n.health?'<span class="badge"><i class="dot '+n.health+'"></i> '+esc(n.health)+'</span>':'')+
        (n.version?'<span class="badge">v'+esc(n.version)+'</span>':'')+'</div>'+
      (rows?'<table class="meta">'+rows+'</table>':'')+
      '<h2>Connected</h2><div class="neighbors">'+nb+'</div>';
    detail.classList.add('open');
    detail.querySelector('.close').onclick = function(){ detail.classList.remove('open'); network.unselectAll(); };
    detail.querySelectorAll('.neighbors a').forEach(function(a){
      a.onclick = function(){ var t=a.getAttribute('data-id');
        network.selectNodes([t]); network.focus(t,{scale:1.3,animation:true}); showDetail(t); };
    });
  }

  function esc(s){ return String(s).replace(/[&<>"]/g, function(c){
    return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]; }); }
})();
</script>
</body>
</html>
`
