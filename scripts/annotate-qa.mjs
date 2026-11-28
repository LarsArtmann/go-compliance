#!/usr/bin/env node
// Annotation QA and repair for swept status reports.
//
// Compares a swept report with a base revision and reports mechanical damage
// the docs gates do not see. With --fix it applies the repairs that are
// purely mechanical; everything else is listed for a human decision.
//
//   node scripts/annotate-qa.mjs --base <rev> [--fix] [--json] <file>...
//
// Findings:
//   R  renumbered: the item's base text sits under another number
//   I  indented: a struck item starts in a column the base does not have
//   S  strike: odd number of ~~ outside code spans in the item
//   X  split: the strike closes before the item's last line
//   U  unclosed: a struck item whose strike never closes
//   A  marker: a routed marker sits before the item's last line
//   W  whitespace: the base's spaces are gone from the item's first line
//   T  text: the item's first line diverges from its base text
//   L  lost: words of the base item are missing
//
// Repairs (--fix): X (strike extended over the continuation lines, the
// verdict moves to the item's end), W (base spaces restored), R (base
// number restored), I (column 0 plus a blank line above). Exit code 1 when
// any finding remains after the repair.

import { execFileSync } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";
import { relative, resolve } from "node:path";

const NUMBERED = /^(\s*)(~~)?(\d+)\.(?:\s+|$)(.*)$/;
const MIN_COMMON_PREFIX = 24;

function parseArgs(argv) {
	const options = { base: null, fix: false, json: false, files: [] };
	for (let index = 0; index < argv.length; index++) {
		const arg = argv[index];
		if (arg === "--base") options.base = argv[++index];
		else if (arg === "--fix") options.fix = true;
		else if (arg === "--json") options.json = true;
		else options.files.push(arg);
	}
	if (!options.base || options.files.length === 0) {
		console.error("usage: annotate-qa.mjs --base <rev> [--fix] [--json] <file>...");
		process.exit(2);
	}
	return options;
}

function splitHead(line) {
	const match = NUMBERED.exec(line);
	if (!match) return null;
	return { indent: match[1], struck: Boolean(match[2]), number: match[3], text: match[4] };
}

function stripCode(text) {
	return text.replace(/`[^`]*`/g, "");
}

function strikeCount(text) {
	return (stripCode(text).match(/~~/g) ?? []).length;
}

// Offset of the n-th code-free `~~` in the raw text.
function rawStrikeOffset(text, nth) {
	let seen = 0;
	for (let index = 0; index + 1 < text.length; index++) {
		if (text[index] === "`") {
			const end = text.indexOf("`", index + 1);
			if (end > index) {
				index = end;
				continue;
			}
		}
		if (text.startsWith("~~", index)) {
			if (seen === nth) return index;
			seen++;
			index++;
		}
	}
	return -1;
}

function normalize(text) {
	return text.replace(/[~*`\s]/g, "");
}

function splitItems(lines) {
	const items = [];
	let current = null;
	for (let index = 0; index < lines.length; index++) {
		const line = lines[index];
		const head = splitHead(line);
		if (head) {
			current = { start: index, end: index, ...head };
			items.push(current);
			continue;
		}
		if (line.trim() === "" || /^\s*(#{1,6}\s|<!--|```)/.test(line)) current = null;
		else if (current) current.end = index;
	}
	return items;
}

function buildBaseIndex(baseLines) {
	const items = splitItems(baseLines);
	const byHead = new Map();
	for (const item of items) {
		item.block = baseLines.slice(item.start, item.end + 1);
		item.key = normalize(item.block.join(""));
		item.headKey = normalize(item.text);
		if (!byHead.has(item.headKey)) byHead.set(item.headKey, item);
	}
	return { items, byHead };
}

