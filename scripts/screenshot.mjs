// screenshot.mjs drives a headless chromium over the DevTools protocol and
// captures the compliance dashboard in three views: light, dark, and a
// narrow phone-width viewport. Every view lands as a viewport PNG plus a
// full-page PNG and one row in dom-truth.json. Failed requests, console
// errors, and uncaught page exceptions are written to console-errors.json
// and fail the run. Beyond the pictures it measures what a string test
// cannot see: text contrast per theme (WCAG AA), the sticky column header,
// and a keyboard pass over the chips and the theme toggle.
//
// Usage: node scripts/screenshot.mjs <url> <outDir>
// Environment: CHROMIUM_BIN (chromium binary path), SETTLE_MS (default 2500).
import { spawn } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const [url, outDir] = process.argv.slice(2);
if (!url || !outDir) {
	console.error("usage: node scripts/screenshot.mjs <url> <outDir>");
	process.exit(2);
}

const settleMs = Number(process.env.SETTLE_MS || 2500);
const chromiumBin = process.env.CHROMIUM_BIN || "chromium";
const maxCaptureCssPx = 16000;
const profileDir = join(outDir, "chromium-profile");

const views = [
	{ name: "dashboard-light", width: 1440, height: 900, scale: 1, scheme: "light" },
	{ name: "dashboard-dark", width: 1440, height: 900, scale: 1, scheme: "dark" },
	{ name: "dashboard-narrow", width: 390, height: 844, scale: 2, scheme: "light" },
];

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Key events carry the character they type, or Chromium performs no default
// action: Enter activates a button only with "\r" on keyDown, Space only with
// " " on keyUp (a Space keyDown that carries text scrolls the page).
const keyText = { Enter: ["\r", ""], " ": ["", " "] };
const keyCodes = { Tab: 9, Enter: 13, " ": 32 };

mkdirSync(outDir, { recursive: true });
rmSync(profileDir, { recursive: true, force: true, maxRetries: 20, retryDelay: 100 });

const chrome = spawn(
	chromiumBin,
	[
		"--headless=new",
		"--no-sandbox",
		"--disable-gpu",
		"--disable-dev-shm-usage",
		"--hide-scrollbars",
		`--user-data-dir=${profileDir}`,
		"--remote-debugging-port=0",
		"about:blank",
	],
	{ stdio: ["ignore", "ignore", "pipe"] },
);
const chromeStderr = [];
chrome.stderr.on("data", (chunk) => chromeStderr.push(String(chunk)));

const pageErrors = [];
const browserNoise = [];
let ws = null;
let nextID = 1;
const pending = new Map();

function handleMessage(message) {
	const msg = JSON.parse(message.data);
	if (msg.id && pending.has(msg.id)) {
		const { resolve, reject } = pending.get(msg.id);
		pending.delete(msg.id);
		if (msg.error) reject(new Error(msg.error.message));
		else resolve(msg.result);
		return;
	}
	if (msg.method === "Network.responseReceived" && msg.params.response.status >= 400) {
		pageErrors.push(`${msg.params.response.status} ${msg.params.response.url}`);
		return;
	}
	if (
		msg.method === "Network.loadingFailed" &&
		!msg.params.canceled &&
		msg.params.errorText !== "net::ERR_ABORTED"
	) {
		pageErrors.push(`request failed: ${msg.params.errorText}`);
		return;
	}
	if (msg.method === "Runtime.exceptionThrown") {
		const details = msg.params.exceptionDetails;
		pageErrors.push(
			`exception ${details.url ?? ""}:${details.lineNumber ?? 0} ${details.exception?.description ?? details.text}`,
		);
		return;
	}
	if (msg.method === "Runtime.consoleAPICalled" && msg.params.type === "error") {
		pageErrors.push(
			`console.error ${msg.params.args.map((arg) => arg.value ?? arg.description ?? "").join(" ")}`,
		);
		return;
	}
	if (msg.method === "Log.entryAdded" && msg.params.entry.level === "error") {
		const text = `${msg.params.entry.source}: ${msg.params.entry.text}`;
		// Chromium's own extension-channel chatter, not the page's.
		if (/^other: Unchecked runtime\.lastError/.test(text)) browserNoise.push(text);
		else if (msg.params.entry.source !== "network") pageErrors.push(text);
	}
}

