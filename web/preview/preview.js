const NATIVE_W = 250;
const NATIVE_H = 122;

function syncPanels() {
  const screen = document.getElementById('screenSel').value;
  document.getElementById('navFields').hidden = screen !== 'nav';
  document.getElementById('mediaFields').hidden = screen !== 'media';
  document.getElementById('statusFields').hidden = screen !== 'status';
}

async function pushState() {
  const screen = document.getElementById('screenSel').value;
  await fetch('/button', { method: 'POST', body: 'action_long' }); // home = nav
  if (screen === 'media') {
    await fetch('/button', { method: 'POST', body: 'next' });
  } else if (screen === 'status') {
    await fetch('/button', { method: 'POST', body: 'next' });
    await fetch('/button', { method: 'POST', body: 'next' });
  }

  if (screen === 'nav') {
    const distance = document.getElementById('distance').value;
    const m = parseInt(distance, 10);
    await fetch('/nav', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type: 'nav',
        active: !document.getElementById('idle').value,
        instruction: document.getElementById('road').value,
        distance_m: Number.isFinite(m) ? m : 200,
        distance_text: distance,
        road: document.getElementById('road').value,
        eta_min: 12,
        maneuver: document.getElementById('maneuver').value,
      }),
    });
  } else if (screen === 'media') {
    await fetch('/media', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type: 'media',
        playing: document.getElementById('playing').value.toUpperCase().includes('PLAY'),
        title: document.getElementById('title').value,
        artist: document.getElementById('artist').value,
      }),
    });
  } else {
    await fetch('/button', { method: 'POST', body: 'action' });
  }
  await refreshFrame();
}

async function refreshFrame() {
  const canvas = document.getElementById('frame');
  if (!canvas) throw new Error('frame canvas missing');

  const res = await fetch(`/frame.png?${Date.now()}`);
  if (!res.ok) throw new Error(`frame.png ${res.status}`);
  const bitmap = await createImageBitmap(await res.blob());

  // Keep the backing store at native Inky resolution — never draw at display size.
  canvas.width = NATIVE_W;
  canvas.height = NATIVE_H;
  const ctx = canvas.getContext('2d', { alpha: false });
  ctx.imageSmoothingEnabled = false;
  ctx.clearRect(0, 0, NATIVE_W, NATIVE_H);
  ctx.drawImage(bitmap, 0, 0, NATIVE_W, NATIVE_H);
  bitmap.close();
}

document.getElementById('screenSel').addEventListener('change', () => {
  syncPanels();
});
document.getElementById('apply').addEventListener('click', () => pushState().catch(console.error));
document.getElementById('reload').addEventListener('click', () => refreshFrame().catch(console.error));

syncPanels();
refreshFrame().catch((err) => {
  const screen = document.getElementById('screen');
  screen.textContent = String(err) + ' — start motohud so /frame.png is available';
});

setInterval(() => refreshFrame().catch(() => {}), 1000);