function matchBase(item, block, baseIndex) {
	const headKey = normalize(
		item.struck ? item.text.slice(0, rawStrikeOffset(item.text, 0)) : item.text,
	);
	const bodyKey = normalize(block.join(""));
	let prefix = baseIndex.byHead.get(headKey);
	if (!prefix) {
		for (const [key, base] of baseIndex.byHead) {
			if (key.length >= 8 && headKey.startsWith(key)) {
				prefix = base;
				break;
			}
		}
	}
	if (prefix) return { base: prefix, kind: "prefix" };
	for (const base of baseIndex.items) {
		if (base.key.length >= 8 && bodyKey.startsWith(base.key)) return { base, kind: "prefix" };
	}
	let best = null;
	let bestLength = 0;
	for (const base of baseIndex.items) {
		let length = 0;
		const limit = Math.min(base.key.length, bodyKey.length);
		while (length < limit && base.key[length] === bodyKey[length]) length++;
		if (length > bestLength) {
			bestLength = length;
			best = base;
		}
	}
	if (best && bestLength >= MIN_COMMON_PREFIX) return { base: best, kind: "diverged", at: bestLength };
	return { base: null, kind: "new" };
}

function missingWords(base, block) {
	const have = new Set(block.join(" ").split(/\s+/).map((word) => normalize(word)));
	return base.block
		.join(" ")
		.split(/\s+/)
		.map((word) => normalize(word))
		.filter((word) => word !== "" && !have.has(word));
}

function insertSpaces(cur, ref) {
	let result = "";
	let refIndex = 0;
	let pending = "";
	for (const ch of cur) {
		if (/\s/.test(ch)) {
			pending += ch;
			continue;
		}
		const refSpace = refIndex > 0 && /\s/.test(ref[refIndex - 1]);
		result += (pending === "" && refSpace ? " " : pending) + ch;
		refIndex++;
		pending = "";
	}
	return result + pending;
}

function inspect(lines, baseLines) {
	const baseIndex = buildBaseIndex(baseLines);
	const findings = [];
	const items = splitItems(lines);
	const add = (code, item, detail) =>
		findings.push({ code, line: item.start + 1, item: item.number, detail });
	for (const item of items) {
		const block = lines.slice(item.start, item.end + 1);
		const { base, kind, at } = matchBase(item, block, baseIndex);
		item.base = base;
		if (kind === "diverged") {
			const marker = normalize(block.join("")).slice(Math.max(0, at - 20), at + 50);
			add("T", item, `diverges from base item ${base.number} at "${marker}"`);
		}
		else if (kind === "new" && baseIndex.items.length > 0) add("T", item, "no base item matches");
		const closeOffset = item.struck ? rawStrikeOffset(item.text, 1) : -1;
		const head = item.struck && closeOffset >= 0 ? item.text.slice(0, closeOffset) : item.text;
		if (base) {
			if (base.number !== item.number) add("R", item, `base number ${base.number}`);
			if (base.indent === "" && item.indent !== "") add("I", item, `indent ${item.indent.length}, base 0`);
			const lost = missingWords(base, block);
			if (lost.length > 0) add("L", item, `${lost.length} word(s): ${lost.slice(0, 8).join(" ")}`);
			if (head !== base.text && normalize(head) === normalize(base.headKey === undefined ? "" : base.text) && normalize(head) !== "") {
				const fixed = insertSpaces(head, base.text);
				if (fixed !== head) {
					add("W", item, `base: ${JSON.stringify(base.text.slice(0, 60))}`);
					item.whitespaceFix = fixed;
				}
			}
		}
		const strikes = strikeCount(block.join("\n"));
		if (item.struck && strikes % 2 !== 0) add("S", item, `${strikes} strike markers`);
		let closeLine = -1;
		if (item.struck) {
			let seen = 0;
			for (let offset = 0; offset < block.length; offset++) {
				seen += strikeCount(block[offset]);
				if (seen >= 2) {
					closeLine = item.start + offset;
					break;
				}
			}
		}
		if (item.struck && closeLine < 0) add("U", item, "strike never closes");
		if (item.struck && closeLine >= 0 && closeLine !== item.end) {
			const later = block.slice(closeLine - item.start + 1).join("\n");
			const nested = block.slice(1).some((line) => /^\s+/.test(line));
			const repairable = strikeCount(block[0]) === 2 && strikeCount(later) === 0 && strikes === 2 && !nested;
			add("X", item, `${item.end - closeLine} continuation line(s)${repairable ? "" : " (manual)"}`);
			item.splitFix = repairable;
		}
		if (!item.struck) {
			const markerOffset = block.findIndex((line) => /→\s*routed/.test(line));
			if (markerOffset >= 0 && markerOffset !== block.length - 1) add("A", item, "routed marker before the item's last line");
		}
	}
	return { findings, items };
}

