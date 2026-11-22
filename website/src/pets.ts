// The app's pets, read from their manifests in resources/pets, so the site
// shows what the app ships. scripts/clips.ts copies the clips named here
// into public/clips/, as the site cannot serve files outside website/.
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

// Both the build and the clips script run from website/.
const root = join(process.cwd(), '../resources/pets');

// The README's order: NSGo-chan first, then each source's pets together.
export const order = [
	'nsgo-chan',
	'bigbluefish',
	'playpets-clover',
	'playpets-daisy',
	'playpets-pip',
	'playpets-miso',
	'pip-mascot',
	'mochi',
	'shimeji',
	'openpets',
	'openpets-halloween',
] as const;

// The pet the home page's hero plays, with every clip of hers.
export const heroPet = 'nsgo-chan';

export type Pet = {
	id: string;
	name: string;
	width: number;
	height: number;
	clips: number;
	/** The idle clip's file name, without its extension. */
	idle: string;
	/** Where the pet takes the pointer, in percent of the frame. */
	body: { left: number; top: number; width: number; height: number };
	source: string;
	license: string;
	nonCommercial: boolean;
};

type Manifest = {
	name: string;
	width: number;
	height: number;
	clips: { idle: string };
	body: Pet['body'];
	source: string;
	license: string;
};
type Animation = {
	Id: string;
	File: string;
	Pool: string;
	Weight: number;
	NightOnly?: boolean;
	Meal?: string;
	MealBoost?: number;
	MoveKind?: string;
};

/** The rules a pet's clips play by, as its animations.json gives them. */
export type Rule = { id: string; pool: string; weight: number; nightOnly: boolean; mealBoost: number; moves: boolean };

export function rules(id: string): Rule[] {
	const animations: Animation[] = JSON.parse(readFileSync(join(root, id, 'animations.json'), 'utf8'));
	return animations.map((a) => ({
		id: a.Id,
		pool: a.Pool,
		weight: a.Weight,
		nightOnly: Boolean(a.NightOnly),
		mealBoost: a.Meal ? (a.MealBoost ?? 0) : 0,
		moves: Boolean(a.MoveKind),
	}));
}

export function clipFile(id: string, clip: string): string {
	const animations: Animation[] = JSON.parse(readFileSync(join(root, id, 'animations.json'), 'utf8'));
	const file = animations.find((a) => a.Id === clip)?.File;
	if (!file) throw new Error(`${id} has no clip ${clip}`);
	return file.replace(/\.webm$/, '');
}

export function loadPets(): Pet[] {
	return order.map((id) => {
		const m: Manifest = JSON.parse(readFileSync(join(root, id, 'pet.json'), 'utf8'));
		const animations: Animation[] = JSON.parse(readFileSync(join(root, id, 'animations.json'), 'utf8'));
		return {
			id,
			name: m.name,
			width: m.width,
			height: m.height,
			clips: animations.length,
			idle: clipFile(id, m.clips.idle),
			body: m.body,
			source: m.source,
			license: m.license.replace(/;\s*see NOTICE\.md$/, '').replace(/\s*\(non-commercial\)/, ''),
			nonCommercial: /non-commercial/i.test(m.license),
		};
	});
}

export const petsDir = root;
