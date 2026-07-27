import { paintJunction, IMPLEMENTED_KINDS, TODO_KINDS } from './junction-draw.js';

const BASE = window.__MOTO_HUD_BASE__ || '/moto-hud/';
const W = 70;
const H = 80;

const listEl = document.getElementById('maneuverList');
const metaEl = document.getElementById('fixtureMeta');
const canvas = document.getElementById('pane');
const jsonEl = document.getElementById('irJson');
const statusEl = document.getElementById('status');
const titleEl = document.getElementById('paneTitle');
const ctx = canvas.getContext('2d');

let poc = null;
let selected = 0;

function redraw() {
  if (!poc?.maneuvers?.length) return;
  const m = poc.maneuvers[selected];
  const j = m.junction;
  paintJunction(ctx, j, W, H);
  titleEl.textContent = `${m.label || m.index} · ${j?.kind || '?'}`;
  const impl = IMPLEMENTED_KINDS.includes(j?.kind);
  const todo = TODO_KINDS.includes(j?.kind);
  metaEl.textContent = impl
    ? `template: ${j.kind}`
    : todo
      ? `template: ${j.kind} (TODO placeholder)`
      : `template: ${j?.kind} (unknown → simple fallback)`;
  jsonEl.textContent = JSON.stringify(j, null, 2);
  statusEl.textContent = `${selected + 1} / ${poc.maneuvers.length}`;
  [...listEl.querySelectorAll('button')].forEach((b, i) => {
    b.classList.toggle('active', i === selected);
  });
}

function renderList() {
  listEl.replaceChildren(
    ...poc.maneuvers.map((m, i) => {
      const btn = document.createElement('button');
      btn.type = 'button';
      btn.textContent = `${m.index ?? i} · ${m.junction?.kind || '?'} · ${m.label || ''}`;
      btn.addEventListener('click', () => {
        selected = i;
        redraw();
      });
      return btn;
    }),
  );
}

async function main() {
  poc = await (await fetch(`${BASE}emulator/junction-poc/whitehall-farringdon.json`)).json();
  document.getElementById('routeTitle').textContent = poc.route || 'POC';
  renderList();
  selected = poc.maneuvers.findIndex((m) => IMPLEMENTED_KINDS.includes(m.junction?.kind));
  if (selected < 0) selected = 0;
  redraw();

  document.getElementById('prev').addEventListener('click', () => {
    selected = (selected - 1 + poc.maneuvers.length) % poc.maneuvers.length;
    redraw();
  });
  document.getElementById('next').addEventListener('click', () => {
    selected = (selected + 1) % poc.maneuvers.length;
    redraw();
  });
}

main().catch((err) => {
  console.error(err);
  statusEl.textContent = String(err);
});
