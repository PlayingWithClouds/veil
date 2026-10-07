<!--
	Full-screen celebration for the o-counter: pulses shoot up from below the
	screen and land on the "glass" as streaks stretched along their flight,
	strands and blobs. The liquid is translucent where thin and dense where
	thick (a soft body layer plus a denser core), with a faint dark rim and
	pinpoint glints. Heavy parts slowly run down, leaving a faint wet trail
	with beads. Plays once and calls onend; skipped under reduced motion.
-->
<script lang="ts">
	interface Props {
		// 0.3–1: how charged the button was; more pulses, reach and volume.
		intensity?: number;
		onend: () => void;
	}

	let { intensity = 1, onend }: Props = $props();

	// Pulses: start (ms) and strength (how far up and how much lands).
	const PULSES = [
		{ start: 0, strength: 1 },
		{ start: 460, strength: 0.88 },
		{ start: 860, strength: 0.7 },
		{ start: 1200, strength: 0.5 },
		{ start: 1480, strength: 0.34 }
	];
	const DROPS_PER_PULSE = 34;
	const SPRAY_PER_PULSE = 3;
	const SPRAY_PULSE = -1;
	// Each pulse's drops leave within this window, which strings them into a strand.
	const PULSE_SPREAD_MS = 170;
	// Landed splats spread to full size over this long.
	const SPREAD_MS = 320;
	const TOTAL_MS = 7000;
	const FADE_MS = 1600;
	// Landed drops closer than this (px) are joined into one strand.
	const STRAND_JOIN_PX = 70;

	// Cool, slightly grey white: thin parts show the page through it.
	const BODY_COLOR = 'rgb(214, 217, 214)';
	const CORE_COLOR = 'rgb(236, 238, 234)';
	const BODY_ALPHA = 0.7;
	const CORE_ALPHA = 0.8;
	// The core covers most of the body, leaving only a thin translucent edge.
	const CORE_SHRINK = 0.8;
	const RIM_ALPHA = 0.28;

	type Bead = { offsetY: number; radius: number };
	type Lump = { offsetX: number; offsetY: number; radius: number };

	type Drop = {
		pulse: number;
		progress: number;
		launchAt: number;
		flightMs: number;
		startX: number;
		startY: number;
		endX: number;
		endY: number;
		radius: number;
		// A glob on the rope: a thick mass that runs and catches light.
		heavy: boolean;
		lumps: Lump[];
		// Direction of flight at impact; splats stretch along it.
		angle: number;
		// When it hit the glass (ms into the animation); -1 while flying.
		landedAt: number;
		// Running down after landing: current length, speed, final length.
		sag: number;
		sagSpeed: number;
		sagMax: number;
		// Small drops left behind along the run.
		beads: Bead[];
		nextBeadAt: number;
	};

	const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
	let canvas = $state<HTMLCanvasElement | null>(null);
	let opacity = $state(1);

	/** A random number in [min, max). */
	function between(min: number, max: number): number {
		return min + Math.random() * (max - min);
	}

	/** The drops of every pulse the charge allows, aimed from below the bottom centre. */
	function makeDrops(width: number, height: number): Drop[] {
		const drops: Drop[] = [];
		const pulseCount = Math.max(2, Math.round(PULSES.length * intensity));
		const volume = 0.75 + intensity * 0.4;
		PULSES.slice(0, pulseCount).forEach((pulse, pulseIndex) => {
			// Mostly upward shots; each sweeps a little sideways.
			const sweepFrom = between(-0.4, 0.4);
			const sweepTo = sweepFrom + between(-0.3, 0.3);
			const baseReach = height * pulse.strength * between(0.55, 0.95) * (0.75 + intensity * 0.25);
			// One thickness per shot; along it the width swells and thins smoothly
			// (no per-drop jitter, which zigzags the rope).
			const thickness = between(10, 15) * (0.7 + pulse.strength * 0.4) * volume;
			const wavePhase = between(0, Math.PI * 2);
			const globAt = globPositions();
			for (let index = 0; index < DROPS_PER_PULSE; index++) {
				const progress = index / DROPS_PER_PULSE;
				const angle = sweepFrom + (sweepTo - sweepFrom) * progress;
				const reach = baseReach * (1 - progress * 0.5) * (1 + 0.02 * Math.sin(progress * 6 + wavePhase));
				const swell = 1 + 0.3 * Math.sin(progress * 9 + wavePhase);
				const heavy = globAt.some((position) => Math.abs(position - progress) < 0.5 / DROPS_PER_PULSE);
				let radius = thickness * swell * (1 - progress * 0.4);
				if (heavy) radius *= between(1.5, 2);
				const drop = makeDrop(width, height, pulseIndex, progress, pulse.start, angle, reach, radius);
				drop.heavy = heavy;
				if (heavy) drop.lumps = makeLumps(radius);
				drops.push(drop);
			}
			for (let index = 0; index < SPRAY_PER_PULSE; index++) {
				const progress = Math.random();
				const angle = sweepFrom + (sweepTo - sweepFrom) * progress + between(-0.12, 0.12);
				const reach = baseReach * (1 - progress * 0.5) * between(0.8, 1.15);
				drops.push(makeDrop(width, height, SPRAY_PULSE, progress, pulse.start, angle, reach, between(1, 2.6)));
			}
		});
		return drops;
	}

	/** Where along a shot its big globs sit (0 = head); the head always gets one. */
	function globPositions(): number[] {
		const positions = [0];
		const extra = Math.round(between(2, 4));
		for (let index = 0; index < extra; index++) positions.push(between(0.15, 0.95));
		return positions;
	}

	/** Overlapping blobs around a glob's centre, so it's an irregular mass. */
	function makeLumps(radius: number): Lump[] {
		const lumps: Lump[] = [];
		const count = Math.round(between(3, 5));
		for (let index = 0; index < count; index++) {
			const angle = between(0, Math.PI * 2);
			const distance = radius * between(0.3, 0.8);
			lumps.push({
				offsetX: Math.cos(angle) * distance,
				offsetY: Math.sin(angle) * distance,
				radius: radius * between(0.5, 0.8)
			});
		}
		return lumps;
	}

	/** One drop leaving the source at its point in the pulse. */
	function makeDrop(
		width: number,
		height: number,
		pulse: number,
		progress: number,
		pulseStart: number,
		angle: number,
		reach: number,
		radius: number
	): Drop {
		const startX = width / 2 + between(-6, 6);
		const startY = height + 30;
		return {
			pulse,
			progress,
			launchAt: pulseStart + progress * PULSE_SPREAD_MS,
			flightMs: between(240, 360),
			startX,
			startY,
			endX: startX + Math.sin(angle) * reach,
			endY: startY - Math.cos(angle) * reach,
			radius,
			heavy: false,
			lumps: [],
			angle: angle - Math.PI / 2,
			landedAt: -1,
			sag: 0,
			sagSpeed: 0,
			sagMax: 0,
			beads: [],
			nextBeadAt: 0
		};
	}

	/** Where a flying drop is: an arc that slows as it nears the glass. */
	function flyingPosition(drop: Drop, elapsed: number): { x: number; y: number; t: number } {
		const t = Math.min(1, (elapsed - drop.launchAt) / drop.flightMs);
		const eased = 1 - (1 - t) * (1 - t);
		const lift = Math.sin(t * Math.PI) * 40;
		return {
			x: drop.startX + (drop.endX - drop.startX) * eased,
			y: drop.startY + (drop.endY - drop.startY) * eased - lift,
			t
		};
	}

	/** Lands a drop; heavier ones will run down the glass. */
	function land(drop: Drop, elapsed: number) {
		drop.landedAt = elapsed;
		if (!drop.heavy || Math.random() < 0.3) return;
		drop.sagMax = between(40, 260) * (drop.radius / 24);
		// Thick liquid runs slowly.
		drop.sagSpeed = between(10, 28);
		drop.nextBeadAt = between(40, 90);
	}

	/** How far a landed splat has spread, 0.75 on impact up to 1. */
	function spreadScale(drop: Drop, elapsed: number): number {
		const t = Math.min(1, (elapsed - drop.landedAt) / SPREAD_MS);
		return 0.75 + 0.25 * (1 - (1 - t) * (1 - t));
	}

	/** Lengthens runs, slowing as they thin out, and leaves beads behind. */
	function sagDrops(drops: Drop[], seconds: number) {
		for (const drop of drops) {
			if (drop.landedAt < 0 || drop.sag >= drop.sagMax) continue;
			const remaining = 1 - drop.sag / drop.sagMax;
			drop.sag = Math.min(drop.sagMax, drop.sag + drop.sagSpeed * seconds * (0.25 + remaining));
			if (drop.sag < drop.nextBeadAt) continue;
			drop.beads.push({ offsetY: drop.sag, radius: drop.radius * between(0.18, 0.3) });
			drop.nextBeadAt = drop.sag + between(50, 120);
		}
	}

	/**
	 * Paints the liquid's shapes in one colour. The core pass shrinks every
	 * shape, so only the thick parts come out dense.
	 */
	function paintShapes(context: CanvasRenderingContext2D, drops: Drop[], elapsed: number, shrink: number) {
		context.clearRect(0, 0, context.canvas.width, context.canvas.height);
		context.lineCap = 'round';
		context.lineJoin = 'round';
		paintStrands(context, drops, shrink);
		for (const drop of drops) {
			if (elapsed < drop.launchAt) continue;
			if (drop.landedAt >= 0) {
				paintSplat(context, drop, elapsed, shrink);
				continue;
			}
			paintFlying(context, drop, elapsed, shrink);
		}
	}

	/** Strands between consecutive landed drops of the same pulse. */
	function paintStrands(context: CanvasRenderingContext2D, drops: Drop[], shrink: number) {
		for (let index = 1; index < drops.length; index++) {
			const previous = drops[index - 1];
			const current = drops[index];
			if (current.pulse === SPRAY_PULSE || previous.pulse !== current.pulse) continue;
			if (previous.landedAt < 0 || current.landedAt < 0) continue;
			if (Math.hypot(current.endX - previous.endX, current.endY - previous.endY) > STRAND_JOIN_PX) continue;
			context.lineWidth = Math.min(previous.radius, current.radius) * 1.5 * shrink;
			context.beginPath();
			context.moveTo(previous.endX, previous.endY);
			context.lineTo(current.endX, current.endY);
			context.stroke();
		}
	}

	/**
	 * A landed splat stretched along its flight, with its run. Strand drops
	 * stay within the strand's width so the strand reads as one smooth rope;
	 * only clumps bulge out of it.
	 */
	function paintSplat(context: CanvasRenderingContext2D, drop: Drop, elapsed: number, shrink: number) {
		const radius = drop.radius * spreadScale(drop, elapsed) * shrink;
		let length = radius * 0.85;
		let breadth = radius * 0.75;
		if (drop.heavy || drop.pulse === SPRAY_PULSE) {
			length = radius * 1.6;
			breadth = radius * 0.95;
		}
		context.beginPath();
		context.ellipse(drop.endX, drop.endY, length, breadth, drop.angle, 0, Math.PI * 2);
		for (const lump of drop.lumps) {
			const lumpX = drop.endX + lump.offsetX * shrink;
			const lumpY = drop.endY + lump.offsetY * shrink;
			const lumpRadius = lump.radius * shrink;
			context.moveTo(lumpX + lumpRadius, lumpY);
			context.arc(lumpX, lumpY, lumpRadius, 0, Math.PI * 2);
		}
		context.fill();
		if (drop.sag <= 0) return;
		paintRun(context, drop, radius);
	}

	/**
	 * The run below a splat: a neck narrowing into a faint thin trail, with the
	 * gathered liquid as one heavy drop at the tip and a rare bead left behind.
	 */
	function paintRun(context: CanvasRenderingContext2D, drop: Drop, radius: number) {
		const x = drop.endX;
		const y = drop.endY;
		const tipY = y + drop.sag;
		const neckLength = Math.min(drop.sag, radius * 3);
		context.beginPath();
		context.moveTo(x - radius * 0.6, y);
		context.quadraticCurveTo(x - radius * 0.2, y + neckLength * 0.7, x - radius * 0.15, y + neckLength);
		context.lineTo(x + radius * 0.15, y + neckLength);
		context.quadraticCurveTo(x + radius * 0.2, y + neckLength * 0.7, x + radius * 0.6, y);
		context.closePath();
		context.fill();
		context.lineWidth = Math.max(0.8, radius * 0.28);
		context.beginPath();
		context.moveTo(x, y + neckLength);
		context.lineTo(x, tipY);
		context.stroke();
		context.beginPath();
		context.ellipse(x, tipY, radius * 0.55, radius * 0.75, 0, 0, Math.PI * 2);
		context.fill();
		for (const bead of drop.beads) {
			context.beginPath();
			context.arc(x, y + bead.offsetY, bead.radius, 0, Math.PI * 2);
			context.fill();
		}
	}

	/** A drop still in the air: a short streak along its path. */
	function paintFlying(context: CanvasRenderingContext2D, drop: Drop, elapsed: number, shrink: number) {
		const now = flyingPosition(drop, elapsed);
		const before = flyingPosition(drop, elapsed - 30);
		context.lineWidth = drop.radius * (0.6 + now.t * 0.6) * shrink;
		context.beginPath();
		context.moveTo(before.x, before.y);
		context.lineTo(now.x, now.y);
		context.stroke();
	}

	/** A few soft glints, only on clumps and the drops at run tips. */
	function paintGlints(context: CanvasRenderingContext2D, drops: Drop[]) {
		context.fillStyle = 'rgba(255, 255, 255, 0.6)';
		for (const drop of drops) {
			if (drop.landedAt < 0 || !drop.heavy) continue;
			glint(context, drop.endX - drop.radius * 0.45, drop.endY - drop.radius * 0.3, drop.radius, drop.angle);
			if (drop.sag <= 0) continue;
			glint(context, drop.endX - drop.radius * 0.2, drop.endY + drop.sag - drop.radius * 0.3, drop.radius * 0.6, 0);
		}
	}

	/** One small soft highlight, stretched along the liquid's direction. */
	function glint(context: CanvasRenderingContext2D, x: number, y: number, size: number, angle: number) {
		context.beginPath();
		context.ellipse(x, y, Math.max(0.8, size * 0.3), Math.max(0.5, size * 0.12), angle, 0, Math.PI * 2);
		context.fill();
	}

	/**
	 * Composes a frame: a faint dark rim (the liquid's refracting edge), the
	 * translucent body, the denser core where it's thick, then the glints.
	 */
	function composeFrame(
		screen: CanvasRenderingContext2D,
		body: CanvasRenderingContext2D,
		core: CanvasRenderingContext2D,
		drops: Drop[],
		elapsed: number
	) {
		body.fillStyle = BODY_COLOR;
		body.strokeStyle = BODY_COLOR;
		paintShapes(body, drops, elapsed, 1);
		core.fillStyle = CORE_COLOR;
		core.strokeStyle = CORE_COLOR;
		paintShapes(core, drops, elapsed, CORE_SHRINK);

		const width = screen.canvas.width;
		const height = screen.canvas.height;
		screen.setTransform(1, 0, 0, 1, 0, 0);
		screen.clearRect(0, 0, width, height);
		screen.globalAlpha = RIM_ALPHA;
		screen.filter = 'brightness(0) blur(1.5px)';
		screen.drawImage(body.canvas, 1.5, 2);
		screen.globalAlpha = BODY_ALPHA;
		screen.filter = 'blur(0.8px)';
		screen.drawImage(body.canvas, 0, 0);
		screen.globalAlpha = CORE_ALPHA;
		screen.filter = 'blur(1.5px)';
		screen.drawImage(core.canvas, 0, 0);
		screen.globalAlpha = 1;
		screen.filter = 'none';
		const scale = width / window.innerWidth;
		screen.setTransform(scale, 0, 0, scale, 0, 0);
		paintGlints(screen, drops);
	}

	/** Buzzes along with the pulses. */
	function vibrateWithPulses() {
		const pattern: number[] = [];
		const pulseCount = Math.max(2, Math.round(PULSES.length * intensity));
		for (let index = 0; index < pulseCount; index++) {
			pattern.push(Math.round(90 * PULSES[index].strength), 300);
		}
		navigator.vibrate?.(pattern);
	}

	/** An offscreen layer the size of the screen canvas, drawn in CSS pixels. */
	function makeLayer(width: number, height: number, scale: number): CanvasRenderingContext2D | null {
		const layer = document.createElement('canvas');
		layer.width = width * scale;
		layer.height = height * scale;
		const context = layer.getContext('2d');
		if (!context) return null;
		context.scale(scale, scale);
		return context;
	}

	$effect(() => {
		if (reducedMotion) {
			onend();
			return;
		}
		if (!canvas) return;
		const scale = window.devicePixelRatio || 1;
		const width = window.innerWidth;
		const height = window.innerHeight;
		canvas.width = width * scale;
		canvas.height = height * scale;
		const screen = canvas.getContext('2d');
		const body = makeLayer(width, height, scale);
		const core = makeLayer(width, height, scale);
		if (!screen || !body || !core) return;
		const drops = makeDrops(width, height);
		vibrateWithPulses();

		let frame = 0;
		let startedAt = 0;
		let lastAt = 0;

		/** Advances and redraws until the overlay has faded out. */
		const step = (now: number) => {
			if (!startedAt) {
				startedAt = now;
				lastAt = now;
			}
			const elapsed = now - startedAt;
			for (const drop of drops) {
				if (drop.landedAt < 0 && elapsed >= drop.launchAt + drop.flightMs) land(drop, elapsed);
			}
			sagDrops(drops, (now - lastAt) / 1000);
			lastAt = now;
			composeFrame(screen, body, core, drops, elapsed);
			opacity = Math.min(1, Math.max(0, (TOTAL_MS - elapsed) / FADE_MS));
			if (elapsed >= TOTAL_MS) {
				onend();
				return;
			}
			frame = requestAnimationFrame(step);
		};
		frame = requestAnimationFrame(step);
		return () => cancelAnimationFrame(frame);
	});
</script>

{#if !reducedMotion}
	<div class="z-modal pointer-events-none fixed inset-0" style:opacity aria-hidden="true">
		<canvas bind:this={canvas} class="absolute inset-0 h-full w-full"></canvas>
	</div>
{/if}
