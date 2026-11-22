// Plays every looping pet clip, video[data-clip], only while it is on
// screen, and only shows its first frame to readers who ask for less motion.
const still = matchMedia('(prefers-reduced-motion: reduce)');
const observer = new IntersectionObserver(
	(entries) => {
		for (const { target, isIntersecting } of entries) {
			const video = target as HTMLVideoElement;
			if (!isIntersecting) video.pause();
			else if (!still.matches) video.play().catch(() => {});
			else if (video.preload === 'none') {
				video.preload = 'metadata';
				video.load();
			}
		}
	},
	{ rootMargin: '200px 0px' },
);
document.querySelectorAll<HTMLVideoElement>('video[data-clip]').forEach((v) => observer.observe(v));