function send(method, params = {}, sessionId) {
	const id = nextID++;
	return new Promise((resolve, reject) => {
		pending.set(id, { resolve, reject });
		ws.send(JSON.stringify({ id, method, params, sessionId }));
	});
}

async function evaluate(expression, sessionId) {
	const result = await send("Runtime.evaluate", { expression, returnByValue: true }, sessionId);
	return result.result.value;
}

async function readDevToolsEndpoint() {
	const deadline = Date.now() + 30000;
	while (Date.now() < deadline) {
		try {
			const [port, path] = readFileSync(join(profileDir, "DevToolsActivePort"), "utf8").split("\n");
			if (port && path) return { port, path };
		} catch {
			// chromium has not written the endpoint file yet
		}
		await sleep(200);
	}
	throw new Error(
		`chromium published no DevTools endpoint within 30s: ${chromeStderr.join("").slice(-800)}`,
	);
}

async function waitForPage(sessionId) {
	const deadline = Date.now() + 30000;
	while (Date.now() < deadline) {
		if ((await evaluate("document.readyState", sessionId)) === "complete") return;
		await sleep(100);
	}
	throw new Error("page did not reach readyState complete within 30s");
}

async function capture(sessionId, name, clip) {
	const shot = await send(
		"Page.captureScreenshot",
		{ format: "png", captureBeyondViewport: true, clip: { ...clip, x: 0, y: 0, scale: 1 } },
		sessionId,
	);
	writeFileSync(join(outDir, `${name}.png`), Buffer.from(shot.data, "base64"));
}

async function shutdownChrome() {
	if (ws) ws.close();
	const exited = new Promise((resolve) => chrome.once("exit", resolve));
	chrome.kill("SIGKILL");
	await Promise.race([exited, sleep(5000)]);
	rmSync(profileDir, { recursive: true, force: true, maxRetries: 20, retryDelay: 100 });
}

const truthExpression = `JSON.stringify({
	dark: document.documentElement.classList.contains("dark"),
	title: document.title,
	rows: document.querySelectorAll("#fleet-rows tr[data-repo]").length,
	visibleRows: [...document.querySelectorAll("#fleet-rows tr[data-repo]")].filter((row) => getComputedStyle(row).display !== "none").length,
	visibleStates: [...document.querySelectorAll("#fleet-rows tr[data-repo]")].filter((row) => getComputedStyle(row).display !== "none").reduce((counts, row) => { const state = row.children[1]?.textContent?.trim() ?? ""; counts[state] = (counts[state] ?? 0) + 1; return counts; }, {}),
	statCards: [...document.querySelectorAll("#fleet-summary [role=group] button")].map((chip) => chip.innerText.replace(/\\s+/g, " ").trim()),
	pressedChips: [...document.querySelectorAll("#fleet-summary [role=group] button")].filter((chip) => chip.getAttribute("aria-pressed") === "true").map((chip) => chip.innerText.trim().split(/\\s+/)[0]),
	pressedChipHighlighted: [...document.querySelectorAll("#fleet-summary [role=group] button")].some((chip) => chip.getAttribute("aria-pressed") === "true" && chip.classList.contains("bg-accent-50")),
	headline: (document.querySelector("#fleet-summary p")?.innerText ?? "").replace(/\\s+/g, " ").trim(),
	barLabel: document.querySelector("#fleet-summary [role=img]")?.getAttribute("aria-label") ?? "",
	favicon: document.querySelector("link[rel~='icon']")?.getAttribute("href") ?? "",
	rowDisplay: getComputedStyle(document.querySelector("#fleet-rows tr[data-repo]")).display,
	tableDisplay: getComputedStyle(document.querySelector("#fleet-region table").closest("[data-show]")).display,
	emptyMessages: [...document.querySelectorAll("#fleet-region .text-center")].filter((el) => el.getClientRects().length > 0).map((el) => el.textContent.trim()),
	findingsCells: [...document.querySelectorAll("#fleet-rows tr[data-repo]")].filter((row) => (row.children[2]?.textContent ?? "").includes("finding")).length,
	viewportWidth: window.innerWidth,
	scrollWidth: document.documentElement.scrollWidth,
	pageHeight: document.documentElement.scrollHeight,
	theadDisplay: getComputedStyle(document.querySelector("#fleet-region thead")).display,
	headerSticky: getComputedStyle(document.querySelector("#fleet-region thead")).position === "sticky",
	tableScrolls: (() => { const box = document.querySelector("#fleet-region .fleet-table")?.parentElement; return box ? box.scrollHeight > box.clientHeight : false; })(),
	pageScrolls: document.documentElement.scrollHeight > window.innerHeight,
})`;

