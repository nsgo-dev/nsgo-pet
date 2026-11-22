// Draws the cards that previews show when a page is shared on X, Telegram
// and the like: one 1200×630 PNG per language in public/share/, from the
// home page's words and a frame of NSGo-chan's clip. Run it with
// `bun run share-cards` after changing either; it needs ffmpeg and Chrome.
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { copy } from '../src/copy';
import { petsDir } from '../src/pets';
import type { Lang } from '../src/site';

const chrome = process.env.CHROME ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const out = join(import.meta.dir, '../public/share');
const work = await mkdtemp(join(tmpdir(), 'share-cards-'));

async function run(cmd: string[]) {
	const proc = Bun.spawn(cmd, { stdout: 'ignore', stderr: 'pipe' });
	if ((await proc.exited) !== 0) throw new Error(`${cmd[0]}: ${await new Response(proc.stderr).text()}`);
}

// The middle of her hop, decoded with libvpx to keep the clip's alpha.
const frame = join(work, 'pet.png');
await run(['ffmpeg', '-loglevel', 'error', '-y', '-c:v', 'libvpx-vp9', '-ss', '0.8',
	'-i', join(petsDir, 'nsgo-chan/webm/click-hop.webm'), '-frames:v', '1', frame]);
const pet = `data:image/png;base64,${(await readFile(frame)).toString('base64')}`;
const logo = `data:image/png;base64,${(await readFile(join(import.meta.dir, '../src/assets/logo.png'))).toString('base64')}`;
const tokens = await readFile(join(import.meta.dir, '../src/styles/tokens.css'), 'utf8');

const html = (lang: Lang) => {
	const t = copy[lang];
	return `<!doctype html><html lang="${t.htmlLang}" data-theme="light"><meta charset="utf-8"><style>
${tokens}
* { box-sizing: border-box; margin: 0; }
html, body { width: 1200px; height: 630px; overflow: hidden; }
body { background: var(--bg); color: var(--text); font-family: var(--font-sans);
	display: grid; grid-template-columns: 1fr 470px; gap: 56px; padding: 64px 72px; -webkit-font-smoothing: antialiased; }
.text { display: flex; flex-direction: column; min-width: 0; }
.brand { display: flex; align-items: center; gap: 14px; font-size: 28px; font-weight: 650; letter-spacing: -0.01em; }
.brand img { width: 52px; height: 52px; border-radius: 14px; }
h1 { margin-top: auto; font-size: ${lang === 'zh' ? 76 : 70}px; line-height: 1.08; font-weight: 750; letter-spacing: -0.03em; }
h1 span { display: block; }
h1 span + span { color: var(--accent); }
.chips { display: flex; gap: 10px; margin-top: 36px; }
.chips span { border: 1.5px solid var(--border-strong); border-radius: 999px; padding: 8px 18px; font-size: 22px; color: var(--muted); }
.tile { position: relative; border-radius: 36px; border: 1.5px solid var(--border);
	background: radial-gradient(60% 55% at 50% 58%, rgb(0 122 255 / 0.16), transparent 70%), var(--surface); overflow: hidden; }
.tile img { position: absolute; width: 560px; height: 560px; left: 50%; top: 40px; transform: translateX(-50%); }
</style><body>
<div class="text">
	<div class="brand"><img src="${logo}" alt="">nsgo-pet</div>
	<h1>${t.headline.map((line) => `<span>${line}</span>`).join('')}</h1>
	<div class="chips">${t.meta.map((m) => `<span>${m}</span>`).join('')}</div>
</div>
<div class="tile"><img src="${pet}" alt=""></div>
</body></html>`;
};

await mkdir(out, { recursive: true });
for (const lang of ['en', 'zh'] as const) {
	const page = join(work, `${lang}.html`);
	await writeFile(page, html(lang));
	await run([chrome, '--headless=new', '--hide-scrollbars', '--force-device-scale-factor=1',
		'--window-size=1200,630', `--screenshot=${join(out, `${lang}.png`)}`, `file://${page}`]);
}
await rm(work, { recursive: true, force: true });
console.log('share cards: public/share/en.png, public/share/zh.png');
