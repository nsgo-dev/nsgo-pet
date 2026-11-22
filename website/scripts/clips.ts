// Copies the clips the site plays from ../resources/pets into public/clips/:
// each pet's idle clip, and every clip of the pet the home page's hero plays.
// Both forms go: Safari plays the HEVC .mov with alpha, the others the WebM.
import { copyFile, mkdir, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { clipFile, heroPet, loadPets, petsDir, rules } from '../src/pets';

const out = join(import.meta.dir, '../public/clips');
await rm(out, { recursive: true, force: true });

async function copy(id: string, file: string, name: string) {
	await mkdir(join(out, id), { recursive: true });
	for (const ext of ['webm', 'mov']) {
		await copyFile(join(petsDir, id, ext, `${file}.${ext}`), join(out, id, `${name}.${ext}`));
	}
}

for (const pet of loadPets()) await copy(pet.id, pet.idle, 'idle');
const hero = rules(heroPet);
for (const rule of hero) await copy(heroPet, clipFile(heroPet, rule.id), rule.id);
console.log(`clips: ${hero.length} of ${heroPet}'s and every pet's idle clip`);