// In-page contrast measurement: every color is resolved through a 1x1
// canvas (Tailwind v4 emits oklch), translucent backgrounds and opacities
// are composited over the nearest opaque ancestor, and the WCAG ratio is
// computed from the relative luminance. Each named target reports the
// ratio of its first match; a missing element reports null.
const contrastExpression = `JSON.stringify((() => {
	const canvas = document.createElement("canvas");
	canvas.width = canvas.height = 1;
	const ctx = canvas.getContext("2d", { willReadFrequently: true });
	const rgba = (color) => {
		ctx.clearRect(0, 0, 1, 1);
		ctx.fillStyle = color;
		ctx.fillRect(0, 0, 1, 1);
		const [r, g, b, a] = ctx.getImageData(0, 0, 1, 1).data;
		return [r, g, b, a / 255];
	};
	const over = (top, bottom) => [0, 1, 2].map((i) => top[i] * top[3] + bottom[i] * (1 - top[3]));
	const background = (el) => {
		const layers = [];
		for (let node = el; node; node = node.parentElement) {
			const layer = rgba(getComputedStyle(node).backgroundColor);
			if (layer[3] === 0) continue;
			layers.push(layer);
			if (layer[3] === 1) break;
		}
		let base = [255, 255, 255];
		for (const layer of layers.reverse()) base = over([...layer.slice(0, 3), layer[3]], base);
		return base;
	};
	const opacity = (el) => {
		let product = 1;
		for (let node = el; node; node = node.parentElement) product *= Number(getComputedStyle(node).opacity);
		return product;
	};
	const luminance = ([r, g, b]) => {
		const f = (v) => {
			v /= 255;
			return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
		};
		return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b);
	};
	const measure = (el) => {
		if (!el) return null;
		const style = getComputedStyle(el);
		const bg = background(el);
		const raw = rgba(style.color);
		const alpha = raw[3] * opacity(el);
		const fg = over([raw[0], raw[1], raw[2], alpha], bg);
		const l1 = luminance(fg);
		const l2 = luminance(bg);
		const size = Number.parseFloat(style.fontSize);
		const weight = Number(style.fontWeight);
		const large = size >= 24 || (size >= 18.66 && weight >= 700);
		const ratio = (Math.max(l1, l2) + 0.05) / (Math.min(l1, l2) + 0.05);
		return { ratio: Math.round(ratio * 100) / 100, size, large, required: large ? 3 : 4.5, text: (el.textContent ?? "").trim().slice(0, 40) };
	};
	const badge = (label) => {
		for (const row of document.querySelectorAll("#fleet-rows tr[data-repo]")) {
			const cell = row.children[1];
			if ((cell?.textContent ?? "").trim() === label) return cell.querySelector("span");
		}
		return null;
	};
	const targets = [
		["headline word", document.querySelector("#fleet-summary p span")],
		["headline figure", document.querySelectorAll("#fleet-summary p span")[1] ?? null],
		["chip selected", document.querySelector('#fleet-summary [role=group] button[aria-pressed="true"]')],
		["chip selected count", document.querySelector('#fleet-summary [role=group] button[aria-pressed="true"] span')],
		["chip unselected", document.querySelector('#fleet-summary [role=group] button[aria-pressed="false"]')],
		["chip unselected count", document.querySelector('#fleet-summary [role=group] button[aria-pressed="false"] span')],
		["badge failing", badge("failing")],
		["badge stale", badge("stale")],
		["badge never-covered", badge("never-covered")],
		["badge compliant", badge("compliant")],
		["column header", document.querySelector("#fleet-region thead th")],
		["repo name", document.querySelector("#fleet-rows tr[data-repo] td:nth-child(1)")],
		["findings", document.querySelector("#fleet-rows tr[data-repo] td:nth-child(3) span")],
		["last run", document.querySelector("#fleet-rows tr[data-repo] td:nth-child(4) time")],
		["rendered status", document.querySelector("[data-tc-relative].text-xs") ?? document.querySelector("header [data-tc-relative]")],
	];
	return Object.fromEntries(targets.map(([name, el]) => [name, measure(el)]));
})())`;

