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

  const archModal = document.getElementById('archModal');
  const btnOpenArchModal = document.getElementById('btnOpenArchModal');
  const btnCloseArchModal = document.getElementById('btnCloseArchModal');

  if (btnOpenArchModal && archModal && btnCloseArchModal) {
    btnOpenArchModal.addEventListener('click', () => {
      archModal.classList.remove('hidden');
    });
    btnCloseArchModal.addEventListener('click', () => {
      archModal.classList.add('hidden');
    });
    archModal.addEventListener('click', (e) => {
      if (e.target === archModal) {
        archModal.classList.add('hidden');
      }
    });
  }

  let localFiles = [];

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
      setTimeout(() => showDashboard(), 1200);
    } catch (err) {
      alert(`Error al crear red: ${err}`);
    }
  });

  // 2. Join Network
  btnJoin.addEventListener('click', async () => {
    const tok = tokenInput.value.trim();
    if (!tok) {
      alert('Introduce un token de invitación válido.');
      return;
    }
    try {
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.JoinNetwork(tok);
      }
      showDashboard();
    } catch (err) {
      alert(`Token inválido: ${err}`);
    }
  });

  // 3. Copy Token
  btnCopy.addEventListener('click', () => {
    navigator.clipboard.writeText(generatedToken.textContent);
    btnCopy.textContent = 'COPIADO';
    setTimeout(() => { btnCopy.textContent = 'COPIAR'; }, 2000);
  });

  // 4. Storage Slider Control (RF-20)
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
        alert(`¡Token de invitación copiado al portapapeles!\n\n${token}`);
      } catch (err) {
        alert(`Error: ${err}`);
      }
    } else {
      const mockTok = `aldea1_invite_${Date.now()}`;
      navigator.clipboard.writeText(mockTok);
      alert(`[MOCK] Token copiado: ${mockTok}`);
    }
  });

  // 5b. Pause / Resume Node Service (RF-22)
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
    } else {
      btnTogglePause.textContent = 'PAUSAR SERVICIO DE NODO';
      statusText.textContent = 'NETWORK ONLINE';
      statusText.style.color = 'var(--emerald)';
      if (healthBadge) {
        healthBadge.textContent = 'HEALTHY';
        healthBadge.className = 'badge badge-emerald';
      }
    }
  });

  // 6. Drag & Drop Zone (RF-21)
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
      renderFilesTable();
    } catch (err) {
      alert(`Error al subir archivo: ${err}`);
    }
  }

  function showDashboard() {
    onboardingView.classList.add('hidden');
    dashboardView.classList.remove('hidden');
    renderFilesTable();
  }

  function renderFilesTable() {
    fileCountBadge.textContent = `${localFiles.length} ARCHIVOS`;
    if (localFiles.length === 0) {
      filesTableBody.innerHTML = `
        <tr class="empty-state">
          <td colspan="5">Ningún archivo almacenado en el pool. Arrastra un archivo arriba para iniciar.</td>
        </tr>
      `;
      return;
    }

    filesTableBody.innerHTML = localFiles.map(f => `
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
    alert(`Descargando archivo [${fileID}] y reconstruyendo desde fragmentos 4+2...`);
  };
});
