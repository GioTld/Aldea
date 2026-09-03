document.addEventListener('DOMContentLoaded', () => {
  // DOM Elements
  const onboardingView = document.getElementById('onboardingView');
  const dashboardView = document.getElementById('dashboardView');
  const trackerInput = document.getElementById('trackerInput');
  const tokenInput = document.getElementById('tokenInput');
  const btnCreate = document.getElementById('btnCreate');
  const btnJoin = document.getElementById('btnJoin');
  const tokenResult = document.getElementById('tokenResult');
  const generatedToken = document.getElementById('generatedToken');
  const btnCopy = document.getElementById('btnCopy');
  
  const storageSlider = document.getElementById('storageSlider');
  const allocDisplay = document.getElementById('allocDisplay');
  const usedDisplay = document.getElementById('usedDisplay');
  const usageBar = document.getElementById('usageBar');
  const btnGenInvite = document.getElementById('btnGenInvite');
  
  const dropZone = document.getElementById('dropZone');
  const fileInput = document.getElementById('fileInput');
  const filesTableBody = document.getElementById('filesTableBody');
  const fileCountBadge = document.getElementById('fileCountBadge');
  const fileSearchInput = document.getElementById('fileSearchInput');

  const computeTableBody = document.getElementById('computeTableBody');
  const computeCountBadge = document.getElementById('computeCountBadge');
  const computeSearchInput = document.getElementById('computeSearchInput');

  const archModal = document.getElementById('archModal');
  const btnOpenArchModal = document.getElementById('btnOpenArchModal');
  const btnCloseArchModal = document.getElementById('btnCloseArchModal');

  const deployModal = document.getElementById('deployModal');
  const btnOpenDeployModal = document.getElementById('btnOpenDeployModal');
  const btnCloseDeployModal = document.getElementById('btnCloseDeployModal');
  const btnConfirmDeploy = document.getElementById('btnConfirmDeploy');
  const yamlManifestInput = document.getElementById('yamlManifestInput');

  const toastContainer = document.getElementById('toastContainer');

  // Tabs
  const tabStorage = document.getElementById('tabStorage');
  const tabCompute = document.getElementById('tabCompute');
  const tabMetrics = document.getElementById('tabMetrics');

  const viewStorage = document.getElementById('viewStorage');
  const viewCompute = document.getElementById('viewCompute');
  const viewMetrics = document.getElementById('viewMetrics');

  // Toast Helper
  function showToast(message, isError = false) {
    const toast = document.createElement('div');
    toast.className = `toast ${isError ? 'toast-error' : 'toast-success'}`;
    toast.textContent = message;
    toastContainer.appendChild(toast);
    setTimeout(() => {
      toast.remove();
    }, 3500);
  }

  // Modals
  if (btnOpenArchModal && archModal && btnCloseArchModal) {
    btnOpenArchModal.addEventListener('click', () => archModal.classList.remove('hidden'));
    btnCloseArchModal.addEventListener('click', () => archModal.classList.add('hidden'));
    archModal.addEventListener('click', (e) => { if (e.target === archModal) archModal.classList.add('hidden'); });
  }

  if (btnOpenDeployModal && deployModal && btnCloseDeployModal) {
    btnOpenDeployModal.addEventListener('click', () => deployModal.classList.remove('hidden'));
    btnCloseDeployModal.addEventListener('click', () => deployModal.classList.add('hidden'));
    deployModal.addEventListener('click', (e) => { if (e.target === deployModal) deployModal.classList.add('hidden'); });
  }

  if (btnConfirmDeploy) {
    btnConfirmDeploy.addEventListener('click', () => {
      const yamlData = yamlManifestInput.value.trim();
      if (!yamlData) {
        showToast('Introduce un manifiesto YAML válido para desplegar.', true);
        return;
      }
      deployModal.classList.add('hidden');
      yamlManifestInput.value = '';
      showToast('Carga de cómputo desplegada con éxito en la red P2P.');
      
      computeWorkloads.push({
        id: `wl-app-${Date.now().toString().slice(-4)}`,
        name: 'custom-workload',
        image: 'nginx:alpine',
        state: 'RUNNING',
        ip: '10.244.0.18'
      });
      renderComputeTable();
    });
  }

  // Tabs Switcher
  function switchTab(activeTab, activeView) {
    [tabStorage, tabCompute, tabMetrics].forEach(t => t.classList.remove('active'));
    [viewStorage, viewCompute, viewMetrics].forEach(v => v.classList.add('hidden'));

    activeTab.classList.add('active');
    activeView.classList.remove('hidden');
  }

  if (tabStorage && tabCompute && tabMetrics) {
    tabStorage.addEventListener('click', () => { switchTab(tabStorage, viewStorage); renderFilesTable(); });
    tabCompute.addEventListener('click', () => { switchTab(tabCompute, viewCompute); renderComputeTable(); });
    tabMetrics.addEventListener('click', () => { switchTab(tabMetrics, viewMetrics); renderMetricsView(); });
  }

  // 1. Create Network
  btnCreate.addEventListener('click', async () => {
    const addr = trackerInput.value.trim() || '127.0.0.1:9090';
    try {
      let token = '';
      if (window.go && window.go.main && window.go.main.App) {
        token = await window.go.main.App.CreateNetwork(addr);
      } else {
        token = `aldea1_mock_token_${Date.now()}`;
      }
      generatedToken.textContent = token;
      tokenResult.classList.remove('hidden');
      showToast('Red creada exitosamente. Token listo para compartir.');
      setTimeout(() => showDashboard(), 1200);
    } catch (err) {
      showToast(`Error al crear red: ${err}`, true);
    }
  });

  // 2. Join Network
  btnJoin.addEventListener('click', async () => {
    const tok = tokenInput.value.trim();
    if (!tok) {
      showToast('Introduce un token de invitación válido.', true);
      return;
    }
    try {
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.JoinNetwork(tok);
      }
      showToast('Nodo integrado a la red privada exitosamente.');
      showDashboard();
    } catch (err) {
      showToast(`Token inválido: ${err}`, true);
    }
  });

  // 3. Copy Token
  btnCopy.addEventListener('click', () => {
    navigator.clipboard.writeText(generatedToken.textContent);
    btnCopy.textContent = 'COPIADO';
    showToast('Token de invitación copiado al portapapeles.');
    setTimeout(() => { btnCopy.textContent = 'COPIAR'; }, 2000);
  });

  // 4. Storage Slider Control
  storageSlider.addEventListener('input', async (e) => {
    const gb = parseInt(e.target.value, 10);
    allocDisplay.textContent = `${gb.toFixed(1)} GB`;
    const bytesAlloc = gb * 1024 * 1024 * 1024;

    if (window.go && window.go.main && window.go.main.App) {
      try {
        await window.go.main.App.SetStorageAllocation(bytesAlloc);
      } catch (err) {
        console.error('Error updating storage allocation:', err);
      }
    }
  });

  // 5. Generate Extra Invite
  btnGenInvite.addEventListener('click', async () => {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const token = await window.go.main.App.CreateNetwork('127.0.0.1:9090');
        navigator.clipboard.writeText(token);
        showToast('Nuevo token de invitación copiado al portapapeles.');
      } catch (err) {
        showToast(`Error: ${err}`, true);
      }
    } else {
      const mockTok = `aldea1_invite_${Date.now()}`;
      navigator.clipboard.writeText(mockTok);
      showToast('Token de invitación copiado al portapapeles.');
    }
  });

  // 5b. Pause / Resume Node Service
  const btnTogglePause = document.getElementById('btnTogglePause');
  const statusText = document.getElementById('statusText');
  const healthBadge = document.getElementById('healthBadge');
  let isNodePaused = false;

  btnTogglePause.addEventListener('click', async () => {
    isNodePaused = !isNodePaused;
    if (window.go && window.go.main && window.go.main.App) {
      await window.go.main.App.PauseNode(isNodePaused);
    }

    if (isNodePaused) {
      btnTogglePause.textContent = 'REANUDAR SERVICIO DE NODO';
      statusText.textContent = 'NETWORK PAUSED';
      statusText.style.color = '#eab308';
      if (healthBadge) {
        healthBadge.textContent = 'PAUSED';
        healthBadge.className = 'badge';
      }
      showToast('Servicio del nodo pausado.');
    } else {
      btnTogglePause.textContent = 'PAUSAR SERVICIO DE NODO';
      statusText.textContent = 'NETWORK ONLINE';
      statusText.style.color = 'var(--emerald)';
      if (healthBadge) {
        healthBadge.textContent = 'HEALTHY';
        healthBadge.className = 'badge badge-emerald';
      }
      showToast('Servicio del nodo reanudado.');
    }
  });

  // 6. Drag & Drop Zone
  dropZone.addEventListener('click', () => fileInput.click());

  ['dragenter', 'dragover'].forEach(eventName => {
    dropZone.addEventListener(eventName, (e) => {
      e.preventDefault();
      e.stopPropagation();
      dropZone.classList.add('dragover');
    }, false);
  });

  ['dragleave', 'drop'].forEach(eventName => {
    dropZone.addEventListener(eventName, (e) => {
      e.preventDefault();
      e.stopPropagation();
      dropZone.classList.remove('dragover');
    }, false);
  });

  dropZone.addEventListener('drop', (e) => {
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      handleFileUpload(files[0]);
    }
  });

  fileInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
      handleFileUpload(e.target.files[0]);
    }
  });

  async function handleFileUpload(file) {
    const size = file.size;
    const name = file.name;

    try {
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.UploadFile(name, size);
      }
      localFiles.push({
        id: `file-${Date.now()}`,
        name: name,
        size: formatBytes(size),
        scheme: '4+4 (XChaCha20)',
        date: new Date().toISOString().split('T')[0]
      });
      showToast(`Archivo '${name}' fragmentado y cifrado exitosamente.`);
      renderFilesTable();
    } catch (err) {
      showToast(`Error al subir archivo: ${err}`, true);
    }
  }

  // Live Search Filters
  if (fileSearchInput) {
    fileSearchInput.addEventListener('input', () => renderFilesTable());
  }

  if (computeSearchInput) {
    computeSearchInput.addEventListener('input', () => renderComputeTable());
  }

  function showDashboard() {
    onboardingView.classList.add('hidden');
    dashboardView.classList.remove('hidden');
    renderFilesTable();
    renderComputeTable();
  }

  function renderFilesTable() {
    const query = fileSearchInput ? fileSearchInput.value.toLowerCase().trim() : '';
    const filtered = localFiles.filter(f => f.name.toLowerCase().includes(query));

    fileCountBadge.textContent = `${filtered.length} ARCHIVOS`;
    if (filtered.length === 0) {
      filesTableBody.innerHTML = `
        <tr class="empty-state">
          <td colspan="5">Ningún archivo almacenado en el pool. Arrastra un archivo arriba para iniciar.</td>
        </tr>
      `;
      return;
    }

    filesTableBody.innerHTML = filtered.map(f => `
      <tr>
        <td class="mono-bold">${escapeHtml(f.name)}</td>
        <td class="mono">${f.size}</td>
        <td><span class="badge badge-emerald">${f.scheme}</span></td>
        <td class="mono">${f.date}</td>
        <td>
          <button class="btn btn-secondary btn-sm" onclick="downloadFile('${f.id}')">DESCARGAR</button>
        </td>
      </tr>
    `).join('');
  }

  let localFiles = [];

  async function renderComputeTable() {
    let workloads = [];
    if (window.go && window.go.main && window.go.main.App) {
      try {
        workloads = await window.go.main.App.GetComputeWorkloads();
      } catch (err) {
        console.error('Error fetching compute workloads:', err);
      }
    }

    const query = computeSearchInput ? computeSearchInput.value.toLowerCase().trim() : '';
    const filtered = (workloads || []).filter(w => (w.name || '').toLowerCase().includes(query) || (w.workload_id || '').toLowerCase().includes(query));

    computeCountBadge.textContent = `${filtered.length} CARGA(S) ACTIVA(S)`;
    if (filtered.length === 0) {
      computeTableBody.innerHTML = `
        <tr class="empty-state">
          <td colspan="5">No hay cargas de cómputo en ejecución. Haz clic en '+ NUEVA CARGA' para desplegar.</td>
        </tr>
      `;
      return;
    }

    computeTableBody.innerHTML = filtered.map(w => `
      <tr>
        <td class="mono">${escapeHtml(w.workload_id)}</td>
        <td>${escapeHtml(w.name)}</td>
        <td class="mono">${escapeHtml(w.image)}</td>
        <td><span class="badge badge-emerald">${w.state}</span></td>
        <td class="mono">${w.ip_address}</td>
      </tr>
    `).join('');
  }

  async function renderMetricsView() {
    if (window.go && window.go.main && window.go.main.App) {
      try {
        const metrics = await window.go.main.App.GetNetworkMetrics();
        const speedUp = document.getElementById('speedUploadVal');
        const speedDown = document.getElementById('speedDownloadVal');
        const peersTable = document.getElementById('peersTableBody');

        if (speedUp) speedUp.textContent = `${(metrics.upload_speed_kbps || 0).toFixed(1)} KB/s`;
        if (speedDown) speedDown.textContent = `${(metrics.download_speed_kbps || 0).toFixed(1)} KB/s`;

        const peers = metrics.peers || [];
        if (peersTable) {
          if (peers.length === 0) {
            peersTable.innerHTML = `<tr class="empty-state"><td colspan="4">No hay nodos pares conectados actualmente.</td></tr>`;
          } else {
            peersTable.innerHTML = peers.map(p => `
              <tr>
                <td class="mono">${escapeHtml(p.node_id)}</td>
                <td class="mono">${escapeHtml(p.os)}</td>
                <td class="mono">${p.latency_ms} ms</td>
                <td><span class="badge badge-emerald">ONLINE</span></td>
              </tr>
            `).join('');
          }
        }
      } catch (err) {
        console.error('Error fetching metrics:', err);
      }
    }
  }

  function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function escapeHtml(str) {
    return str.replace(/[&<>"']/g, m => ({
      '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'
    }[m]));
  }

  window.downloadFile = function(fileID) {
    showToast(`Iniciando descarga y reconstrucción desde fragmentos 4+4...`);
  };
});