function applyFixes(lines, report) {
	let changes = 0;
	const codesOf = (item) =>
		new Set(report.findings.filter((finding) => finding.line === item.start + 1).map((finding) => finding.code));
	const inserts = [];
	for (const item of [...report.items].sort((left, right) => right.start - left.start)) {
		const codes = codesOf(item);
		if (codes.size === 0) continue;
		let { indent, struck, number, text } = item;
		if (codes.has("W") && item.whitespaceFix) {
			const closeOffset = struck ? rawStrikeOffset(text, 1) : -1;
			text = item.whitespaceFix + (closeOffset >= 0 ? text.slice(closeOffset) : "");
			changes++;
		}
		if (codes.has("X") && item.splitFix) {
			const closeOffset = rawStrikeOffset(text, 1);
			const tail = text.slice(closeOffset + 2).trim();
			text = text.slice(0, closeOffset);
			lines[item.end] = `${lines[item.end].trimEnd()}~~ ${tail}`.trimEnd();
			changes++;
		}
		if (codes.has("R") && item.base) {
			number = item.base.number;
			changes++;
		}
		if (codes.has("I") && item.base && item.base.indent === "") {
			for (let index = item.start; index <= item.end; index++) lines[index] = lines[index].trimStart();
			if (item.start > 0 && lines[item.start - 1].trim() !== "") inserts.push(item.start);
			indent = "";
			changes++;
		}
		lines[item.start] = `${indent}${struck ? "~~" : ""}${number}. ${text}`.trimEnd();
	}
	for (const index of inserts.sort((left, right) => right - left)) lines.splice(index, 0, "");
	return changes;
}

function run(options, file) {
	const abs = resolve(file);
	const rel = relative(process.cwd(), abs);
	let baseText;
	try {
		baseText = execFileSync("git", ["show", `${options.base}:${rel}`], {
			encoding: "utf8",
			stdio: ["ignore", "pipe", "ignore"],
		});
	} catch {
		return { file: rel, error: `not in ${options.base}` };
	}
	const baseLines = baseText.split("\n");
	let text = readFileSync(abs, "utf8");
	const hadFinalNewline = text.endsWith("\n");
	let lines = text.replace(/\n$/, "").split("\n");
	let report = inspect(lines, baseLines);
	let changes = 0;
	if (options.fix && report.findings.length > 0) {
		changes = applyFixes(lines, report);
		if (changes > 0) {
			writeFileSync(abs, lines.join("\n") + (hadFinalNewline ? "\n" : ""));
			report = inspect(lines, baseLines);
		}
	}
	return { file: rel, changes, findings: report.findings };
}

function main() {
	const options = parseArgs(process.argv.slice(2));
	const results = options.files.map((file) => run(options, file));
	const remaining = results.reduce((sum, result) => sum + (result.findings?.length ?? 0), 0);
	if (options.json) {
		console.log(JSON.stringify({ results, remaining }, null, 2));
	} else {
		for (const result of results) {
			if (result.error) {
				console.log(`${result.file}: ${result.error}`);
				continue;
			}
			console.log(`${result.file}: ${result.changes} repaired, ${result.findings.length} remain`);
			for (const finding of result.findings) {
				console.log(`  ${finding.code} L${finding.line} #${finding.item}: ${finding.detail}`);
			}
		}
		console.log(`total: ${remaining} finding(s) remain`);
	}
	process.exit(remaining > 0 || results.some((result) => result.error) ? 1 : 0);
}

main();