// The keyboard pass: Tab reaches the chips and the theme toggle, Enter
// activates them, and the focused chip shows a focus ring.
const keyboardExpression = `JSON.stringify((() => {
	const active = () => {
		const el = document.activeElement;
		if (!el || el === document.body) return "body";
		return el.tagName.toLowerCase() + ":" + (el.getAttribute("aria-label") ?? el.textContent ?? "").trim().replace(/\\s+/g, " ").slice(0, 30);
	};
	return { active: active(), chipFocusVisible: document.activeElement?.matches?.("#fleet-summary [role=group] button:focus-visible") ?? false, outline: document.activeElement ? getComputedStyle(document.activeElement).outlineStyle + " " + getComputedStyle(document.activeElement).outlineWidth : "", dark: document.documentElement.classList.contains("dark") };
})())`;

const truth = [];
const filterChecks = [];
const contrast = [];
let keyboard = null;
let fatal = null;

const clickChip = async (label, sessionId) => {
	await evaluate(
		`(() => { const chip = [...document.querySelectorAll("#fleet-summary [role=group] button")].find((button) => button.innerText.trim().startsWith(${JSON.stringify(label)})); if (!chip) throw new Error("no chip labelled ${label}"); chip.click(); })()`,
		sessionId,
	);
	await sleep(500);

	return { clicked: label, ...JSON.parse(await evaluate(truthExpression, sessionId)) };
};

const press = async (key, sessionId) => {
	const [down, up] = keyText[key] ?? ["", ""];
	const base = { key, code: key === " " ? "Space" : key, windowsVirtualKeyCode: keyCodes[key], nativeVirtualKeyCode: keyCodes[key] };
	await send("Input.dispatchKeyEvent", { type: "keyDown", ...base, text: down }, sessionId);
	await send("Input.dispatchKeyEvent", { type: "keyUp", ...base, text: up }, sessionId);
	await sleep(150);
};

// tabUntil presses Tab until the active element's description starts with
// want, or stops after max presses; it returns every stop it passed.
const tabUntil = async (want, sessionId, max = 14) => {
	const stops = [];
	for (let i = 0; i < max; i++) {
		const state = JSON.parse(await evaluate(keyboardExpression, sessionId));
		if (state.active.startsWith(want)) return { reached: true, stops, state };
		await press("Tab", sessionId);
		stops.push(JSON.parse(await evaluate(keyboardExpression, sessionId)).active);
	}
	const state = JSON.parse(await evaluate(keyboardExpression, sessionId));

	return { reached: state.active.startsWith(want), stops, state };
};

// keyboardPass drives chips and theme toggle with the keyboard alone.
const keyboardPass = async (sessionId) => {
	await evaluate("document.activeElement?.blur?.()", sessionId);
	const chip = await tabUntil("button:Failing", sessionId);
	if (!chip.reached) return { reached: false, stops: chip.stops };
	await press("Enter", sessionId);
	const afterChip = JSON.parse(await evaluate(truthExpression, sessionId));
	const toggle = await tabUntil("button:Toggle dark mode", sessionId);
	if (!toggle.reached) return { reached: false, stops: toggle.stops, chipActivated: afterChip };
	const darkBefore = toggle.state.dark;
	await press("Enter", sessionId);
	const darkAfter = JSON.parse(await evaluate(keyboardExpression, sessionId)).dark;
	await press("Enter", sessionId);
	const darkRestored = JSON.parse(await evaluate(keyboardExpression, sessionId)).dark;

	return {
		reached: true,
		stopsToChip: chip.stops,
		chipFocusVisible: chip.state.chipFocusVisible,
		chipOutline: chip.state.outline,
		chipActivated: afterChip,
		stopsToToggle: toggle.stops,
		themeToggled: darkAfter !== darkBefore && darkRestored === darkBefore,
	};
};

