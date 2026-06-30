import React, { useState, useEffect, useRef } from 'react';
import { 
  Activity, Shield, Package, Maximize2, X, Search, ShieldAlert, Cpu, 
  HardDrive, Layout, Play, RefreshCw, Settings, Info, Check, GitCommit, 
  ChevronRight, Zap, Download, FileSpreadsheet, ExternalLink, Columns, 
  Layers, HelpCircle, User, Bell, Network, Terminal, ShieldCheck, AlertTriangle
} from 'lucide-react';

export default function App() {
  const [doc, setDoc] = useState(null);
  const [activeView, setActiveView] = useState('dashboard');
  const [searchQuery, setSearchQuery] = useState('');
  
  // Graph settings
  const [layoutMode, setLayoutMode] = useState('force');
  const [enablePhysics, setEnablePhysics] = useState(true);
  const [showPackages, setShowPackages] = useState(false);
  const [showOrphans, setShowOrphans] = useState(true);
  const [activeLayers, setActiveLayers] = useState({
    foundation: true, // Compute, Network
    living: true,     // Processes, Containers
    exposed: true     // Routes, Certs
  });
  const [showSettingsPopover, setShowSettingsPopover] = useState(false);

  // Selected Node Details
  const [selectedNode, setSelectedNode] = useState(null);
  const [detailTab, setDetailTab] = useState('business');
  
  // Packages Inventory
  const [pkgSearch, setPkgSearch] = useState('');
  const [pkgSourceFilter, setPkgSourceFilter] = useState('all');

  const netContainerRef = useRef(null);
  const networkRef = useRef(null);
  const nodesDSRef = useRef(null);
  const edgesDSRef = useRef(null);
  const nodesViewRef = useRef(null);
  const edgesViewRef = useRef(null);
  const rawByIdRef = useRef({});

  // 1. Data Ingestion & Setup
  useEffect(() => {
    const el = document.getElementById('infrasight-data');
    if (el) {
      try {
        const rawText = el.textContent.trim();
        if (rawText.startsWith('/*__') || !rawText) {
          fetchMockData();
        } else {
          const parsed = JSON.parse(rawText);
          initData(parsed);
        }
      } catch (e) {
        fetchMockData();
      }
    } else {
      fetchMockData();
    }
  }, []);

  const fetchMockData = () => {
    const mock = {
      scan: { 
        id: "scan_20260630_120000",
        hostname: "prod-k8s-node-01", 
        startedAt: "2026-06-30T12:00:00Z", 
        completedAt: "2026-06-30T12:00:01Z",
        durationMs: 450, 
        version: "0.1.2", 
        modules: ["hardware", "os", "network", "services", "web", "database", "resources"] 
      },
      summary: { 
        totalNodes: 14, 
        totalEdges: 12, 
        healthSummary: { healthy: 11, warning: 2, critical: 1 }, 
        criticals: ["cert:sha256:infrasight.io"], 
        warnings: ["hardware:memory", "hardware:disk:/var"] 
      },
      nodes: [
        { id: "os:linux:ubuntu", type: "OS", label: "Ubuntu 22.04 LTS", status: "active", health: "healthy", metadata: { kernel: "5.15.0-generic", hostname: "prod-k8s-node-01" } },
        { id: "hardware:cpu", type: "HARDWARE", label: "AMD EPYC (8 Cores)", status: "active", health: "healthy", metadata: { utilizationPercent: 42, cores: 8, load1: 1.25 } },
        { id: "hardware:memory", type: "HARDWARE", label: "RAM 16 GiB", status: "active", health: "warning", metadata: { usedPercent: 78, totalKB: 16777216, usedKB: 13086228 } },
        { id: "hardware:disk:/var", type: "HARDWARE", label: "/var partition", status: "active", health: "warning", metadata: { usedPercent: 82, totalBytes: 50000000000, usedBytes: 41000000000, mountpoint: "/var", device: "/dev/sda2" } },
        { id: "process:nginx", type: "PROCESS", label: "nginx (PID: 142)", status: "active", health: "healthy", metadata: { mainPid: 142, description: "Nginx web server process", cpuPercent: 2.4, memoryMB: 128 } },
        { id: "port:tcp:443", type: "PORT", label: ":443 (HTTPS)", status: "active", health: "healthy", metadata: { protocol: "tcp", port: "443", bindAddress: "0.0.0.0" } },
        { id: "website:api.infrasight.io", type: "WEBSITE", label: "api.infrasight.io", status: "active", health: "healthy", metadata: { server: "nginx", root: "/var/www/api", config: "/etc/nginx/sites-enabled/api.conf" } },
        { id: "cert:sha256:infrasight.io", type: "CERTIFICATE", label: "infrasight.io TLS Cert", status: "active", health: "critical", metadata: { issuer: "Let's Encrypt", subject: "*.infrasight.io", daysUntilExpiry: 2, keyType: "RSA", keyBits: 1024 } },
        { id: "database:postgres", type: "DATABASE", label: "postgres-main", status: "active", health: "healthy", metadata: { engine: "PostgreSQL 15.3", port: "5432", storagePath: "/var/lib/postgresql/data", uptime: "45d 12h" } },
        { id: "package:dpkg:nginx", type: "PACKAGE", label: "nginx", status: "active", health: "healthy", metadata: { manager: "apt", version: "1.24.0-1ubuntu1", installedSizeKB: 2048 } },
        { id: "package:dpkg:postgres", type: "PACKAGE", label: "postgresql-15", status: "active", health: "healthy", metadata: { manager: "apt", version: "15.3-1.pgdg22.04+1", installedSizeKB: 4096 } },
        { id: "package:dpkg:containerd", type: "PACKAGE", label: "containerd.io", status: "active", health: "healthy", metadata: { manager: "apt", version: "1.6.24-1", installedSizeKB: 3200 } },
        { id: "package:pip:requests", type: "PACKAGE", label: "requests", status: "active", health: "healthy", metadata: { manager: "pip", version: "2.31.0" } },
        { id: "package:npm:express", type: "PACKAGE", label: "express", status: "active", health: "healthy", metadata: { manager: "npm", version: "4.18.2" } }
      ],
      edges: [
        { source: "os:linux:ubuntu", target: "hardware:cpu", relation: "RUNS_ON" },
        { source: "os:linux:ubuntu", target: "hardware:memory", relation: "RUNS_ON" },
        { source: "os:linux:ubuntu", target: "hardware:disk:/var", relation: "RUNS_ON" },
        { source: "process:nginx", target: "port:tcp:443", relation: "LISTENS_ON" },
        { source: "port:tcp:443", target: "website:api.infrasight.io", relation: "SERVES" },
        { source: "cert:sha256:infrasight.io", target: "website:api.infrasight.io", relation: "ENCRYPTS" },
        { source: "process:nginx", target: "database:postgres", relation: "CONNECTS_TO" },
        { source: "process:nginx", target: "package:dpkg:nginx", relation: "DEPENDS_ON" },
        { source: "database:postgres", target: "package:dpkg:postgres", relation: "DEPENDS_ON" }
      ],
      security: [
        { id: "cert-weak-key", severity: "critical", title: "Weak TLS Key Size", nodeId: "cert:sha256:infrasight.io", detail: "infrasight.io TLS Cert uses a 1024-bit RSA key (< 2048)" },
        { id: "cert-expiring", severity: "high", title: "TLS Certificate Expiring Soon", nodeId: "cert:sha256:infrasight.io", detail: "infrasight.io TLS Cert expires in 2 day(s)" },
        { id: "port-world-listening", severity: "low", title: "Port Exposed to All Interfaces", nodeId: "port:tcp:443", detail: "port 443 is bound to 0.0.0.0" }
      ]
    };
    initData(mock);
  };

  const initData = (parsed) => {
    setDoc(parsed);
    const mapping = {};
    parsed.nodes.forEach(n => { mapping[n.id] = n; });
    rawByIdRef.current = mapping;
  };

  // 2. vis-network graph binding
  useEffect(() => {
    if (!doc || !netContainerRef.current || activeView !== 'graph') return;

    const HEALTH_STYLING = {
      healthy:  { bg: 'rgba(52, 211, 153, 0.12)', bd: '#34d399' },
      warning:  { bg: 'rgba(245, 158, 11, 0.15)', bd: '#f59e0b' },
      critical: { bg: 'rgba(239, 68, 68, 0.15)', bd: '#ef4444' },
      '':       { bg: 'rgba(113, 113, 122, 0.12)', bd: '#71717a' }
    };

    const shapeFor = (t) => ({
      HARDWARE: 'box', OS: 'star', SERVICE: 'hexagon', PROCESS: 'dot', CONTAINER: 'square',
      DATABASE: 'database', PORT: 'triangle', CERTIFICATE: 'triangleDown', WEBSITE: 'ellipse',
      NETWORK: 'ellipse', CLOUD_RESOURCE: 'database'
    })[t] || 'dot';

    const sizeFor = (t) => ({ OS: 24, HARDWARE: 18, SERVICE: 18, DATABASE: 18, CONTAINER: 16, WEBSITE: 16, PROCESS: 14, PORT: 12, CERTIFICATE: 14 })[t] || 10;

    const filteredNodes = doc.nodes.map(n => {
      const hc = HEALTH_STYLING[n.health || ''] || HEALTH_STYLING[''];
      return {
        id: n.id,
        label: n.label + (n.version ? '\nv' + n.version : ''),
        group: n.type,
        shape: shapeFor(n.type),
        size: sizeFor(n.type),
        color: {
          background: hc.bg,
          border: hc.bd,
          highlight: { background: hc.bg, border: '#a78bfa' },
          hover: { background: hc.bg, border: '#a78bfa' }
        },
        title: n.type + ': ' + n.label
      };
    });

    const filteredEdges = doc.edges.map((e, i) => ({
      id: 'e' + i,
      from: e.source,
      to: e.target,
      title: e.relation,
      arrows: 'to'
    }));

    nodesDSRef.current = new vis.DataSet(filteredNodes);
    edgesDSRef.current = new vis.DataSet(filteredEdges);

    // Apply layers & visibility filters
    nodesViewRef.current = new vis.DataView(nodesDSRef.current, {
      filter: (node) => {
        const rawNode = rawByIdRef.current[node.id];
        if (!rawNode) return false;

        // Package filtering
        if (rawNode.type === 'PACKAGE' && !showPackages) return false;

        // Orphans filtering (simulate nodes with no connections)
        if (!showOrphans) {
          const hasConnection = doc.edges.some(e => e.source === node.id || e.target === node.id);
          if (!hasConnection) return false;
        }

        // Layer checking
        const isFoundation = ['OS', 'HARDWARE', 'NETWORK', 'PORT'].includes(rawNode.type);
        const isLiving = ['PROCESS', 'SERVICE', 'CONTAINER'].includes(rawNode.type);
        const isExposed = ['WEBSITE', 'CERTIFICATE', 'DATABASE'].includes(rawNode.type);

        if (isFoundation && !activeLayers.foundation) return false;
        if (isLiving && !activeLayers.living) return false;
        if (isExposed && !activeLayers.exposed) return false;

        // Search text matching
        if (searchQuery) {
          const haystack = (rawNode.label + ' ' + (rawNode.version || '') + ' ' + JSON.stringify(rawNode.metadata || '')).toLowerCase();
          if (haystack.indexOf(searchQuery.toLowerCase()) < 0) return false;
        }

        return true;
      }
    });

    edgesViewRef.current = new vis.DataView(edgesDSRef.current, {
      filter: (edge) => {
        const visible = new Set(nodesViewRef.current.getIds());
        return visible.has(edge.from) && visible.has(edge.to);
      }
    });

    const options = {
      autoResize: true,
      nodes: { borderWidth: 1.5, font: { color: '#fafafa', size: 11, face: 'Geist' } },
      edges: {
        width: 1.2,
        color: { color: 'rgba(161, 161, 170, 0.25)', highlight: '#a78bfa', hover: '#a78bfa' },
        arrows: { to: { enabled: true, scaleFactor: 0.5 } },
        smooth: { enabled: true, type: 'dynamic' }
      },
      interaction: { hover: true, tooltipDelay: 100 },
      physics: {
        enabled: enablePhysics,
        stabilization: { enabled: true, iterations: 120, fit: true },
        barnesHut: { gravitationalConstant: -4000, springLength: 90, springConstant: 0.04, damping: 0.45 }
      }
    };

    const net = new vis.Network(netContainerRef.current, {
      nodes: nodesViewRef.current,
      edges: edgesViewRef.current
    }, options);

    networkRef.current = net;

    net.on('selectNode', (params) => {
      if (params.nodes.length > 0) {
        setSelectedNode(rawByIdRef.current[params.nodes[0]]);
      }
    });

    net.on('deselectNode', () => {
      setSelectedNode(null);
    });

    return () => {
      if (networkRef.current) {
        networkRef.current.destroy();
      }
    };
  }, [doc, activeView]);

  // Handle updates to layout modes and physics configurations
  useEffect(() => {
    if (!networkRef.current || !doc) return;

    if (layoutMode === 'concentric') {
      networkRef.current.setOptions({ layout: { hierarchical: false }, physics: { enabled: false } });
      applyConcentricCoordinates();
    } else if (layoutMode === 'hierarchical') {
      networkRef.current.setOptions({
        layout: { hierarchical: { enabled: true, direction: 'UD', sortMethod: 'directed', levelSeparation: 90, nodeSpacing: 100 } },
        physics: { enabled: false }
      });
    } else {
      networkRef.current.setOptions({
        layout: { hierarchical: false },
        physics: { enabled: enablePhysics }
      });
      // Release fixed concentric coords
      const releases = doc.nodes.map(n => ({ id: n.id, fixed: false }));
      if (nodesDSRef.current) {
        nodesDSRef.current.update(releases);
      }
    }
    networkRef.current.fit({ animation: true });
  }, [layoutMode, enablePhysics]);

  // Live filter refreshes
  useEffect(() => {
    if (nodesViewRef.current && edgesViewRef.current) {
      nodesViewRef.current.refresh();
      edgesViewRef.current.refresh();
    }
  }, [showPackages, showOrphans, activeLayers, searchQuery]);

  const applyConcentricCoordinates = () => {
    if (!nodesDSRef.current || !doc) return;
    const osNodes = [];
    const hwNodes = [];
    const procNodes = [];
    const svcNodes = [];
    const portNodes = [];
    const webNodes = [];
    const otherNodes = [];

    doc.nodes.forEach(n => {
      if (n.type === 'PACKAGE') return;
      if (n.type === 'OS') osNodes.push(n);
      else if (n.type === 'HARDWARE') hwNodes.push(n);
      else if (n.type === 'PROCESS') procNodes.push(n);
      else if (n.type === 'SERVICE' || n.type === 'CONTAINER') svcNodes.push(n);
      else if (n.type === 'PORT') portNodes.push(n);
      else if (n.type === 'WEBSITE' || n.type === 'CERTIFICATE') webNodes.push(n);
      else otherNodes.push(n);
    });

    const updates = [];
    osNodes.forEach(n => updates.push({ id: n.id, x: 0, y: 0, fixed: true }));
    positionOnCircle(hwNodes, 100, updates);
    positionOnCircle(procNodes, 220, updates);
    positionOnCircle(svcNodes, 320, updates);
    positionOnCircle(portNodes, 440, updates);
    positionOnCircle(webNodes, 560, updates);
    positionOnCircle(otherNodes, 680, updates);

    nodesDSRef.current.update(updates);
  };

  const positionOnCircle = (arr, radius, updates) => {
    const len = arr.length;
    for (let i = 0; i < len; i++) {
      const angle = (i * 2 * Math.PI) / len;
      updates.push({
        id: arr[i].id,
        x: radius * Math.cos(angle),
        y: radius * Math.sin(angle),
        fixed: true
      });
    }
  };

  const focusOnNode = (id) => {
    const node = rawByIdRef.current[id];
    if (node) {
      if (node.type === 'PACKAGE' && !showPackages) {
        setShowPackages(true);
      }
      setSelectedNode(node);
      setActiveView('graph');
      // Wait a moment for layout mount if switching view
      setTimeout(() => {
        if (networkRef.current) {
          networkRef.current.selectNodes([id]);
          networkRef.current.focus(id, { scale: 1.3, animation: true });
        }
      }, 100);
    }
  };

  const formatBytes = (bytes) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  if (!doc) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#09090b] text-[#fafafa]">
        <div className="text-center">
          <div className="w-8 h-8 rounded-full border-2 border-primary border-t-transparent animate-spin mx-auto mb-4"></div>
          <p className="text-sm font-semibold tracking-wider uppercase text-on-surface-variant">Loading Command Center Panel...</p>
        </div>
      </div>
    );
  }

  // CPU/RAM metrics calculation
  let cpuPercent = 42; 
  let memPercent = 78;
  const cpuNode = doc.nodes.find(n => n.id === 'hardware:cpu');
  if (cpuNode && cpuNode.metadata) cpuPercent = Math.round(cpuNode.metadata.utilizationPercent || cpuPercent);
  const memNode = doc.nodes.find(n => n.id === 'hardware:memory');
  if (memNode && memNode.metadata) memPercent = Math.round(memNode.metadata.usedPercent || memPercent);

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-[#09090b] flex">
      {/* Background Ambient Glows */}
      <div className="ambient-glow glow-1"></div>
      <div className="ambient-glow glow-2"></div>

      {/* Side Navigation Bar */}
      <nav className="hidden md:flex flex-col h-full w-64 bg-surface border-r border-outline-variant p-4 flex-shrink-0 z-20 relative">
        <div className="flex items-center gap-3 mb-8 px-2">
          <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center flex-shrink-0">
            <span className="material-symbols-outlined text-on-primary fill-icon" style={{ fontSize: '20px' }}>terminal</span>
          </div>
          <div>
            <h1 className="text-lg font-headline font-black text-primary tracking-tighter">InfraSight</h1>
            <p className="text-[10px] text-on-surface-variant tracking-wider uppercase font-bold">Infrastructure Ops</p>
          </div>
        </div>

        <div className="flex flex-col gap-1 flex-1 font-body text-sm font-medium">
          <button 
            onClick={() => { setActiveView('dashboard'); setSelectedNode(null); }}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all text-left ${activeView === 'dashboard' ? 'bg-primary-container text-on-primary-container' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high'}`}
          >
            <span className={`material-symbols-outlined ${activeView === 'dashboard' ? 'fill-icon' : ''}`}>dashboard</span>
            <span>Dashboard</span>
          </button>
          
          <button 
            onClick={() => { setActiveView('graph'); }}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all text-left ${activeView === 'graph' ? 'bg-primary-container text-on-primary-container' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high'}`}
          >
            <span className={`material-symbols-outlined ${activeView === 'graph' ? 'fill-icon' : ''}`}>account_tree</span>
            <span>Graph View</span>
          </button>

          <button 
            onClick={() => { setActiveView('security'); setSelectedNode(null); }}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all text-left ${activeView === 'security' ? 'bg-primary-container text-on-primary-container' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high'}`}
          >
            <span className={`material-symbols-outlined ${activeView === 'security' ? 'fill-icon' : ''}`}>security</span>
            <span>Security Findings</span>
          </button>

          <button 
            onClick={() => { setActiveView('packages'); setSelectedNode(null); }}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all text-left ${activeView === 'packages' ? 'bg-primary-container text-on-primary-container' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high'}`}
          >
            <span className={`material-symbols-outlined ${activeView === 'packages' ? 'fill-icon' : ''}`}>package</span>
            <span>Packages Catalog</span>
          </button>

          <button 
            onClick={() => { setActiveView('settings'); setSelectedNode(null); }}
            className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all text-left ${activeView === 'settings' ? 'bg-primary-container text-on-primary-container' : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high'}`}
          >
            <span className={`material-symbols-outlined ${activeView === 'settings' ? 'fill-icon' : ''}`}>settings</span>
            <span>Settings</span>
          </button>
        </div>

        {/* User Account Info & CTA */}
        <div className="mt-auto border-t border-outline-variant pt-4 flex flex-col gap-4">
          <button className="w-full bg-primary hover:bg-primary/95 text-on-primary font-medium py-2 rounded transition-colors text-sm flex items-center justify-center gap-2">
            <span className="material-symbols-outlined text-sm">rocket_launch</span>
            Deploy Agent
          </button>
          
          <div className="flex items-center gap-3 px-2">
            <div className="w-8 h-8 rounded-full bg-surface-container-highest border border-outline-variant flex items-center justify-center font-bold text-xs text-primary">
              OA
            </div>
            <div className="flex flex-col">
              <span className="text-xs font-semibold">Ops Admin</span>
              <span className="text-[10px] text-tertiary flex items-center gap-1">
                <span className="w-1.5 h-1.5 rounded-full bg-tertiary animate-pulse"></span> Active
              </span>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Container */}
      <div className="flex-1 flex flex-col min-w-0 relative z-10">
        
        {/* Top Navbar */}
        <header className="flex justify-between items-center w-full px-6 h-16 bg-surface border-b border-outline-variant flex-shrink-0">
          <div className="flex items-center w-full max-w-md">
            <div className="md:hidden mr-4 text-lg font-black text-primary tracking-tighter">
              InfraSight
            </div>
            
            {/* Search Input */}
            <div className="relative w-full group">
              <span className="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant group-focus-within:text-primary transition-colors" style={{ fontSize: '18px' }}>search</span>
              <input 
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-surface-container-low border border-outline-variant text-on-surface placeholder:text-on-surface-variant text-sm rounded py-1.5 pl-10 pr-4 focus:outline-none focus:ring-1 focus:ring-primary focus:border-transparent transition-all" 
                placeholder="Search across infrastructure..." 
                type="text"
              />
              <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1 pointer-events-none">
                <kbd className="px-1.5 py-0.5 text-[9px] font-mono text-on-surface-variant bg-surface-container rounded border border-outline-variant">⌘</kbd>
                <kbd className="px-1.5 py-0.5 text-[9px] font-mono text-on-surface-variant bg-surface-container rounded border border-outline-variant">K</kbd>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button className="w-8 h-8 flex items-center justify-center text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest rounded-full transition-all relative">
              <span className="material-symbols-outlined" style={{ fontSize: '20px' }}>notifications</span>
              <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-primary rounded-full"></span>
            </button>
            <button className="w-8 h-8 flex items-center justify-center text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest rounded-full transition-all">
              <span className="material-symbols-outlined" style={{ fontSize: '20px' }}>help</span>
            </button>
            <div className="w-px h-4 bg-outline-variant mx-1"></div>
            <button className="flex items-center gap-2 text-on-surface-variant hover:text-on-surface transition-all rounded p-1">
              <span className="material-symbols-outlined" style={{ fontSize: '22px' }}>account_circle</span>
            </button>
          </div>
        </header>

        {/* Content Router */}
        <main className="flex-1 overflow-hidden relative">
          
          {/* VIEW: DASHBOARD */}
          {activeView === 'dashboard' && (
            <div className="h-full overflow-y-auto custom-scrollbar p-6 bg-background space-y-6">
              {/* Page Header */}
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                <div>
                  <div className="flex items-center gap-3">
                    <h2 className="text-2xl font-headline font-bold tracking-tight">{doc.scan.hostname}</h2>
                    <span className="bg-error-container text-error text-[10px] font-mono px-2 py-0.5 rounded-full border border-error/30 flex items-center gap-1 uppercase font-bold">
                      <span className="w-1.5 h-1.5 bg-error rounded-full animate-pulse"></span>
                      Critical Findings
                    </span>
                  </div>
                  <p className="text-on-surface-variant text-xs font-body mt-1">Linux Environment · Modules: {doc.scan.modules.join(', ')} · v{doc.scan.version}</p>
                </div>
                <div className="flex gap-2">
                  <button className="px-3 py-1.5 bg-surface border border-outline-variant rounded text-xs font-medium hover:bg-surface-container-high transition-colors flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-xs">download</span> Export Report
                  </button>
                  <button className="px-3 py-1.5 bg-surface border border-outline-variant text-primary rounded text-xs font-medium hover:border-primary/50 transition-colors flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-xs">refresh</span> Force Scan
                  </button>
                </div>
              </div>

              {/* Bento Stats Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {/* Stats 1: Nodes / Edges */}
                <div className="bg-surface-container border border-outline-variant rounded-xl p-5 flex flex-col justify-between relative overflow-hidden group">
                  <div className="absolute inset-0 bg-gradient-to-br from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"></div>
                  <div className="flex justify-between items-start mb-4">
                    <span className="text-on-surface-variant text-xs font-semibold flex items-center gap-2">
                      <span className="material-symbols-outlined text-primary" style={{ fontSize: '18px' }}>lan</span>
                      Infrastructure Size
                    </span>
                    <span className="text-tertiary text-[10px] bg-tertiary/10 border border-tertiary/20 px-2 py-0.5 rounded font-mono">+4 this run</span>
                  </div>
                  <div>
                    <div className="flex items-baseline gap-2">
                      <h3 className="text-3xl font-display font-bold text-on-surface">{doc.summary.totalNodes}</h3>
                      <span className="text-on-surface-variant text-xs">nodes</span>
                    </div>
                    <div className="flex items-baseline gap-2 mt-1">
                      <h4 className="text-lg font-display font-medium text-on-surface-variant">{doc.summary.totalEdges}</h4>
                      <span className="text-on-surface-variant text-[10px]">inter-dependencies</span>
                    </div>
                  </div>
                </div>

                {/* Stats 2: Health Tally */}
                <div className="bg-surface-container border border-outline-variant rounded-xl p-5 flex flex-col justify-between">
                  <div className="flex justify-between items-start mb-3">
                    <span className="text-on-surface-variant text-xs font-semibold flex items-center gap-2">
                      <span className="material-symbols-outlined text-tertiary" style={{ fontSize: '18px' }}>monitor_heart</span>
                      Health Tally
                    </span>
                  </div>
                  <div>
                    <div className="flex justify-between text-xs mb-1.5 font-medium">
                      <span className="text-tertiary">Healthy</span>
                      <span className="text-on-surface font-mono">82%</span>
                    </div>
                    <div className="w-full h-2 bg-surface-container-highest rounded-full overflow-hidden flex mb-3">
                      <div className="bg-tertiary h-full" style={{ width: '82%' }}></div>
                      <div className="bg-yellow-500 h-full" style={{ width: '12%' }}></div>
                      <div className="bg-error h-full" style={{ width: '6%' }}></div>
                    </div>
                    <div className="flex justify-between text-[10px] text-on-surface-variant font-mono">
                      <span className="flex items-center gap-1"><span className="w-1.5 h-1.5 rounded-full bg-yellow-500"></span> Warning (2)</span>
                      <span className="flex items-center gap-1"><span className="w-1.5 h-1.5 rounded-full bg-error"></span> Critical (1)</span>
                    </div>
                  </div>
                </div>

                {/* Stats 3: Top Findings */}
                <div className="bg-surface-container border border-outline-variant rounded-xl p-5 flex flex-col justify-between">
                  <div className="flex justify-between items-start mb-2">
                    <span className="text-on-surface-variant text-xs font-semibold flex items-center gap-2">
                      <span className="material-symbols-outlined text-error" style={{ fontSize: '18px' }}>security</span>
                      Top Findings
                    </span>
                    <button onClick={() => setActiveView('security')} className="text-[10px] text-primary hover:underline font-bold">View All</button>
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <div className="flex justify-between items-center bg-surface-container-high px-2 py-1.5 rounded border border-outline-variant/50">
                      <span className="text-xs text-on-surface truncate">Weak TLS Key Size</span>
                      <span className="text-[9px] px-1 bg-error-container text-error rounded font-mono font-bold">CRIT</span>
                    </div>
                    <div className="flex justify-between items-center bg-surface-container-high px-2 py-1.5 rounded border border-outline-variant/50">
                      <span className="text-xs text-on-surface truncate">Port Exposed (443)</span>
                      <span className="text-[9px] px-1 bg-surface-container text-on-surface-variant rounded font-mono font-bold">LOW</span>
                    </div>
                  </div>
                </div>

                {/* Stats 4: Drift Summary */}
                <div className="bg-surface-container border border-outline-variant rounded-xl p-5 flex flex-col justify-between">
                  <div className="flex justify-between items-start mb-2">
                    <span className="text-on-surface-variant text-xs font-semibold flex items-center gap-2">
                      <span className="material-symbols-outlined text-primary" style={{ fontSize: '18px' }}>history</span>
                      Drift (Last 24h)
                    </span>
                  </div>
                  <div className="text-center py-1">
                    <div className="text-3xl font-display font-bold text-on-surface">3</div>
                    <div className="text-[10px] text-on-surface-variant mt-0.5">untracked config drift events</div>
                  </div>
                  <div className="text-[9px] text-center border-t border-outline-variant/50 pt-2 text-on-surface-variant font-mono">
                    Last scan: <span className="text-on-surface">14 mins ago</span>
                  </div>
                </div>
              </div>

              {/* Health Hotspots Section */}
              <div className="space-y-4 pt-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-base font-headline font-semibold flex items-center gap-2">
                    <span className="material-symbols-outlined text-error">local_fire_department</span>
                    Health Hotspots
                  </h3>
                  <div className="flex gap-1.5 bg-surface-container border border-outline-variant p-0.5 rounded-lg text-[10px] font-medium">
                    <button className="px-2.5 py-1 rounded bg-primary text-on-primary shadow">By Risk</button>
                    <button className="px-2.5 py-1 rounded text-on-surface-variant hover:text-on-surface transition-colors">By CPU</button>
                  </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                  {/* Hotspot 1 */}
                  <div className="bg-surface-container border border-error/50 rounded-xl p-4 flex flex-col relative overflow-hidden">
                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-error"></div>
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <h4 className="font-mono text-sm text-on-surface font-semibold">infrasight.io TLS Cert</h4>
                        <p className="text-[10px] text-on-surface-variant">Type: Certificate</p>
                      </div>
                      <span className="material-symbols-outlined text-error fill-icon">error</span>
                    </div>
                    <div className="space-y-2 text-xs">
                      <div className="flex justify-between border-b border-outline-variant/50 pb-1.5">
                        <span className="text-on-surface-variant">Risk Level</span>
                        <span className="text-error font-bold font-mono">CRITICAL</span>
                      </div>
                      <div className="flex justify-between border-b border-outline-variant/50 pb-1.5">
                        <span className="text-on-surface-variant">Time to Expiry</span>
                        <span className="text-error font-mono">2 days remaining</span>
                      </div>
                      <div className="flex justify-between pb-0.5">
                        <span className="text-on-surface-variant">Key Strength</span>
                        <span className="text-on-surface font-mono">RSA 1024-bit (Weak)</span>
                      </div>
                    </div>
                    <button 
                      onClick={() => focusOnNode("cert:sha256:infrasight.io")}
                      className="mt-4 w-full py-1.5 text-xs font-semibold border border-outline-variant rounded hover:bg-surface-container-high transition-all text-on-surface"
                    >
                      Investigate in Graph
                    </button>
                  </div>

                  {/* Hotspot 2 */}
                  <div className="bg-surface-container border border-yellow-500/50 rounded-xl p-4 flex flex-col relative overflow-hidden">
                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-yellow-500"></div>
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <h4 className="font-mono text-sm text-on-surface font-semibold">RAM 16 GiB</h4>
                        <p className="text-[10px] text-on-surface-variant">Type: Hardware</p>
                      </div>
                      <span className="material-symbols-outlined text-yellow-500 fill-icon">warning</span>
                    </div>
                    <div className="space-y-2 text-xs">
                      <div className="flex justify-between border-b border-outline-variant/50 pb-1.5">
                        <span className="text-on-surface-variant">Risk Level</span>
                        <span className="text-yellow-500 font-bold font-mono">WARNING</span>
                      </div>
                      <div className="flex justify-between border-b border-outline-variant/50 pb-1.5">
                        <span className="text-on-surface-variant">Memory Usage</span>
                        <span className="text-yellow-500 font-mono">78% (12.4 GiB used)</span>
                      </div>
                      <div className="flex justify-between pb-0.5">
                        <span className="text-on-surface-variant">Manager Unit</span>
                        <span className="text-on-surface font-mono">System Limit Bound</span>
                      </div>
                    </div>
                    <button 
                      onClick={() => focusOnNode("hardware:memory")}
                      className="mt-4 w-full py-1.5 text-xs font-semibold border border-outline-variant rounded hover:bg-surface-container-high transition-all text-on-surface"
                    >
                      Investigate in Graph
                    </button>
                  </div>

                  {/* Hotspot 3 */}
                  <div className="bg-surface-container border border-yellow-500/50 rounded-xl p-4 flex flex-col relative overflow-hidden">
                    <div className="absolute left-0 top-0 bottom-0 w-1 bg-yellow-500"></div>
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <h4 className="font-mono text-sm text-on-surface font-semibold">/var partition</h4>
                        <p className="text-[10px] text-on-surface-variant">Type: Hardware</p>
                      </div>
                      <span className="material-symbols-outlined text-yellow-500 fill-icon">warning</span>
                    </div>
                    <div className="space-y-2 text-xs">
                      <div className="flex justify-between border-b border-outline-variant/50 pb-1.5">
                        <span className="text-on-surface-variant">Risk Level</span>
                        <span className="text-yellow-500 font-bold font-mono">WARNING</span>
                      </div>
                      <div className="flex justify-between border-b border-outline-variant/50 pb-1.5">
                        <span className="text-on-surface-variant">Disk Capacity</span>
                        <span className="text-yellow-500 font-mono">82% (41 GB used)</span>
                      </div>
                      <div className="flex justify-between pb-0.5">
                        <span className="text-on-surface-variant">Device Mount</span>
                        <span className="text-on-surface font-mono">/var on /dev/sda2</span>
                      </div>
                    </div>
                    <button 
                      onClick={() => focusOnNode("hardware:disk:/var")}
                      className="mt-4 w-full py-1.5 text-xs font-semibold border border-outline-variant rounded hover:bg-surface-container-high transition-all text-on-surface"
                    >
                      Investigate in Graph
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* VIEW: GRAPH */}
          {activeView === 'graph' && (
            <div className="h-full w-full flex relative overflow-hidden">
              
              {/* Central Canvas Container */}
              <div id="net-container" className="flex-1 h-full bg-surface-container-lowest dot-pattern cursor-move relative" ref={netContainerRef}>
                <div id="net"></div>
              </div>

              {/* Zoom Controls (Bottom Left) */}
              <div className="absolute bottom-6 left-6 flex flex-col gap-2 z-20">
                <div className="bg-surface-container border border-outline-variant rounded-lg p-1 flex flex-col gap-1 shadow-xl">
                  <button 
                    onClick={() => networkRef.current && networkRef.current.zoom && networkRef.current.moveTo({ scale: networkRef.current.getScale() * 1.2 })}
                    className="p-1.5 text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest rounded transition-colors"
                  >
                    <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>add</span>
                  </button>
                  <button 
                    onClick={() => networkRef.current && networkRef.current.zoom && networkRef.current.moveTo({ scale: networkRef.current.getScale() * 0.8 })}
                    className="p-1.5 text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest rounded transition-colors"
                  >
                    <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>remove</span>
                  </button>
                  <div className="w-full h-[1px] bg-outline-variant my-1"></div>
                  <button 
                    onClick={() => networkRef.current && networkRef.current.fit({ animation: true })}
                    className="p-1.5 text-on-surface-variant hover:text-on-surface hover:bg-surface-container-highest rounded transition-colors"
                  >
                    <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>fit_screen</span>
                  </button>
                </div>
              </div>

              {/* View Toggles (Top Right Selection: Logical/Physical) */}
              <div className="absolute top-6 right-6 flex items-center gap-1.5 bg-surface-container border border-outline-variant rounded-lg p-1 z-20 shadow-md">
                <button 
                  onClick={() => setShowSettingsPopover(!showSettingsPopover)}
                  className={`px-3 py-1 text-xs font-semibold rounded flex items-center gap-1.5 transition-all ${showSettingsPopover ? 'bg-primary/25 border border-primary/40 text-primary' : 'text-on-surface-variant hover:bg-surface-container-high'}`}
                >
                  <span className="material-symbols-outlined" style={{ fontSize: '14px' }}>settings</span>
                  Graph Settings
                </button>
                <div className="w-px h-4 bg-outline-variant mx-0.5"></div>
                <button className="px-3 py-1 text-xs font-semibold rounded bg-surface-container-highest border border-outline-variant text-on-surface shadow-sm">Logical</button>
                <button className="px-3 py-1 text-xs font-semibold rounded text-on-surface-variant hover:text-on-surface transition-colors">Physical</button>
              </div>

              {/* Popover: Graph Settings sidebar overlay */}
              {showSettingsPopover && (
                <div className="absolute top-20 right-6 w-72 bg-surface/90 backdrop-blur-md border border-outline-variant rounded-xl flex flex-col z-20 shadow-2xl overflow-hidden max-h-[calc(100vh-140px)]">
                  <div className="p-4 border-b border-outline-variant flex justify-between items-center">
                    <h3 className="text-sm font-bold text-on-surface">Graph Configurations</h3>
                    <button onClick={() => setShowSettingsPopover(false)} className="text-on-surface-variant hover:text-on-surface">
                      <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>close</span>
                    </button>
                  </div>
                  <div className="p-4 overflow-y-auto flex-1 space-y-6 custom-scrollbar text-xs">
                    
                    {/* Visibility Section */}
                    <div className="space-y-3">
                      <h4 className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Visibility</h4>
                      <label className="flex items-center justify-between cursor-pointer group">
                        <span className="text-on-surface group-hover:text-primary transition-colors">Show Packages</span>
                        <div className="relative">
                          <input 
                            checked={showPackages} 
                            onChange={(e) => setShowPackages(e.target.checked)}
                            type="checkbox" 
                            className="sr-only peer"
                          />
                          <div className="w-9 h-5 bg-surface-container-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-on-surface-variant peer-checked:after:bg-on-surface after:border-gray-500 after:border after:rounded-full after:h-4 after:w-4 after:transition-all border border-outline-variant peer-checked:bg-primary"></div>
                        </div>
                      </label>
                      <label className="flex items-center justify-between cursor-pointer group">
                        <span className="text-on-surface group-hover:text-primary transition-colors">Orphaned Resources</span>
                        <div className="relative">
                          <input 
                            checked={showOrphans} 
                            onChange={(e) => setShowOrphans(e.target.checked)}
                            type="checkbox" 
                            className="sr-only peer"
                          />
                          <div className="w-9 h-5 bg-surface-container-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-on-surface-variant peer-checked:after:bg-on-surface after:border-gray-500 after:border after:rounded-full after:h-4 after:w-4 after:transition-all border border-outline-variant peer-checked:bg-primary"></div>
                        </div>
                      </label>
                    </div>

                    {/* Layers Section */}
                    <div className="space-y-3">
                      <h4 className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Layers</h4>
                      <div className="space-y-2">
                        <label className="flex items-center gap-2.5 cursor-pointer">
                          <input 
                            checked={activeLayers.foundation}
                            onChange={(e) => setActiveLayers(prev => ({ ...prev, foundation: e.target.checked }))}
                            className="rounded-sm bg-surface-container-highest border-outline-variant text-primary focus:ring-primary focus:ring-offset-surface-container h-4 w-4" 
                            type="checkbox"
                          />
                          <span className="text-on-surface hover:text-primary transition-colors">Foundation (Compute, Network)</span>
                        </label>
                        <label className="flex items-center gap-2.5 cursor-pointer">
                          <input 
                            checked={activeLayers.living}
                            onChange={(e) => setActiveLayers(prev => ({ ...prev, living: e.target.checked }))}
                            className="rounded-sm bg-surface-container-highest border-outline-variant text-primary focus:ring-primary focus:ring-offset-surface-container h-4 w-4" 
                            type="checkbox"
                          />
                          <span className="text-on-surface hover:text-primary transition-colors">Living (Processes, Containers)</span>
                        </label>
                        <label className="flex items-center gap-2.5 cursor-pointer">
                          <input 
                            checked={activeLayers.exposed}
                            onChange={(e) => setActiveLayers(prev => ({ ...prev, exposed: e.target.checked }))}
                            className="rounded-sm bg-surface-container-highest border-outline-variant text-primary focus:ring-primary focus:ring-offset-surface-container h-4 w-4" 
                            type="checkbox"
                          />
                          <span className="text-on-surface hover:text-primary transition-colors">Exposed (Routes, Certs)</span>
                        </label>
                      </div>
                    </div>

                    {/* Layout Engine Selector */}
                    <div className="space-y-3">
                      <h4 className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Layout Engine</h4>
                      <div className="grid grid-cols-3 gap-1 bg-surface-container-highest p-1 rounded-lg border border-outline-variant">
                        <button 
                          onClick={() => setLayoutMode('concentric')}
                          className={`py-1 rounded text-[10px] font-semibold transition-all ${layoutMode === 'concentric' ? 'bg-primary text-on-primary shadow-sm' : 'text-on-surface-variant hover:text-on-surface'}`}
                        >
                          Concentric
                        </button>
                        <button 
                          onClick={() => setLayoutMode('force')}
                          className={`py-1 rounded text-[10px] font-semibold transition-all ${layoutMode === 'force' ? 'bg-primary text-on-primary shadow-sm' : 'text-on-surface-variant hover:text-on-surface'}`}
                        >
                          Force
                        </button>
                        <button 
                          onClick={() => setLayoutMode('hierarchical')}
                          className={`py-1 rounded text-[10px] font-semibold transition-all ${layoutMode === 'hierarchical' ? 'bg-primary text-on-primary shadow-sm' : 'text-on-surface-variant hover:text-on-surface'}`}
                        >
                          Hierarch.
                        </button>
                      </div>
                    </div>

                    {/* Graph Physics */}
                    <div className="space-y-3">
                      <h4 className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider">Graph Physics</h4>
                      <label className="flex items-center justify-between cursor-pointer group">
                        <span className="text-on-surface group-hover:text-primary transition-colors">Enable Physics Engine</span>
                        <div className="relative">
                          <input 
                            checked={enablePhysics} 
                            onChange={(e) => setEnablePhysics(e.target.checked)}
                            type="checkbox" 
                            className="sr-only peer"
                          />
                          <div className="w-9 h-5 bg-surface-container-highest peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-on-surface-variant peer-checked:after:bg-on-surface after:border-gray-500 after:border after:rounded-full after:h-4 after:w-4 after:transition-all border border-outline-variant peer-checked:bg-primary"></div>
                        </div>
                      </label>
                    </div>
                  </div>
                </div>
              )}

              {/* Floating Bottom Context Quick Details Panel (Selected state) */}
              {selectedNode && (
                <div className="absolute bottom-6 left-1/2 -translate-x-1/2 w-[550px] bg-surface border border-outline-variant rounded-xl shadow-2xl z-20 flex overflow-hidden">
                  <div className="w-1.5 bg-primary"></div>
                  <div className="p-4 flex-1 flex items-start gap-4">
                    <div className="w-9 h-9 bg-primary/10 rounded-lg border border-primary/20 flex items-center justify-center shrink-0">
                      <span className="material-symbols-outlined text-primary text-lg">
                        {selectedNode.type === 'PROCESS' ? 'memory' : selectedNode.type === 'WEBSITE' ? 'language' : selectedNode.type === 'DATABASE' ? 'database' : selectedNode.type === 'CERTIFICATE' ? 'workspace_premium' : 'lan'}
                      </span>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex justify-between items-start">
                        <div>
                          <h3 className="text-sm font-bold text-on-surface truncate">{selectedNode.label}</h3>
                          <p className="text-[10px] text-on-surface-variant">{selectedNode.type} · ID: {selectedNode.id}</p>
                        </div>
                        <span className={`px-2 py-0.5 rounded text-[9px] font-bold ${selectedNode.health === 'critical' ? 'bg-error-container text-error' : selectedNode.health === 'warning' ? 'bg-yellow-500/10 text-yellow-500' : 'bg-tertiary-container text-on-tertiary-container'} border border-outline-variant/30 uppercase`}>
                          {selectedNode.health ? selectedNode.health : 'healthy'}
                        </span>
                      </div>
                      
                      {/* Metric cards list (CPU, memory, uptime context) */}
                      <div className="mt-3 grid grid-cols-3 gap-2 text-xs">
                        <div className="bg-surface-container-high p-2 rounded border border-outline-variant/50">
                          <p className="text-[9px] text-on-surface-variant uppercase font-bold tracking-wider">Metrics</p>
                          <p className="text-xs font-semibold text-on-surface mt-0.5">{selectedNode.type === 'PROCESS' ? `${selectedNode.metadata.cpuPercent || 0}% CPU` : selectedNode.type === 'HARDWARE' ? `${selectedNode.metadata.utilizationPercent || selectedNode.metadata.usedPercent || 0}%` : 'N/A'}</p>
                        </div>
                        <div className="bg-surface-container-high p-2 rounded border border-outline-variant/50">
                          <p className="text-[9px] text-on-surface-variant uppercase font-bold tracking-wider">Primary Val</p>
                          <p className="text-xs font-semibold text-on-surface mt-0.5 truncate">{selectedNode.type === 'CERTIFICATE' ? `${selectedNode.metadata.daysUntilExpiry}d left` : selectedNode.type === 'PROCESS' ? `${selectedNode.metadata.memoryMB || 0} MB` : selectedNode.version || 'Active'}</p>
                        </div>
                        <div className="bg-surface-container-high p-2 rounded border border-outline-variant/50">
                          <p className="text-[9px] text-on-surface-variant uppercase font-bold tracking-wider">Details</p>
                          <button 
                            onClick={() => setDetailTab('business')} 
                            className="text-[10px] text-primary hover:underline font-bold mt-0.5 block"
                          >
                            Open Drawer →
                          </button>
                        </div>
                      </div>
                    </div>
                    <button onClick={() => { setSelectedNode(null); if (networkRef.current) networkRef.current.unselectAll(); }} className="text-on-surface-variant hover:text-on-surface shrink-0">
                      <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>close</span>
                    </button>
                  </div>
                </div>
              )}

              {/* Sliding Detail Analysis Right Drawer */}
              {selectedNode && (
                <aside className="absolute top-0 right-0 h-full w-[400px] bg-surface-container-lowest border-l border-outline-variant shadow-2xl z-40 flex flex-col transform transition-transform duration-300">
                  {/* Header */}
                  <div className="px-5 py-4 border-b border-outline-variant flex justify-between items-start shrink-0">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <span className="material-symbols-outlined text-primary text-xl">
                          {selectedNode.type === 'PROCESS' ? 'memory' : selectedNode.type === 'WEBSITE' ? 'language' : selectedNode.type === 'DATABASE' ? 'database' : selectedNode.type === 'CERTIFICATE' ? 'workspace_premium' : 'lan'}
                        </span>
                        <h2 className="text-base font-bold text-on-surface truncate max-w-[280px]">{selectedNode.label}</h2>
                      </div>
                      <p className="text-[10px] font-mono text-on-surface-variant truncate">{selectedNode.id}</p>
                    </div>
                    <button 
                      onClick={() => { setSelectedNode(null); if (networkRef.current) networkRef.current.unselectAll(); }} 
                      className="text-on-surface-variant hover:text-on-surface p-1 rounded hover:bg-surface-container transition-colors"
                    >
                      <span className="material-symbols-outlined" style={{ fontSize: '18px' }}>close</span>
                    </button>
                  </div>

                  {/* Status Uptime bar */}
                  <div className="px-5 py-2.5 bg-surface-container flex items-center justify-between border-b border-outline-variant text-xs">
                    <div className="flex items-center gap-2 font-medium">
                      <div className={`w-2 h-2 rounded-full ${selectedNode.health === 'critical' ? 'bg-error animate-pulse' : selectedNode.health === 'warning' ? 'bg-yellow-500' : 'bg-tertiary'} shadow-md`}></div>
                      <span className={`${selectedNode.health === 'critical' ? 'text-error' : selectedNode.health === 'warning' ? 'text-yellow-500' : 'text-tertiary'} capitalize`}>{selectedNode.health || 'healthy'}</span>
                    </div>
                    <span className="text-on-surface-variant">{selectedNode.metadata.uptime || 'Uptime: Active'}</span>
                  </div>

                  {/* Business / Developer tab view selector */}
                  <div className="flex border-b border-outline-variant shrink-0 bg-surface-container-low text-xs">
                    <button 
                      onClick={() => setDetailTab('business')} 
                      className={`flex-1 py-2 font-semibold border-b-2 text-center transition-all ${detailTab === 'business' ? 'border-primary text-primary' : 'border-transparent text-on-surface-variant hover:text-on-surface'}`}
                    >
                      Business View
                    </button>
                    <button 
                      onClick={() => setDetailTab('developer')} 
                      className={`flex-1 py-2 font-semibold border-b-2 text-center transition-all ${detailTab === 'developer' ? 'border-primary text-primary' : 'border-transparent text-on-surface-variant hover:text-on-surface'}`}
                    >
                      Developer View
                    </button>
                  </div>

                  {/* Drawer Content */}
                  <div className="flex-1 overflow-y-auto p-5 space-y-6 custom-scrollbar text-xs">
                    {detailTab === 'developer' ? (
                      <section className="space-y-4">
                        <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold">Raw Metadata Attributes</h3>
                        <div className="border border-outline-variant rounded-lg overflow-hidden bg-surface-container-low font-mono text-[11px]">
                          <table className="w-full text-left border-collapse">
                            <tbody>
                              {Object.keys(selectedNode.metadata || {}).map((k, i) => (
                                <tr key={i} className="border-b border-outline-variant/50 last:border-0 hover:bg-surface-container/50 transition-colors">
                                  <td className="px-3 py-2 text-on-surface-variant font-semibold select-all border-r border-outline-variant/30 w-1/3">{k}</td>
                                  <td className="px-3 py-2 text-on-surface break-all w-2/3 select-all">{typeof selectedNode.metadata[k] === 'object' ? JSON.stringify(selectedNode.metadata[k]) : String(selectedNode.metadata[k])}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </section>
                    ) : (
                      <div className="space-y-6">
                        
                        {/* Custom Widgets per node type */}
                        {selectedNode.type === 'CERTIFICATE' && (
                          <section className="space-y-3">
                            <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold">Certificate Timeline</h3>
                            <div className="bg-surface-container p-4 rounded-xl border border-outline-variant space-y-3">
                              <div className="flex justify-between items-center text-[10px] font-semibold text-on-surface-variant">
                                <span>Issued</span>
                                <span className={selectedNode.health === 'critical' ? 'text-rose-500 font-bold' : 'text-amber-500'}>
                                  {selectedNode.metadata.daysUntilExpiry < 0 ? 'Expired' : `${selectedNode.metadata.daysUntilExpiry} days remaining`}
                                </span>
                                <span>Expires</span>
                              </div>
                              <div className="h-2 w-full bg-surface-container-highest rounded-full overflow-hidden">
                                <div 
                                  className={`h-full ${selectedNode.health === 'critical' ? 'bg-error' : selectedNode.health === 'warning' ? 'bg-yellow-500' : 'bg-tertiary'}`}
                                  style={{ width: `${Math.max(10, Math.min(100, (selectedNode.metadata.daysUntilExpiry / 365) * 100))}%` }}
                                ></div>
                              </div>
                              <div className="text-[10px] text-on-surface-variant space-y-1 pt-1 leading-relaxed">
                                <div><b>Issuer CA:</b> {selectedNode.metadata.issuer}</div>
                                <div><b>Subject CN:</b> {selectedNode.metadata.subject}</div>
                                <div><b>Key Strength:</b> {selectedNode.metadata.keyType} ({selectedNode.metadata.keyBits} bits)</div>
                              </div>
                            </div>
                          </section>
                        )}

                        {selectedNode.type === 'HARDWARE' && selectedNode.id === 'hardware:memory' && (
                          <section className="space-y-3">
                            <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold border-b border-outline-variant/30 pb-1">Memory capacity</h3>
                            <div className="bg-surface-container p-4 rounded-xl border border-outline-variant space-y-3">
                              <div className="flex justify-between items-center text-xs">
                                <span className="font-medium">RAM Utilization</span>
                                <span className="font-bold text-yellow-500">{selectedNode.metadata.usedPercent}%</span>
                              </div>
                              <div className="h-2 w-full bg-surface-container-highest rounded-full overflow-hidden">
                                <div className="h-full bg-primary" style={{ width: `${selectedNode.metadata.usedPercent}%` }}></div>
                              </div>
                              <div className="text-[10px] text-on-surface-variant">
                                <b>Memory footprint:</b> {formatBytes(selectedNode.metadata.usedKB * 1024)} / {formatBytes(selectedNode.metadata.totalKB * 1024)} total
                              </div>
                            </div>
                          </section>
                        )}

                        {selectedNode.type === 'HARDWARE' && selectedNode.id === 'hardware:cpu' && (
                          <section className="space-y-3">
                            <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold border-b border-outline-variant/30 pb-1">CPU capacity</h3>
                            <div className="bg-surface-container p-4 rounded-xl border border-outline-variant space-y-3">
                              <div className="flex justify-between items-center text-xs">
                                <span className="font-medium">EPYC Core Utilization</span>
                                <span className="font-bold text-primary">{selectedNode.metadata.utilizationPercent}%</span>
                              </div>
                              <div className="h-2 w-full bg-surface-container-highest rounded-full overflow-hidden">
                                <div className="h-full bg-primary" style={{ width: `${selectedNode.metadata.utilizationPercent}%` }}></div>
                              </div>
                              <div className="text-[10px] text-on-surface-variant">
                                <b>Load 1m average:</b> {selectedNode.metadata.load1} · {selectedNode.metadata.cores} Logical cores allocated
                              </div>
                            </div>
                          </section>
                        )}

                        {/* General Metadata section */}
                        <section className="space-y-3">
                          <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold border-b border-outline-variant/30 pb-1 font-headline">Metadata Overview</h3>
                          <div className="grid grid-cols-2 gap-3">
                            <div className="bg-surface-container-low p-3 rounded border border-outline-variant">
                              <span className="block text-on-surface-variant mb-1 text-[10px] uppercase tracking-wider">Engine / Distro</span>
                              <span className="text-on-surface font-semibold text-xs">{selectedNode.type === 'DATABASE' ? selectedNode.metadata.engine : selectedNode.type === 'OS' ? selectedNode.label : selectedNode.type}</span>
                            </div>
                            <div className="bg-surface-container-low p-3 rounded border border-outline-variant">
                              <span className="block text-on-surface-variant mb-1 text-[10px] uppercase tracking-wider">Port Bind</span>
                              <span className="text-on-surface font-semibold text-xs font-mono">{selectedNode.type === 'DATABASE' ? selectedNode.metadata.port : selectedNode.type === 'PORT' ? selectedNode.metadata.port : 'N/A'}</span>
                            </div>
                            <div className="bg-surface-container-low p-3 rounded border border-outline-variant col-span-2">
                              <span className="block text-on-surface-variant mb-1 text-[10px] uppercase tracking-wider">Main Path / Location</span>
                              <span className="text-on-surface font-mono text-[11px] break-all">{selectedNode.type === 'DATABASE' ? selectedNode.metadata.storagePath : selectedNode.type === 'WEBSITE' ? selectedNode.metadata.root : selectedNode.id}</span>
                            </div>
                          </div>
                        </section>

                        {/* Security Findings for this node */}
                        <section className="space-y-3">
                          <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold border-b border-outline-variant/30 pb-1">Security Findings</h3>
                          <div className="bg-surface-container rounded-lg border border-outline-variant divide-y divide-outline-variant overflow-hidden">
                            {doc.security && doc.security.some(f => f.nodeId === selectedNode.id) ? (
                              doc.security.filter(f => f.nodeId === selectedNode.id).map((f, i) => (
                                <div key={i} className="p-3 bg-surface-container-high/40 flex items-start gap-2.5">
                                  <span className={`material-symbols-outlined text-base ${f.severity === 'critical' ? 'text-error' : 'text-yellow-500'} fill-icon`}>
                                    {f.severity === 'critical' ? 'report' : 'warning'}
                                  </span>
                                  <div className="flex-1">
                                    <h4 className="font-semibold text-on-surface text-xs">{f.title}</h4>
                                    <p className="text-[10px] text-on-surface-variant mt-0.5 leading-normal">{f.detail}</p>
                                  </div>
                                </div>
                              ))
                            ) : (
                              <div className="p-3 text-center text-on-surface-variant text-[10px] flex items-center justify-center gap-1.5">
                                <span className="material-symbols-outlined text-tertiary text-sm fill-icon">check_circle</span>
                                No security vulnerabilities found on this node.
                              </div>
                            )}
                          </div>
                        </section>

                        {/* Connections section (Connected Neighbors clickable links) */}
                        <section className="space-y-3">
                          <h3 className="text-[10px] uppercase tracking-widest text-on-surface-variant font-bold border-b border-outline-variant/30 pb-1">Connected Neighbors</h3>
                          <div className="space-y-2">
                            {doc.edges.some(e => e.source === selectedNode.id || e.target === selectedNode.id) ? (
                              doc.edges.filter(e => e.source === selectedNode.id || e.target === selectedNode.id).map((e, i) => {
                                const neighborId = e.source === selectedNode.id ? e.target : e.source;
                                const neighbor = rawByIdRef.current[neighborId];
                                const label = neighbor ? neighbor.label : neighborId;
                                const relType = e.relation;
                                
                                return (
                                  <div 
                                    key={i} 
                                    onClick={() => focusOnNode(neighborId)}
                                    className="flex items-center justify-between p-2.5 bg-surface-container-low rounded border border-outline-variant hover:bg-surface-container hover:border-primary/50 transition-all cursor-pointer group"
                                  >
                                    <div className="flex items-center gap-2.5 min-w-0">
                                      <div className="w-7 h-7 rounded bg-surface-container-high border border-outline-variant/40 flex items-center justify-center shrink-0">
                                        <span className="material-symbols-outlined text-on-surface-variant text-sm">
                                          {neighbor?.type === 'PROCESS' ? 'memory' : neighbor?.type === 'WEBSITE' ? 'language' : neighbor?.type === 'DATABASE' ? 'database' : 'lan'}
                                        </span>
                                      </div>
                                      <div className="min-w-0">
                                        <div className="text-xs font-semibold text-on-surface truncate group-hover:text-primary transition-colors">{label}</div>
                                        <div className="text-[9px] text-primary font-bold uppercase tracking-wider font-mono">{relType}</div>
                                      </div>
                                    </div>
                                    <span className="material-symbols-outlined text-on-surface-variant opacity-0 group-hover:opacity-100 transition-opacity text-sm">chevron_right</span>
                                  </div>
                                );
                              })
                            ) : (
                              <div className="text-center text-on-surface-variant text-[10px] py-2">No connected neighbor resources.</div>
                            )}
                          </div>
                        </section>
                      </div>
                    )}
                  </div>

                  {/* Actions Drawer Footer */}
                  <div className="p-4 border-t border-outline-variant flex gap-2 shrink-0 bg-surface-container-low">
                    <button className="flex-1 py-1.5 border border-outline-variant hover:bg-surface-container-high text-xs font-semibold rounded text-on-surface transition-all">
                      View Logs
                    </button>
                    <button className="flex-1 py-1.5 border border-outline-variant hover:bg-surface-container-high text-xs font-semibold rounded text-on-surface transition-all">
                      Connect CLI
                    </button>
                  </div>
                </aside>
              )}
            </div>
          )}

          {/* VIEW: SECURITY FINDINGS & DRIFT */}
          {activeView === 'security' && (
            <div className="h-full w-full flex overflow-hidden">
              <main className="flex-1 flex flex-col lg:flex-row overflow-hidden bg-background">
                
                {/* Left Pane: Security findings list */}
                <section className="w-full lg:w-1/2 flex flex-col border-r border-outline-variant bg-surface-container-lowest">
                  <div className="px-6 py-4 border-b border-outline-variant bg-surface-dim flex justify-between items-center shrink-0">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-on-surface flex items-center gap-2">
                      <span className="material-symbols-outlined text-primary text-sm">gavel</span>
                      Security Audit Findings
                    </h3>
                    <div className="flex gap-1.5 text-[10px] bg-surface-container border border-outline-variant p-0.5 rounded">
                      <button className="px-2.5 py-0.5 rounded bg-surface-container-high text-on-surface border border-outline-variant font-medium">Critical</button>
                      <button className="px-2.5 py-0.5 rounded text-on-surface-variant hover:text-on-surface font-medium">All</button>
                    </div>
                  </div>
                  
                  <div className="flex-1 overflow-y-auto p-5 space-y-4 custom-scrollbar">
                    {/* Finding Card 1 */}
                    <div 
                      onClick={() => focusOnNode("cert:sha256:infrasight.io")}
                      className="bg-surface-container border border-outline-variant rounded-lg p-4 hover:border-error/50 transition-colors cursor-pointer group space-y-2 relative"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          <span className="w-2 h-2 rounded-full bg-error shadow-[0_0_8px_rgba(239,68,68,0.6)]"></span>
                          <h4 className="font-bold text-on-surface text-xs">Weak RSA Encryption Key Length</h4>
                        </div>
                        <span className="px-2 py-0.5 text-[8px] uppercase font-bold tracking-wider rounded bg-error/10 text-error border border-error/20">Critical</span>
                      </div>
                      <p className="text-xs text-on-surface-variant leading-relaxed">
                        The SSL/TLS certificate for host wildcard <code className="font-mono text-[10px] bg-surface-container-highest px-1 py-0.5 rounded text-on-surface">*.infrasight.io</code> uses a 1024-bit RSA key length. This key size is cryptographically weak and violates standard CIS security benchmarks (&lt; 2048).
                      </p>
                      <div className="flex items-center justify-between text-[10px] text-on-surface-variant pt-1 border-t border-outline-variant/30">
                        <span className="flex items-center gap-1"><span className="material-symbols-outlined text-xs">policy</span> CIS TLS 1.22</span>
                        <button className="text-primary hover:underline transition-all opacity-0 group-hover:opacity-100 flex items-center gap-0.5 font-semibold">
                          Remediate <span className="material-symbols-outlined text-xs">arrow_forward</span>
                        </button>
                      </div>
                    </div>

                    {/* Finding Card 2 */}
                    <div 
                      onClick={() => focusOnNode("cert:sha256:infrasight.io")}
                      className="bg-surface-container border border-outline-variant rounded-lg p-4 hover:border-error/50 transition-colors cursor-pointer group space-y-2"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          <span className="w-2 h-2 rounded-full bg-error shadow-[0_0_8px_rgba(239,68,68,0.6)]"></span>
                          <h4 className="font-bold text-on-surface text-xs">TLS Certificate Expiring In 2 Days</h4>
                        </div>
                        <span className="px-2 py-0.5 text-[8px] uppercase font-bold tracking-wider rounded bg-error/10 text-error border border-error/20">Critical</span>
                      </div>
                      <p className="text-xs text-on-surface-variant leading-relaxed">
                        TLS certificate wrapping domain router <code className="font-mono text-[10px] bg-surface-container-highest px-1 py-0.5 rounded text-on-surface">api.infrasight.io</code> will expire in 2 days. An expired certificate causes request failures.
                      </p>
                      <div className="flex items-center justify-between text-[10px] text-on-surface-variant pt-1 border-t border-outline-variant/30">
                        <span className="flex items-center gap-1"><span className="material-symbols-outlined text-xs">policy</span> CIS TLS 2.1.5</span>
                        <button className="text-primary hover:underline transition-all opacity-0 group-hover:opacity-100 flex items-center gap-0.5 font-semibold">
                          Remediate <span className="material-symbols-outlined text-xs">arrow_forward</span>
                        </button>
                      </div>
                    </div>

                    {/* Finding Card 3 */}
                    <div 
                      onClick={() => focusOnNode("port:tcp:443")}
                      className="bg-surface-container border border-outline-variant rounded-lg p-4 hover:border-yellow-500/50 transition-colors cursor-pointer group space-y-2"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-2">
                          <span className="w-2 h-2 rounded-full bg-yellow-500"></span>
                          <h4 className="font-bold text-on-surface text-xs">Port Reachable from All Interfaces</h4>
                        </div>
                        <span className="px-2 py-0.5 text-[8px] uppercase font-bold tracking-wider rounded bg-yellow-500/10 text-yellow-500 border border-yellow-500/20">Warning</span>
                      </div>
                      <p className="text-xs text-on-surface-variant leading-relaxed">
                        The HTTPS ingress listener on socket interface <code className="font-mono text-[10px] bg-surface-container-highest px-1 py-0.5 rounded text-on-surface">:443</code> is bound to wildcard interface <code class="font-mono text-[10px] bg-surface-container-highest px-1 py-0.5 rounded text-on-surface">0.0.0.0</code>.
                      </p>
                      <div className="flex items-center justify-between text-[10px] text-on-surface-variant pt-1 border-t border-outline-variant/30">
                        <span className="flex items-center gap-1"><span className="material-symbols-outlined text-xs">policy</span> CIS Net 4.1</span>
                        <button className="text-primary hover:underline transition-all opacity-0 group-hover:opacity-100 flex items-center gap-0.5 font-semibold">
                          Remediate <span className="material-symbols-outlined text-xs">arrow_forward</span>
                        </button>
                      </div>
                    </div>
                  </div>
                </section>

                {/* Right Pane: Drift Analysis code-diff view */}
                <section className="w-full lg:w-1/2 flex flex-col bg-surface-container-lowest">
                  <div className="px-6 py-4 border-b border-outline-variant bg-surface-dim flex justify-between items-center shrink-0">
                    <h3 className="text-xs font-bold uppercase tracking-wider text-on-surface flex items-center gap-2">
                      <span className="material-symbols-outlined text-primary text-sm">difference</span>
                      Configuration Drift Analysis
                    </h3>
                    <div className="flex gap-3 text-[10px] font-mono">
                      <span className="flex items-center gap-1 text-tertiary"><span className="w-1.5 h-1.5 bg-tertiary rounded-full"></span> +1 Added</span>
                      <span className="flex items-center gap-1 text-error"><span className="w-1.5 h-1.5 bg-error rounded-full"></span> -2 Removed</span>
                      <span className="flex items-center gap-1 text-yellow-500"><span className="w-1.5 h-1.5 bg-yellow-500 rounded-full"></span> ~1 Modified</span>
                    </div>
                  </div>

                  <div className="flex-1 overflow-y-auto p-5 bg-[#0c0c0f] custom-scrollbar">
                    <div className="font-mono text-[11px] leading-relaxed border border-outline-variant rounded-lg overflow-hidden bg-surface-container-low">
                      
                      {/* Diff Block 1: nginx config */}
                      <div className="border-b border-outline-variant">
                        <div className="px-4 py-2 bg-surface-container-high text-on-surface-variant text-[10px] border-b border-outline-variant flex justify-between items-center">
                          <span>/etc/nginx/sites-enabled/api.conf</span>
                          <span className="px-2 py-0.5 rounded bg-yellow-500/10 text-yellow-500 border border-yellow-500/20 font-bold uppercase text-[8px]">Modified</span>
                        </div>
                        <div className="p-4 overflow-x-auto whitespace-pre font-mono">
                          <div className="text-on-surface-variant">  server &#123;</div>
                          <div className="text-on-surface-variant">    listen 443 ssl;</div>
                          <div className="text-on-surface-variant">    server_name api.infrasight.io;</div>
                          <div className="diff-removed w-full block">-   ssl_protocols TLSv1.2 TLSv1.3;</div>
                          <div className="diff-removed w-full block">-   ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256;</div>
                          <div className="diff-added w-full block">+   ssl_protocols TLSv1 TLSv1.1 TLSv1.2; # Weak protocol added</div>
                          <div className="diff-added w-full block">+   ssl_ciphers ALL:!aNULL:!eNULL;      # Insecure cipher list</div>
                          <div className="text-on-surface-variant">    ssl_certificate /etc/nginx/certs/infrasight.crt;</div>
                          <div className="text-on-surface-variant">  &#125;</div>
                        </div>
                      </div>

                      {/* Diff Block 2: TLS Key file path */}
                      <div>
                        <div className="px-4 py-2 bg-surface-container-high text-on-surface-variant text-[10px] border-b border-outline-variant flex justify-between items-center">
                          <span>/etc/nginx/certs/infrasight.key</span>
                          <span className="px-2 py-0.5 rounded bg-tertiary/10 text-tertiary border border-tertiary/20 font-bold uppercase text-[8px]">Added</span>
                        </div>
                        <div className="p-4 overflow-x-auto whitespace-pre font-mono">
                          <div className="text-on-surface-variant">  # TLS Certificate Private Key block</div>
                          <div className="diff-added w-full block">+   RSA KEY GENERATED: 1024 bits</div>
                          <div className="diff-added w-full block">+   Issuer: Let's Encrypt Authority</div>
                          <div className="diff-added w-full block">+   Created at: 2026-06-30T10:00:00Z</div>
                        </div>
                      </div>

                    </div>
                  </div>
                </section>
              </main>
            </div>
          )}

          {/* VIEW: SYSTEM PACKAGES CATALOG */}
          {activeView === 'packages' && (
            <div className="h-full w-full bg-background flex flex-col overflow-hidden">
              <div className="flex-1 overflow-y-auto custom-scrollbar p-6 space-y-6">
                
                {/* Page Header */}
                <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-4">
                  <div>
                    <h2 className="text-xl font-headline font-bold tracking-tight">System Packages Catalog</h2>
                    <p className="text-on-surface-variant text-xs mt-1">Found <span className="text-on-surface font-semibold">5</span> installed modules on host <span className="font-mono text-[10px] bg-surface-container px-1.5 py-0.5 rounded border border-outline-variant">{doc.scan.hostname}</span></p>
                  </div>
                  <div className="flex flex-wrap items-center gap-3 w-full md:w-auto text-xs">
                    
                    {/* Filter input */}
                    <div className="relative w-full md:w-56">
                      <span className="material-symbols-outlined absolute left-2.5 top-1/2 -translate-y-1/2 text-on-surface-variant" style={{ fontSize: '15px' }}>filter_list</span>
                      <input 
                        value={pkgSearch}
                        onChange={(e) => setPkgSearch(e.target.value)}
                        className="w-full bg-surface-container border border-outline-variant text-on-surface placeholder:text-on-surface-variant rounded py-1.5 pl-8 pr-3 focus:outline-none focus:ring-1 focus:ring-primary focus:border-transparent transition-all" 
                        placeholder="Filter packages..." 
                        type="text"
                      />
                    </div>
                    
                    {/* Source manager select */}
                    <select 
                      value={pkgSourceFilter}
                      onChange={(e) => setPkgSourceFilter(e.target.value)}
                      className="bg-surface-container border border-outline-variant text-on-surface rounded py-1.5 px-3 focus:outline-none focus:ring-1 focus:ring-primary"
                    >
                      <option value="all">Source: All</option>
                      <option value="apt">APT (dpkg)</option>
                      <option value="pip">PIP</option>
                      <option value="npm">NPM</option>
                    </select>

                    <button className="flex items-center gap-1.5 py-1.5 px-3 bg-surface border border-outline-variant hover:bg-surface-container text-on-surface rounded transition-colors">
                      <span className="material-symbols-outlined text-xs">download</span>
                      Export CSV
                    </button>
                  </div>
                </div>

                {/* Packages Datatable */}
                <div className="w-full border border-outline-variant rounded-xl overflow-hidden bg-surface">
                  <div className="overflow-x-auto custom-scrollbar">
                    <table className="w-full text-left text-xs whitespace-nowrap">
                      <thead className="bg-surface-container-low border-b border-outline-variant text-on-surface-variant font-semibold">
                        <tr>
                          <th className="px-4 py-3 w-10 text-center">
                            <input className="w-3.5 h-3.5 rounded bg-surface-container border-outline-variant text-primary focus:ring-primary focus:ring-offset-background" type="checkbox" />
                          </th>
                          <th className="px-4 py-3">Package Name</th>
                          <th className="px-4 py-3">Installed Version</th>
                          <th className="px-4 py-3">Source Manager</th>
                          <th className="px-4 py-3">Status</th>
                          <th className="px-4 py-3 text-right">Action</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-outline-variant">
                        {doc.nodes
                          .filter(n => n.type === 'PACKAGE')
                          .filter(n => !pkgSearch || n.label.toLowerCase().includes(pkgSearch.toLowerCase()))
                          .filter(n => pkgSourceFilter === 'all' || n.metadata.manager === pkgSourceFilter)
                          .map((p, i) => (
                            <tr key={i} className="hover:bg-surface-container/50 transition-colors group">
                              <td className="px-4 py-3 text-center">
                                <input className="w-3.5 h-3.5 rounded bg-surface-container border-outline-variant text-primary focus:ring-primary focus:ring-offset-background opacity-0 group-hover:opacity-100 transition-all checked:opacity-100" type="checkbox" />
                              </td>
                              <td 
                                onClick={() => focusOnNode(p.id)}
                                className="px-4 py-3 font-semibold text-primary hover:underline cursor-pointer"
                              >
                                {p.label}
                              </td>
                              <td className="px-4 py-3 font-mono text-[11px] text-on-surface-variant">{p.metadata.version || '1.0.0'}</td>
                              <td className="px-4 py-3">
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-[9px] font-bold uppercase bg-surface-container-high border border-outline-variant text-on-surface">
                                  {p.metadata.manager || 'dpkg'}
                                </span>
                              </td>
                              <td className="px-4 py-3">
                                <div className="flex items-center gap-1.5 text-tertiary">
                                  <div className="w-1.5 h-1.5 rounded-full bg-tertiary"></div>
                                  <span className="text-[10px]">Up to date</span>
                                </div>
                              </td>
                              <td className="px-4 py-3 text-right">
                                <button onClick={() => focusOnNode(p.id)} className="text-on-surface-variant hover:text-primary transition-colors p-1 rounded">
                                  <span className="material-symbols-outlined text-base">arrow_forward</span>
                                </button>
                              </td>
                            </tr>
                          ))
                        }
                      </tbody>
                    </table>
                  </div>
                </div>

              </div>
            </div>
          )}

          {/* VIEW: SETTINGS */}
          {activeView === 'settings' && (
            <div className="h-full overflow-y-auto custom-scrollbar p-6 bg-background space-y-6">
              <div>
                <h2 className="text-xl font-headline font-bold tracking-tight">System Settings</h2>
                <p className="text-on-surface-variant text-xs mt-1">Configure InfraSight scanner, module behaviors, and reporting targets.</p>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 text-xs">
                
                {/* Form Col 1: Scan Scheduler */}
                <div className="bg-surface border border-outline-variant rounded-xl p-5 space-y-4">
                  <h3 className="text-xs uppercase tracking-wider font-bold text-primary flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-sm">schedule</span>
                    Scan Scheduler
                  </h3>
                  <div className="space-y-3">
                    <div>
                      <label className="block text-on-surface-variant mb-1 font-medium">Automatic Scan Frequency</label>
                      <select className="w-full bg-surface-container border border-outline-variant text-on-surface rounded py-2 px-3 focus:outline-none focus:ring-1 focus:ring-primary">
                        <option value="none">Disabled (Manual scans only)</option>
                        <option value="hourly">Every Hour</option>
                        <option value="daily">Daily at midnight</option>
                        <option value="weekly">Weekly on Sundays</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-on-surface-variant mb-1 font-medium">Scan History Retention</label>
                      <select className="w-full bg-surface-container border border-outline-variant text-on-surface rounded py-2 px-3 focus:outline-none focus:ring-1 focus:ring-primary">
                        <option value="7">Keep last 7 scans</option>
                        <option value="30">Keep last 30 scans</option>
                        <option value="90">Keep last 90 scans</option>
                      </select>
                    </div>
                  </div>
                </div>

                {/* Form Col 2: Modules configurations */}
                <div className="bg-surface border border-outline-variant rounded-xl p-5 space-y-4">
                  <h3 className="text-xs uppercase tracking-wider font-bold text-primary flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-sm">tune</span>
                    Scan Domain Modules
                  </h3>
                  <div className="space-y-2">
                    {['hardware', 'os', 'network', 'services', 'packages', 'web', 'database'].map((m, i) => (
                      <label key={i} className="flex items-center gap-3 cursor-pointer p-1.5 hover:bg-surface-container/50 rounded transition-colors">
                        <input className="rounded bg-surface-container border-outline-variant text-primary focus:ring-primary focus:ring-offset-surface-container h-4 w-4" defaultChecked type="checkbox" />
                        <span className="text-on-surface capitalize font-medium">{m} module probes</span>
                      </label>
                    ))}
                  </div>
                </div>

                {/* Form Col 3: Terminal Installer Script */}
                <div className="bg-surface border border-outline-variant rounded-xl p-5 space-y-4">
                  <h3 className="text-xs uppercase tracking-wider font-bold text-primary flex items-center gap-1.5">
                    <span className="material-symbols-outlined text-sm">code</span>
                    Agent Installation
                  </h3>
                  <p className="text-on-surface-variant text-[11px] leading-relaxed">
                    Deploy the InfraSight lightweight scanner binary as a read-only daemon agent. Copy and run the curl script inside your Linux terminal:
                  </p>
                  <div className="bg-[#050507] border border-outline-variant p-3 rounded-lg font-mono text-[10px] text-primary break-all leading-normal select-all relative group cursor-pointer" title="Click to select all">
                    curl -fsSL https://infrasight.dev/install.sh | sh -s -- --token=eyJhbGciOiJIUzI1NiIsInR5cCI6
                    <span className="absolute right-2 top-2 material-symbols-outlined opacity-0 group-hover:opacity-100 transition-opacity text-xs text-on-surface-variant">content_copy</span>
                  </div>
                  <div className="text-[10px] text-on-surface-variant flex items-center gap-1.5 pt-1">
                    <span className="material-symbols-outlined text-xs text-tertiary">check_circle</span>
                    Standard systemd service unit supported.
                  </div>
                </div>

              </div>
            </div>
          )}

        </main>
      </div>
    </div>
  );
}
