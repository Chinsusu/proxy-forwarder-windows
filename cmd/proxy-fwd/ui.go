package main

const indexHTML = `<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Proxy Forward Grid</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <script src="https://cdn.tailwindcss.com"></script>
  <style>
    body{background:#f8f9fa;font-family:'Inter',system-ui,-apple-system,sans-serif}
    input,textarea,select,button{font-family:'Inter',system-ui,-apple-system,sans-serif}
    .sidebar{background:linear-gradient(180deg,#2d1b4e 0%,#1a0f2e 100%)}
    .sidebar-item{transition:all .2s;border-radius:.5rem;margin:.25rem 0}
    .sidebar-item:hover{background:rgba(255,255,255,.1)}
    .sidebar-item.active{background:rgba(147,51,234,.3);border-left:3px solid #a855f7}
    .table-header{background:#2d1b4e;color:#fff}
    .status-active{background:#d1fae5;color:#065f46;padding:.25rem .75rem;border-radius:9999px;font-size:.75rem;font-weight:600}
    .status-inactive{background:#fee2e2;color:#991b1b;padding:.25rem .75rem;border-radius:9999px;font-size:.75rem;font-weight:600}
    .action-btn{padding:.4rem;border-radius:.375rem;transition:all .2s;font-size:.75rem}
    .action-btn:hover{background:#f3f4f6}
    .avatar{width:2rem;height:2rem;border-radius:9999px}
    .tab-content{display:none}
    .tab-content.active{display:block}
  </style>
</head>
<body class="flex h-screen">
  <aside class="sidebar w-64 p-4 flex flex-col">
    <div class="flex items-center gap-2 mb-8">
      <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500 to-purple-600 grid place-items-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="currentColor"><path d="M12 1a5 5 0 0 0-5 5v2H6a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V10a2 2 0 0 0-2-2h-1V6a5 5 0 0 0-5-5m-3 7V6a3 3 0 1 1 6 0v2z"/></svg>
      </div>
      <div class="text-white font-bold text-lg">Proxy Forward</div>
    </div>
    <nav class="flex-1 space-y-1">
      <a href="#" onclick="switchTab('proxies'); return false;" id="tab-proxies" class="sidebar-item active flex items-center gap-3 px-3 py-2 text-white">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2M9 17H7v-7h2v7m4 0h-2V7h2v10m4 0h-2v-4h2v4z"/></svg>
        <span>Proxies</span>
      </a>
      <a href="#" onclick="switchTab('pool'); return false;" id="tab-pool" class="sidebar-item flex items-center gap-3 px-3 py-2 text-white">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M12 3L1 9l4 2.18v6L12 21l7-3.82v-6l2-1.09V17h2V9L12 3zm6.82 6L12 12.72L5.18 9L12 5.28L18.82 9zM17 15.99l-5 2.73l-5-2.73v-3.72L12 15l5-2.73v3.72z"/></svg>
        <span>📦 Pool</span>
      </a>
      <a href="#" onclick="switchTab('order'); return false;" id="tab-order" class="sidebar-item flex items-center gap-3 px-3 py-2 text-white">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M19.35 10.04A7.49 7.49 0 0 0 12 4C9.11 4 6.6 5.64 5.35 8.04A5.994 5.994 0 0 0 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5c0-2.64-2.05-4.78-4.65-4.96z"/></svg>
        <span>☁️ Order</span>
      </a>
    </nav>
    <div class="pt-4 border-t border-slate-700">
      <div class="p-2 bg-slate-800/50 rounded-lg mb-2">
        <label class="text-xs text-slate-400 block mb-1">Admin Token</label>
        <input id="tokenInput" type="password" placeholder="Enter token..." class="w-full px-2 py-1 text-xs bg-slate-900/70 border border-slate-700 rounded text-white">
      </div>
      <div class="flex items-center gap-3 px-3 py-2">
        <img src="https://ui-avatars.com/api/?name=Admin&background=8b5cf6&color=fff" class="avatar">
        <div class="flex-1">
          <div class="text-white text-sm font-medium">Admin</div>
          <div class="text-slate-400 text-xs">127.0.0.1 Only</div>
        </div>
      </div>
    </div>
  </aside>
  <main class="flex-1 overflow-auto">
    <header class="bg-white border-b px-6 py-4 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-800">Proxies Management</h1>
        <p class="text-sm text-gray-500">Proxy Forward Dashboard</p>
      </div>
      <div class="flex items-center gap-2">
        <button onclick="handleExport()" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
          <span>📦 Export Local</span>
        </button>
        <button onclick="openBulkModal()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">
          <span>📥 Bulk Add</span>
        </button>
      </div>
    </header>
    
    <div class="m-6 space-y-4">
      <!-- Proxies Tab -->
      <div id="content-proxies" class="tab-content active">
      <div class="bg-white p-4 rounded-xl shadow-sm">
        <h3 class="font-bold mb-3">Add Single Proxy</h3>
        <div class="flex gap-2">
          <input id="singleProxyInput" type="text" placeholder="ip:port:user:pass hoặc ip:port" class="flex-1 px-3 py-2 border rounded-lg">
          <button onclick="handleAddSingle()" class="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700">
            ➕ Add Proxy
          </button>
        </div>
      </div>

      <div class="bg-white p-4 rounded-xl shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-bold">Proxy List</h3>
          <div class="flex items-center gap-2">
            <label class="flex items-center gap-2">
              <input id="autoRefreshCheck" type="checkbox" class="accent-purple-600" onchange="toggleAutoRefresh()">
              <span class="text-sm">Auto 5s</span>
            </label>
            <input id="searchInput" type="text" placeholder="Search..." class="px-3 py-2 border rounded-lg">
            <button onclick="handleSearch()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">
              🔍 Search
            </button>
          </div>
        </div>
        <div class="overflow-hidden rounded-lg border">
          <table class="min-w-full">
            <thead class="table-header">
              <tr>
                <th class="py-3 px-4 text-left">#</th>
                <th class="py-3 px-4 text-left">Proxy Address</th>
                <th class="py-3 px-4 text-left">Local Port</th>
                <th class="py-3 px-4 text-left">Status</th>
                <th class="py-3 px-4 text-left">Exit IP</th>
                <th class="py-3 px-4 text-left">Action</th>
              </tr>
            </thead>
            <tbody id="rows" class="divide-y divide-gray-200"></tbody>
          </table>
        </div>
        <div class="mt-3 flex items-center justify-between">
          <div id="countText" class="text-sm text-gray-600">Loading...</div>
          <div class="text-xs text-gray-500">Local ports start at 127.0.0.1:10001+</div>
        </div>
      </div>
      </div>
      <!-- End Proxies Tab -->

      <!-- Order Tab -->
      <div id="content-order" class="tab-content">
      <div class="bg-gradient-to-r from-purple-50 to-pink-50 p-4 rounded-xl shadow-sm border-2 border-purple-200">
        <h3 class="font-bold mb-3 text-purple-800">☁️ CloudMini Order</h3>
        <div class="grid grid-cols-2 gap-2 mb-2">
          <input id="cloudminiToken" type="password" placeholder="CloudMini API Token" class="px-3 py-2 border rounded-lg">
          <button onclick="loadCloudminiRegions()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">
            🔄 Load Regions
          </button>
        </div>
        <div class="grid grid-cols-2 gap-2 mb-2">
          <select id="cloudminiType" class="px-3 py-2 border rounded-lg">
            <option value="proxy-res">Residential (proxy-res)</option>
            <option value="proxy-isp">ISP (proxy-isp)</option>
          </select>
          <select id="cloudminiRegion" class="px-3 py-2 border rounded-lg">
            <option value="">-- Chọn region --</option>
          </select>
        </div>
        <div class="mb-2">
          <input id="cloudminiQuantity" type="number" placeholder="Số lượng proxy" value="1" min="1" class="w-full px-3 py-2 border rounded-lg">
        </div>
        <div class="flex gap-2">
          <button onclick="handleCloudMiniOrder()" class="flex-1 px-6 py-2 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-lg hover:from-purple-700 hover:to-pink-700 font-semibold">
            ⚡ Order Now
          </button>
          <label class="flex items-center gap-2 px-4 py-2 border-2 border-purple-300 rounded-lg bg-white">
            <input id="cloudminiAutoStart" type="checkbox">
            <span class="text-sm font-medium">Auto Start</span>
          </label>
        </div>
        <div class="mt-2 text-xs text-purple-600">
          💡 Proxies will be added to pool. Check "Auto Start" to start immediately.
        </div>
      </div>
      </div>
      <!-- End Order Tab -->

      <!-- Pool Tab -->
      <div id="content-pool" class="tab-content">
      <div class="bg-gradient-to-r from-blue-50 to-cyan-50 p-6 rounded-xl shadow-sm border-2 border-blue-200 mb-4">
        <h2 class="text-xl font-bold text-blue-800 mb-2">📦 Proxy Pool</h2>
        <p class="text-blue-600 text-sm">Proxies in pool are stopped and not assigned ports. Click Start to activate.</p>
      </div>

      <div class="bg-white p-4 rounded-xl shadow-sm mb-4">
        <h3 class="font-bold mb-3">☁️ CloudMini Sync</h3>
        <div class="flex gap-2">
          <input id="cloudminiSyncToken" type="password" placeholder="CloudMini API Token" class="flex-1 px-3 py-2 border rounded-lg">
          <button onclick="handleCloudMiniSyncToPool()" class="px-6 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">
            ☁️ Sync All Proxy-Res
          </button>
        </div>
        <p class="text-xs text-gray-500 mt-2">💡 Sync all existing proxy-res from CloudMini to pool (not auto-started)</p>
      </div>

      <div class="bg-white p-4 rounded-xl shadow-sm mb-4">
        <h3 class="font-bold mb-3">🔄 Sync from API</h3>
        <div class="flex gap-2">
          <input id="apiUrlInput" type="text" placeholder="API URL trả về danh sách proxy (text hoặc JSON array)" class="flex-1 px-3 py-2 border rounded-lg">
          <button onclick="handleSyncAPI()" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            🔄 Sync API
          </button>
        </div>
        <p class="text-xs text-gray-500 mt-2">💡 Synced proxies will be auto-started</p>
      </div>

      <div class="bg-white p-4 rounded-xl shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-bold">Pool Proxies (Stopped)</h3>
          <div class="flex items-center gap-2">
            <button onclick="startAllPool()" class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700">
              ▶ Start All
            </button>
            <button onclick="clearPool()" class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">
              🗑 Clear Pool
            </button>
          </div>
        </div>
        <div class="overflow-hidden rounded-lg border">
          <table class="min-w-full">
            <thead class="table-header">
              <tr>
                <th class="py-3 px-4 text-left">#</th>
                <th class="py-3 px-4 text-left">Proxy Address</th>
                <th class="py-3 px-4 text-left">Status</th>
                <th class="py-3 px-4 text-left">Action</th>
              </tr>
            </thead>
            <tbody id="poolRows" class="divide-y divide-gray-200"></tbody>
          </table>
        </div>
        <div class="mt-3 flex items-center justify-between">
          <div id="poolCountText" class="text-sm text-gray-600">Loading...</div>
          <div class="text-xs text-gray-500">Pool proxies will get ports when started</div>
        </div>
      </div>
      </div>
      <!-- End Pool Tab -->
    </div>
  </main>

  <div id="bulkModal" class="hidden fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
    <div class="bg-white rounded-xl shadow-xl max-w-2xl w-full">
      <div class="p-4 border-b flex items-center justify-between">
        <h3 class="font-bold text-lg">Bulk Add Proxies</h3>
        <button onclick="closeBulkModal()" class="text-gray-500 hover:text-gray-700">✖</button>
      </div>
      <div class="p-4">
        <p class="text-sm text-gray-600 mb-2">Mỗi dòng 1 proxy: <code class="bg-gray-100 px-2 py-1 rounded">ip:port:user:pass</code> hoặc <code class="bg-gray-100 px-2 py-1 rounded">ip:port</code></p>
        <textarea id="bulkTextarea" class="w-full h-64 px-3 py-2 border rounded-lg font-mono text-sm" placeholder="1.2.3.4:8080:user:pass
5.6.7.8:3128
9.10.11.12:1080:admin:secret"></textarea>
      </div>
      <div class="p-4 border-t flex justify-end gap-2">
        <button onclick="closeBulkModal()" class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
        <button onclick="handleBulkAdd()" class="px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700">Add All</button>
      </div>
    </div>
  </div>

  <div id="toast" class="hidden fixed bottom-4 right-4 bg-gray-900 text-white px-4 py-3 rounded-lg shadow-lg"></div>

  <script>
    var rowsEl = document.getElementById('rows');
    var countEl = document.getElementById('countText');
    var toastEl = document.getElementById('toast');
    var tokenInput = document.getElementById('tokenInput');
    var DATA = { items: [] };
    var autoRefreshTimer = null;

    tokenInput.value = localStorage.getItem('admintoken') || '';
    tokenInput.addEventListener('change', function(){
      localStorage.setItem('admintoken', tokenInput.value.trim());
      showToast('Token saved');
    });

    // Load CloudMini Token from localStorage
    var cloudminiTokenInput = document.getElementById('cloudminiToken');
    if(cloudminiTokenInput){
      cloudminiTokenInput.value = localStorage.getItem('cloudmini_token') || '';
      cloudminiTokenInput.addEventListener('change', function(){
        localStorage.setItem('cloudmini_token', cloudminiTokenInput.value.trim());
        showToast('CloudMini Token saved');
      });
    }

    // Load CloudMini Sync Token from localStorage
    var cloudminiSyncTokenInput = document.getElementById('cloudminiSyncToken');
    if(cloudminiSyncTokenInput){
      cloudminiSyncTokenInput.value = localStorage.getItem('cloudmini_token') || '';
      cloudminiSyncTokenInput.addEventListener('change', function(){
        localStorage.setItem('cloudmini_token', cloudminiSyncTokenInput.value.trim());
        showToast('CloudMini Token saved');
      });
    }

    function hdr(){ 
      var t = localStorage.getItem('admintoken') || ''; 
      var h = {}; 
      if(t){ h['X-Admin-Token'] = t; } 
      return h; 
    }
    
    function GET(u){ 
      return fetch(u, {headers: hdr()}).then(function(r){ 
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); }); 
        return r.json(); 
      }); 
    }

    function POST(u, body){ 
      var opt = {method:'POST', headers: hdr()}; 
      if(body) opt.body = body;
      return fetch(u, opt).then(function(r){ 
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); }); 
        return r.text().then(function(t){ try{ return JSON.parse(t); }catch(e){ return {}; } }); 
      }); 
    }

    function showToast(msg){
      toastEl.textContent = msg;
      toastEl.classList.remove('hidden');
      setTimeout(function(){ toastEl.classList.add('hidden'); }, 2000);
    }

    function switchTab(tabName){
      // Hide all tabs
      document.querySelectorAll('.tab-content').forEach(function(el){ el.classList.remove('active'); });
      // Remove active from all sidebar items
      document.querySelectorAll('.sidebar-item').forEach(function(el){ el.classList.remove('active'); });
      // Show selected tab
      document.getElementById('content-' + tabName).classList.add('active');
      // Activate sidebar item
      document.getElementById('tab-' + tabName).classList.add('active');
    }

    function getInitial(name){ 
      return name ? name.charAt(0).toUpperCase() : 'P'; 
    }
    
    function getRandomColor(){ 
      var colors = ['ef4444','f59e0b','10b981','3b82f6','8b5cf6','ec4899']; 
      return colors[Math.floor(Math.random() * colors.length)]; 
    }

    function rowItem(it, idx){
      var tr = document.createElement('tr');
      tr.className = 'hover:bg-gray-50';
      
      var tdIdx = document.createElement('td'); 
      tdIdx.className = 'py-3 px-4 text-gray-600'; 
      tdIdx.textContent = String(idx+1); 
      tr.appendChild(tdIdx);

      // Display only hostname:port (no credentials, no ID)
      var proxyAddr = it.host + ':' + it.port;
      var color = getRandomColor();
      var initial = getInitial(it.host);
      
      var tdUp = document.createElement('td'); 
      tdUp.className = 'py-3 px-4';
      var divFlex = document.createElement('div'); 
      divFlex.className = 'flex items-center gap-3';
      var avatar = document.createElement('div');
      avatar.className = 'w-8 h-8 rounded-full grid place-items-center text-white font-bold text-sm';
      avatar.style.background = '#' + color;
      avatar.textContent = initial;
      var divText = document.createElement('div');
      var divMain = document.createElement('div');
      divMain.className = 'font-medium text-gray-800'; 
      divMain.textContent = proxyAddr;
      divText.appendChild(divMain);
      divFlex.appendChild(avatar); 
      divFlex.appendChild(divText);
      tdUp.appendChild(divFlex); 
      tr.appendChild(tdUp);

      var local = '127.0.0.1:' + it.local_port;
      var tdLocal = document.createElement('td'); 
      tdLocal.className = 'py-3 px-4';
      var localDiv = document.createElement('div');
      localDiv.className = 'font-mono text-sm text-gray-800';
      localDiv.textContent = local;
      var copyBtn = document.createElement('button');
      copyBtn.className = 'text-xs text-blue-600 hover:underline mt-1';
      copyBtn.textContent = 'Copy';
      copyBtn.onclick = function(){ 
        navigator.clipboard.writeText(local);
        showToast('Copied: ' + local);
      };
      tdLocal.appendChild(localDiv);
      tdLocal.appendChild(copyBtn);
      tr.appendChild(tdLocal);

      var tdStatus = document.createElement('td'); 
      tdStatus.className = 'py-3 px-4';
      var badge = document.createElement('span');
      badge.className = it.status === 'live' ? 'status-active' : 'status-inactive';
      badge.textContent = it.status === 'live' ? 'Active' : 'Inactive';
      tdStatus.appendChild(badge); 
      tr.appendChild(tdStatus);

      // Exit IP column
      var tdExitIP = document.createElement('td');
      tdExitIP.className = 'py-3 px-4';
      if(it.status === 'live'){
        var btnCheckIP = document.createElement('button');
        btnCheckIP.className = 'action-btn text-purple-600 hover:bg-purple-50';
        btnCheckIP.textContent = '🌐 Check IP';
        btnCheckIP.onclick = function(){ checkExitIP(it.id, tdExitIP); };
        tdExitIP.appendChild(btnCheckIP);
      } else {
        tdExitIP.textContent = '-';
      }
      tr.appendChild(tdExitIP);

      var tdAction = document.createElement('td'); 
      tdAction.className = 'py-3 px-4';
      var btnStart = document.createElement('button'); 
      btnStart.className = 'action-btn mr-2 text-green-600'; 
      btnStart.textContent = '▶';
      btnStart.onclick = function(){ startProxy(it.id); };
      var btnStop = document.createElement('button'); 
      btnStop.className = 'action-btn text-orange-600'; 
      btnStop.textContent = '⏸';
      btnStop.onclick = function(){ stopProxy(it.id); };
      tdAction.appendChild(btnStart); 
      tdAction.appendChild(btnStop);
      tr.appendChild(tdAction);

      return tr;
    }

    function checkExitIP(id, cellEl){
      cellEl.textContent = '...';
      GET('/api/check-ip?id=' + encodeURIComponent(id)).then(function(result){
        cellEl.textContent = '✅ ' + result.ip;
        cellEl.className = 'py-3 px-4 font-mono text-xs text-green-700';
      }).catch(function(e){
        cellEl.textContent = '❌ ' + e.message;
        cellEl.className = 'py-3 px-4 text-xs text-red-600';
      });
    }

    function render(list){
      rowsEl.innerHTML = '';
      countEl.textContent = 'Total: ' + list.length + ' proxies';
      if(list.length === 0){
        var tr = document.createElement('tr');
        var td = document.createElement('td');
        td.colSpan = 6;
        td.className = 'py-8 text-center text-gray-500';
        td.textContent = 'No proxies yet. Add one above!';
        tr.appendChild(td);
        rowsEl.appendChild(tr);
        return;
      }
      for(var i=0; i<list.length; i++){ 
        rowsEl.appendChild(rowItem(list[i], i)); 
      }
    }

    function reload(){
      GET('/api/list').then(function(data){ 
        DATA = data; 
        // Filter active proxies (has local_port > 0) for Proxies tab
        var activeProxies = DATA.items.filter(function(it){
          return it.local_port > 0;
        });
        render(activeProxies); 
        renderPool(DATA.items);
      }).catch(function(e){ 
        console.error(e); 
        showToast('Error loading');
      });
    }

    function renderPool(items){
      var poolRows = document.getElementById('poolRows');
      var poolCount = document.getElementById('poolCountText');
      if(!poolRows) return;
      
      // Filter pool items (stopped or no local port)
      var pool = items.filter(function(it){ 
        return it.status === 'stopped' || it.local_port === 0; 
      });
      
      poolRows.innerHTML = '';
      poolCount.textContent = 'Pool: ' + pool.length + ' proxies';
      
      if(pool.length === 0){
        var tr = document.createElement('tr');
        var td = document.createElement('td');
        td.colSpan = 4;
        td.className = 'py-8 text-center text-gray-500';
        td.textContent = 'Pool is empty. Order proxies to add to pool.';
        tr.appendChild(td);
        poolRows.appendChild(tr);
        return;
      }
      
      pool.forEach(function(it, idx){
        var tr = document.createElement('tr');
        tr.className = 'hover:bg-gray-50';
        
        var tdIdx = document.createElement('td');
        tdIdx.className = 'py-3 px-4 text-gray-600';
        tdIdx.textContent = String(idx+1);
        tr.appendChild(tdIdx);
        
        var up = (it.user ? it.user + ':' + it.pass + '@' : '') + it.host + ':' + it.port;
        var tdUp = document.createElement('td');
        tdUp.className = 'py-3 px-4 font-mono text-sm';
        tdUp.textContent = up;
        tr.appendChild(tdUp);
        
        var tdStatus = document.createElement('td');
        tdStatus.className = 'py-3 px-4';
        var badge = document.createElement('span');
        badge.className = 'status-inactive';
        badge.textContent = 'In Pool';
        tdStatus.appendChild(badge);
        tr.appendChild(tdStatus);
        
        var tdAction = document.createElement('td');
        tdAction.className = 'py-3 px-4';
        var btnStart = document.createElement('button');
        btnStart.className = 'action-btn mr-2 text-green-600';
        btnStart.textContent = '▶ Start';
        btnStart.onclick = function(){ startProxy(it.id); };
        var btnDel = document.createElement('button');
        btnDel.className = 'action-btn text-red-600';
        btnDel.textContent = '🗑 Del';
        btnDel.onclick = function(){ deleteProxy(it.id); };
        tdAction.appendChild(btnStart);
        tdAction.appendChild(btnDel);
        tr.appendChild(tdAction);
        
        poolRows.appendChild(tr);
      });
    }

    function startAllPool(){
      var pool = DATA.items.filter(function(it){ 
        return it.status === 'stopped' || it.local_port === 0; 
      });
      
      if(pool.length === 0){
        showToast('Pool is empty');
        return;
      }
      
      var ok = 0, fail = 0;
      function startNext(i){
        if(i >= pool.length){
          showToast('Started: ' + ok + ' ok, ' + fail + ' fail');
          reload();
          return;
        }
        POST('/api/start?id=' + encodeURIComponent(pool[i].id)).then(function(){
          ok++;
          startNext(i+1);
        }).catch(function(){
          fail++;
          startNext(i+1);
        });
      }
      startNext(0);
    }

    function clearPool(){
      var pool = DATA.items.filter(function(it){ 
        return it.status === 'stopped' || it.local_port === 0; 
      });
      
      if(pool.length === 0){
        showToast('Pool is empty');
        return;
      }
      
      if(!confirm('Delete all ' + pool.length + ' proxies in pool?')) return;
      
      var ok = 0, fail = 0;
      function delNext(i){
        if(i >= pool.length){
          showToast('Deleted: ' + ok + ' ok, ' + fail + ' fail');
          reload();
          return;
        }
        POST('/api/remove?id=' + encodeURIComponent(pool[i].id)).then(function(){
          ok++;
          delNext(i+1);
        }).catch(function(){
          fail++;
          delNext(i+1);
        });
      }
      delNext(0);
    }

    function handleSearch(){
      var query = document.getElementById('searchInput').value.toLowerCase();
      // Always filter only active proxies (local_port > 0)
      var activeProxies = DATA.items.filter(function(it){
        return it.local_port > 0;
      });
      if(!query){ render(activeProxies); return; }
      var filtered = activeProxies.filter(function(it){
        var up = (it.user ? it.user + ':' + it.pass + '@' : '') + it.host + ':' + it.port;
        var local = '127.0.0.1:' + it.local_port;
        return up.toLowerCase().indexOf(query) !== -1 || local.indexOf(query) !== -1;
      });
      render(filtered);
    }

    function handleAddSingle(){
      var line = document.getElementById('singleProxyInput').value.trim();
      if(!line){ showToast('Enter proxy address'); return; }
      POST('/api/add', line).then(function(){
        document.getElementById('singleProxyInput').value = '';
        showToast('Added successfully');
        reload();
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function handleSyncAPI(){
      var url = document.getElementById('apiUrlInput').value.trim();
      if(!url){ showToast('Enter API URL'); return; }
      GET('/api/sync?url=' + encodeURIComponent(url)).then(function(result){
        showToast('Synced: ' + (result.added || 0) + ' added');
        reload();
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function handleExport(){
      window.open('/api/export-local', '_blank');
    }

    function openBulkModal(){
      document.getElementById('bulkModal').classList.remove('hidden');
    }

    function closeBulkModal(){
      document.getElementById('bulkModal').classList.add('hidden');
    }

    function handleBulkAdd(){
      var text = document.getElementById('bulkTextarea').value;
      var lines = text.split('\n').map(function(s){ return s.trim(); }).filter(Boolean);
      if(!lines.length){ showToast('No lines'); return; }
      var ok = 0, fail = 0;
      function addLine(i){
        if(i >= lines.length){
          closeBulkModal();
          document.getElementById('bulkTextarea').value = '';
          showToast('Done: ' + ok + ' ok, ' + fail + ' fail');
          reload();
          return;
        }
        POST('/api/add', lines[i]).then(function(){
          ok++;
          addLine(i+1);
        }).catch(function(){
          fail++;
          addLine(i+1);
        });
      }
      addLine(0);
    }

    function startProxy(id){
      POST('/api/start?id=' + encodeURIComponent(id)).then(function(){ 
        showToast('Started');
        reload(); 
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function stopProxy(id){
      POST('/api/stop?id=' + encodeURIComponent(id)).then(function(){ 
        showToast('Stopped');
        reload(); 
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function deleteProxy(id){
      if(!confirm('Delete this proxy?')) return;
      POST('/api/remove?id=' + encodeURIComponent(id)).then(function(){ 
        showToast('Deleted');
        reload(); 
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function toggleAutoRefresh(){
      var checked = document.getElementById('autoRefreshCheck').checked;
      if(autoRefreshTimer){ clearInterval(autoRefreshTimer); autoRefreshTimer = null; }
      if(checked){
        autoRefreshTimer = setInterval(reload, 5000);
        showToast('Auto refresh ON');
      } else {
        showToast('Auto refresh OFF');
      }
    }

    function loadCloudminiRegions(){
      var token = document.getElementById('cloudminiToken').value.trim();
      var type = document.getElementById('cloudminiType').value;
      
      if(!token){ showToast('Enter CloudMini Token'); return; }
      
      var url = '/api/cloudmini/regions?token=' + encodeURIComponent(token) + '&type=' + type;
      
      showToast('Loading regions...');
      
      fetch(url, {
        headers: hdr()
      }).then(function(r){
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); });
        return r.json();
      }).then(function(result){
        if(!result.data || !result.data.length){
          throw new Error('No regions returned');
        }
        
        var typeData = result.data.find(function(d){ return d.type === type; });
        if(!typeData || !typeData.region){
          throw new Error('No regions for type: ' + type);
        }
        
        var select = document.getElementById('cloudminiRegion');
        select.innerHTML = '<option value="">-- Chọn region --</option>';
        
        typeData.region.forEach(function(region){
          var opt = document.createElement('option');
          opt.value = region;
          opt.textContent = region;
          select.appendChild(opt);
        });
        
        showToast('Loaded ' + typeData.region.length + ' regions');
      }).catch(function(e){
        showToast('Error: ' + e.message);
      });
    }

    function handleCloudMiniOrder(){
      var token = document.getElementById('cloudminiToken').value.trim();
      var type = document.getElementById('cloudminiType').value;
      var region = document.getElementById('cloudminiRegion').value;
      var quantity = document.getElementById('cloudminiQuantity').value;
      var autoStart = document.getElementById('cloudminiAutoStart').checked;
      
      if(!token){ showToast('Enter CloudMini Token'); return; }
      if(!region){ showToast('Select region'); return; }
      if(!quantity || quantity < 1){ showToast('Enter quantity'); return; }
      
      var url = '/api/cloudmini/order';
      var body = JSON.stringify({
        type: type,
        region: region,
        quantity: parseInt(quantity),
        period: 30
      });
      
      showToast('Ordering from CloudMini... (this may take up to 5 minutes)');
      
      var headers = hdr();
      headers['X-CloudMini-Token'] = token;
      headers['Content-Type'] = 'application/json';
      
      fetch(url, {
        method: 'POST',
        headers: headers,
        body: body
      }).then(function(r){
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); });
        return r.json();
      }).then(function(result){
        if(!result.data || !result.data.length){
          throw new Error(result.msg || 'No proxies returned');
        }
        
        showToast('CloudMini order complete! Adding ' + result.data.length + ' proxies...');
        
        var ok = 0, fail = 0;
        var proxies = result.data;
        
        function addProxy(i){
          if(i >= proxies.length){
            showToast('CloudMini: ' + ok + ' added, ' + fail + ' failed');
            reload();
            return;
          }
          
          var item = proxies[i];
          // Parse CloudMini format: ip="hostname:port", https=port, user=username
          var hostname = item.ip.split(':')[0];
          var proxyLine = hostname + ':' + item.https + ':' + item.user + ':' + item.password;
          
          var endpoint = autoStart ? '/api/add' : '/api/add-pool';
          
          POST(endpoint, proxyLine).then(function(){
            ok++;
            addProxy(i+1);
          }).catch(function(){
            fail++;
            addProxy(i+1);
          });
        }
        
        addProxy(0);
      }).catch(function(e){
        showToast('CloudMini Error: ' + e.message);
      });
    }

    function handleCloudMiniSyncToPool(){
      var token = document.getElementById('cloudminiSyncToken').value.trim();
      
      if(!token){ showToast('Enter CloudMini Token'); return; }
      
      var url = '/api/cloudmini/sync?token=' + encodeURIComponent(token);
      
      showToast('Syncing from CloudMini... (this may take a while)');
      
      fetch(url, {
        headers: hdr()
      }).then(function(r){
        if(!r.ok) return r.text().then(function(t){ throw new Error(t); });
        return r.json();
      }).then(function(result){
        var msg = 'CloudMini Sync: ' + result.total + ' total, ' + result.added + ' added, ' + result.existing + ' existing';
        if(result.errors && result.errors.length > 0){
          msg += ', ' + result.errors.length + ' errors';
        }
        showToast(msg);
        reload();
      }).catch(function(e){
        showToast('CloudMini Sync Error: ' + e.message);
      });
    }

    reload();
  </script>
</body>
</html>`
