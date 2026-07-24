async function loadCandidates() {
  const res = await fetch('/fonts.json');
  if (!res.ok) throw new Error(`/fonts.json ${res.status}`);
  return res.json();
}

async function paintSpecimen(canvas, id) {
  const res = await fetch(`/font-specimen.png?id=${encodeURIComponent(id)}&${Date.now()}`);
  if (!res.ok) throw new Error(`specimen ${id}: ${res.status}`);
  const bitmap = await createImageBitmap(await res.blob());
  canvas.width = 250;
  canvas.height = 122;
  const ctx = canvas.getContext('2d', { alpha: false });
  ctx.imageSmoothingEnabled = false;
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, 250, 122);
  ctx.drawImage(bitmap, 0, 0);
  bitmap.close();
}

async function main() {
  const list = document.getElementById('list');
  const cands = await loadCandidates();
  for (const c of cands) {
    const card = document.createElement('article');
    card.className = 'card';
    card.innerHTML = `
      <h2>${c.name}</h2>
      <p class="notes">${c.notes || ''} · <a href="${c.url}" target="_blank" rel="noopener">${c.url.replace(/^https?:\/\//, '')}</a></p>
      <div class="bezel"><div class="screen"><canvas width="250" height="122" aria-label="${c.id}"></canvas></div></div>
      <p class="err" hidden></p>
    `;
    list.appendChild(card);
    const canvas = card.querySelector('canvas');
    const err = card.querySelector('.err');
    try {
      await paintSpecimen(canvas, c.id);
    } catch (e) {
      err.hidden = false;
      err.textContent = String(e);
    }
  }
}

main().catch((e) => {
  document.getElementById('list').textContent =
    String(e) + ' — start motohud so /fonts.json is available';
});
