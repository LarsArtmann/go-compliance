// screenshot.mjs drives a headless chromium over the DevTools protocol and
// captures the compliance dashboard in three views: light, dark, and a
// narrow phone-width viewport. Every view lands as a viewport PNG plus a
// full-page PNG and one row in dom-truth.json. Failed requests, console
// errors, and uncaught page exceptions are written to console-errors.json
// and fail the run.
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
	if (msg.method === "Network.loadingFailed" && !msg.params.canceled && msg.params.errorText !== "net::ERR_ABORTED") {
		pageErrors.push(`request failed: ${msg.params.errorText}`);
		return;
	}
	if (msg.method === "Runtime.exceptionThrown") {
		const details = msg.params.exceptionDetails;
		pageErrors.push(`exception ${details.url ?? ""}:${details.lineNumber ?? 0} ${details.exception?.description ?? details.text}`);
		return;
	}
	if (msg.method === "Runtime.consoleAPICalled" && msg.params.type === "error") {
		pageErrors.push(`console.error ${msg.params.args.map((arg) => arg.value ?? arg.description ?? "").join(" ")}`);
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
	throw new Error(`chromium published no DevTools endpoint within 30s: ${chromeStderr.join("").slice(-800)}`);
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
	findingsCells: [...document.querySelectorAll("#fleet-rows tr[data-repo]")].filter((row) => (row.children[2]?.textContent ?? "").includes("finding")).length,
	viewportWidth: window.innerWidth,
	scrollWidth: document.documentElement.scrollWidth,
	pageHeight: document.documentElement.scrollHeight,
})`;

const truth = [];
const filterChecks = [];
let fatal = null;

const clickChip = async (label, sessionId) => {
	await evaluate(
		`(() => { const chip = [...document.querySelectorAll("#fleet-summary [role=group] button")].find((button) => button.innerText.trim().startsWith(${JSON.stringify(label)})); if (!chip) throw new Error("no chip labelled ${label}"); chip.click(); })()`,
		sessionId,
	);
	await sleep(500);

	return { clicked: label, ...JSON.parse(await evaluate(truthExpression, sessionId)) };
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

		if (view.name === "dashboard-light") {
			filterChecks.push(await clickChip("Stale", sessionId));
			filterChecks.push(await clickChip("Failing", sessionId));
			await capture(sessionId, "dashboard-light-failing", { width: view.width, height: view.height });
			filterChecks.push(await clickChip("All", sessionId));
		}

		await send("Target.closeTarget", { targetId });
	}
} catch (error) {
	fatal = error;
} finally {
	await shutdownChrome();
}

writeFileSync(join(outDir, "dom-truth.json"), `${JSON.stringify({ views: truth, filterChecks }, null, 2)}\n`);
writeFileSync(
	join(outDir, "console-errors.json"),
	`${JSON.stringify({ pageErrors, browserNoise }, null, 2)}\n`,
);

for (const entry of truth) {
	console.log(
		`${entry.view}: dark=${entry.dark} rows=${entry.rows} visible=${entry.visibleRows} rowDisplay=${entry.rowDisplay} pressed=[${entry.pressedChips}] highlighted=${entry.pressedChipHighlighted} ` +
			`headline="${entry.headline}" chips=[${entry.statCards.join(" | ")}] bar="${entry.barLabel}" favicon=${entry.favicon} ` +
			`width=${entry.viewportWidth} scrollWidth=${entry.scrollWidth} pageHeight=${entry.pageHeight}`,
	);
}

if (fatal) {
	console.error(`screenshot harness FAIL: ${fatal.message}`);
	process.exit(1);
}

const failures = [...pageErrors];

for (const entry of truth) {
	const wantsDark = entry.scheme === "dark";
	if (entry.dark !== wantsDark) failures.push(`${entry.view}: expected ${entry.scheme}, got ${entry.dark ? "dark" : "light"}`);
	if (entry.rows === 0 && process.env.ALLOW_EMPTY !== "1") failures.push(`${entry.view}: no fleet rows rendered`);
	if (entry.scrollWidth > entry.viewportWidth) failures.push(`${entry.view}: horizontal overflow (${entry.scrollWidth} > ${entry.viewportWidth})`);
	if (entry.pressedChips.join() !== "All" || !entry.pressedChipHighlighted) failures.push(`${entry.view}: the All chip must read as selected on first paint (pressed=[${entry.pressedChips}] highlighted=${entry.pressedChipHighlighted})`);
}

if (filterChecks.length === 3) {
	const [stale, failing, all] = filterChecks;
	const light = truth.find((entry) => entry.view === "dashboard-light");

	if (stale.visibleRows !== 0 || stale.pressedChips.join() !== "Stale" || !stale.pressedChipHighlighted) {
		failures.push(`filter: after clicking Stale, ${stale.visibleRows} rows are visible and pressed=[${stale.pressedChips}]`);
	}

	const failingWanted = light?.visibleStates.failing ?? 0;
	if (failing.visibleRows !== failingWanted || Object.keys(failing.visibleStates).join() !== "failing" || failing.pressedChips.join() !== "Failing" || !failing.pressedChipHighlighted) {
		failures.push(`filter: after clicking Failing, ${failing.visibleRows} rows are visible (${JSON.stringify(failing.visibleStates)}), expected ${failingWanted} failing; pressed=[${failing.pressedChips}]`);
	}

	if (all.visibleRows !== all.rows || all.pressedChips.join() !== "All") {
		failures.push(`filter: after clicking All, ${all.visibleRows} of ${all.rows} rows are visible; pressed=[${all.pressedChips}]`);
	}

	for (const check of filterChecks) {
		console.log(`filter ${check.clicked}: visible=${check.visibleRows} states=${JSON.stringify(check.visibleStates)} pressed=[${check.pressedChips}] highlighted=${check.pressedChipHighlighted}`);
	}
}

if (failures.length > 0) {
	console.error(`screenshot harness FAIL:\n${failures.map((failure) => `  - ${failure}`).join("\n")}`);
	process.exit(1);
}

console.log(`screenshot harness OK: ${views.length} views in ${outDir}`);
