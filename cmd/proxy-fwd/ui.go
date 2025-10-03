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
      <a href="#" class="sidebar-item active flex items-center gap-3 px-3 py-2 text-white">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor"><path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2M9 17H7v-7h2v7m4 0h-2V7h2v10m4 0h-2v-4h2v4z"/></svg>
        <span>Proxies</span>
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
        <h3 class="font-bold mb-3">Sync from API</h3>
        <div class="flex gap-2">
          <input id="apiUrlInput" type="text" placeholder="API URL trả về danh sách proxy (text hoặc JSON array)" class="flex-1 px-3 py-2 border rounded-lg">
          <button onclick="handleSyncAPI()" class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            🔄 Sync API
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

      var up = (it.user ? it.user + ':' + it.pass + '@' : '') + it.host + ':' + it.port;
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
      divMain.textContent = up;
      var divId = document.createElement('div');
      divId.className = 'text-xs text-gray-500';
      divId.textContent = 'ID: ' + it.id;
      divText.appendChild(divMain);
      divText.appendChild(divId);
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
      btnStart.className = 'action-btn mr-1 text-green-600'; 
      btnStart.textContent = '▶';
      btnStart.onclick = function(){ startProxy(it.id); };
      var btnStop = document.createElement('button'); 
      btnStop.className = 'action-btn mr-1 text-orange-600'; 
      btnStop.textContent = '⏸';
      btnStop.onclick = function(){ stopProxy(it.id); };
      var btnDel = document.createElement('button'); 
      btnDel.className = 'action-btn text-red-600'; 
      btnDel.textContent = '🗑';
      btnDel.onclick = function(){ deleteProxy(it.id); };
      tdAction.appendChild(btnStart); 
      tdAction.appendChild(btnStop); 
      tdAction.appendChild(btnDel);
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
        render(DATA.items); 
      }).catch(function(e){ 
        console.error(e); 
        showToast('Error loading');
      });
    }

    function handleSearch(){
      var query = document.getElementById('searchInput').value.toLowerCase();
      if(!query){ render(DATA.items); return; }
      var filtered = DATA.items.filter(function(it){
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

    reload();
  </script>
</body>
</html>`