try {
	const { port, path } = await readDevToolsEndpoint();
	ws = new WebSocket(`ws://127.0.0.1:${port}${path}`);
	ws.onmessage = handleMessage;
	await new Promise((resolve, reject) => {
		ws.onopen = resolve;
		ws.onerror = () => reject(new Error("DevTools websocket connect failed"));
	});

	for (const view of views) {
		const { targetId } = await send("Target.createTarget", { url: "about:blank" });
		const { sessionId } = await send("Target.attachToTarget", { targetId, flatten: true });

		await send("Page.enable", {}, sessionId);
		await send("Runtime.enable", {}, sessionId);
		await send("Log.enable", {}, sessionId);
		await send("Network.enable", {}, sessionId);
		await send(
			"Emulation.setDeviceMetricsOverride",
			{ width: view.width, height: view.height, deviceScaleFactor: view.scale, mobile: false },
			sessionId,
		);
		await send(
			"Emulation.setEmulatedMedia",
			{ features: [{ name: "prefers-color-scheme", value: view.scheme }] },
			sessionId,
		);

		await send("Page.navigate", { url }, sessionId);
		await waitForPage(sessionId);
		await sleep(settleMs);

		const facts = JSON.parse(await evaluate(truthExpression, sessionId));
		const fullHeight = Math.min(facts.pageHeight, maxCaptureCssPx / view.scale);

		await capture(sessionId, view.name, { width: view.width, height: view.height });
		await capture(sessionId, `${view.name}-full`, { width: view.width, height: fullHeight });
		truth.push({ view: view.name, scheme: view.scheme, ...facts });
		contrast.push({ view: view.name, scheme: view.scheme, ...JSON.parse(await evaluate(contrastExpression, sessionId)) });

		if (view.name === "dashboard-light") {
			keyboard = await keyboardPass(sessionId);

			if (facts.tableScrolls) {
				await evaluate(
					'(() => { const box = document.querySelector("#fleet-region .fleet-table"); box.scrollTop = Math.min(1400, box.scrollHeight - box.clientHeight); })()',
					sessionId,
				);
				await sleep(300);
				await capture(sessionId, "dashboard-light-scrolled", { width: view.width, height: view.height });
			}

			filterChecks.push(await clickChip("Stale", sessionId));
			await capture(sessionId, "dashboard-light-stale", { width: view.width, height: view.height });
			filterChecks.push(await clickChip("Failing", sessionId));
			await capture(sessionId, "dashboard-light-failing", {
				width: view.width,
				height: view.height,
			});
			filterChecks.push(await clickChip("All", sessionId));
		}

		await send("Target.closeTarget", { targetId });
	}
} catch (error) {
	fatal = error;
} finally {
	await shutdownChrome();
}

writeFileSync(
	join(outDir, "dom-truth.json"),
	`${JSON.stringify({ views: truth, filterChecks }, null, 2)}\n`,
);
writeFileSync(
	join(outDir, "contrast.json"),
	`${JSON.stringify({ contrast, keyboard }, null, 2)}\n`,
);
writeFileSync(
	join(outDir, "console-errors.json"),
	`${JSON.stringify({ pageErrors, browserNoise }, null, 2)}\n`,
);

for (const entry of truth) {
	console.log(
		`${entry.view}: dark=${entry.dark} rows=${entry.rows} visible=${entry.visibleRows} rowDisplay=${entry.rowDisplay} pressed=[${entry.pressedChips}] highlighted=${entry.pressedChipHighlighted} ` +
			`headline="${entry.headline}" chips=[${entry.statCards.join(" | ")}] bar="${entry.barLabel}" favicon=${entry.favicon} ` +
			`width=${entry.viewportWidth} scrollWidth=${entry.scrollWidth} pageHeight=${entry.pageHeight} th=${entry.thPosition}/${entry.theadDisplay} tableScrolls=${entry.tableScrolls}`,
	);
}

for (const entry of contrast) {
	console.log(
		`contrast ${entry.view}: ` +
			Object.entries(entry)
				.filter(([name]) => name !== "view" && name !== "scheme")
				.map(([name, m]) => `${name}=${m ? m.ratio : "n/a"}`)
				.join(" "),
	);
}

if (keyboard) {
	console.log(
		`keyboard: reached=${keyboard.reached} chipFocusVisible=${keyboard.chipFocusVisible} outline="${keyboard.chipOutline}" chipPressed=[${keyboard.chipActivated?.pressedChips ?? ""}] themeToggled=${keyboard.themeToggled}`,
	);
}

if (fatal) {
	console.error(`screenshot harness FAIL: ${fatal.message}`);
	process.exit(1);
}

const failures = [...pageErrors];

for (const entry of truth) {
	const wantsDark = entry.scheme === "dark";
	if (entry.dark !== wantsDark)
		failures.push(`${entry.view}: expected ${entry.scheme}, got ${entry.dark ? "dark" : "light"}`);
	if (entry.rows === 0 && process.env.ALLOW_EMPTY !== "1")
		failures.push(`${entry.view}: no fleet rows rendered`);
	if (entry.scrollWidth > entry.viewportWidth)
		failures.push(
			`${entry.view}: horizontal overflow (${entry.scrollWidth} > ${entry.viewportWidth})`,
		);
	if (entry.pressedChips.join() !== "All" || !entry.pressedChipHighlighted)
		failures.push(
			`${entry.view}: the All chip must read as selected on first paint (pressed=[${entry.pressedChips}] highlighted=${entry.pressedChipHighlighted})`,
		);
	const wide = entry.viewportWidth >= 640;
	if (wide && entry.thPosition !== "sticky")
		failures.push(`${entry.view}: the column header is ${entry.thPosition}, expected sticky`);
	if (wide && entry.theadDisplay === "none")
		failures.push(`${entry.view}: the column header is hidden in a wide layout`);
	if (!wide && (entry.thPosition === "sticky" || entry.theadDisplay !== "none"))
		failures.push(`${entry.view}: the stacked layout must drop the column header (th=${entry.thPosition}, thead=${entry.theadDisplay})`);
}

for (const entry of contrast) {
	for (const [name, m] of Object.entries(entry)) {
		if (name === "view" || name === "scheme" || !m) continue;
		if (m.ratio < m.required)
			failures.push(
				`contrast ${entry.view}: ${name} "${m.text}" is ${m.ratio}:1 at ${m.size}px, WCAG AA needs ${m.required}:1`,
			);
	}
}

if (!keyboard || !keyboard.reached) {
	failures.push(`keyboard: Tab never reached the chips and the theme toggle (stops: ${(keyboard?.stops ?? []).join(" > ")})`);
} else {
	if (!keyboard.chipFocusVisible || keyboard.chipOutline.startsWith("none") || keyboard.chipOutline.endsWith(" 0px"))
		failures.push(`keyboard: the focused chip shows no focus ring (visible=${keyboard.chipFocusVisible}, outline="${keyboard.chipOutline}")`);
	if (keyboard.chipActivated?.pressedChips?.join() !== "Failing")
		failures.push(`keyboard: Enter on the Failing chip did not select it (pressed=[${keyboard.chipActivated?.pressedChips ?? ""}])`);
	if (!keyboard.themeToggled) failures.push("keyboard: Enter on the theme toggle did not flip and restore the theme");
}

if (filterChecks.length === 3) {
	const [stale, failing, all] = filterChecks;
	const light = truth.find((entry) => entry.view === "dashboard-light");

	if (
		stale.visibleRows !== 0 ||
		stale.pressedChips.join() !== "Stale" ||
		!stale.pressedChipHighlighted ||
		stale.tableDisplay !== "none" ||
		stale.emptyMessages.join() !== "No stale repos"
	) {
		failures.push(
			`filter: after clicking Stale, ${stale.visibleRows} rows are visible, table display is ${stale.tableDisplay}, messages are [${stale.emptyMessages}], pressed=[${stale.pressedChips}]`,
		);
	}

	const failingWanted = light?.visibleStates.failing ?? 0;
	if (
		failing.visibleRows !== failingWanted ||
		Object.keys(failing.visibleStates).join() !== "failing" ||
		failing.pressedChips.join() !== "Failing" ||
		!failing.pressedChipHighlighted ||
		failing.tableDisplay === "none" ||
		failing.emptyMessages.length > 0
	) {
		failures.push(
			`filter: after clicking Failing, ${failing.visibleRows} rows are visible (${JSON.stringify(failing.visibleStates)}), expected ${failingWanted} failing; pressed=[${failing.pressedChips}]`,
		);
	}

	if (all.visibleRows !== all.rows || all.pressedChips.join() !== "All") {
		failures.push(
			`filter: after clicking All, ${all.visibleRows} of ${all.rows} rows are visible; pressed=[${all.pressedChips}]`,
		);
	}

	for (const check of filterChecks) {
		console.log(
			`filter ${check.clicked}: visible=${check.visibleRows} table=${check.tableDisplay} messages=[${check.emptyMessages}] states=${JSON.stringify(check.visibleStates)} pressed=[${check.pressedChips}] highlighted=${check.pressedChipHighlighted}`,
		);
	}
}

if (failures.length > 0) {
	console.error(
		`screenshot harness FAIL:\n${failures.map((failure) => `  - ${failure}`).join("\n")}`,
	);
	process.exit(1);
}

console.log(`screenshot harness OK: ${views.length} views in ${outDir}`);
